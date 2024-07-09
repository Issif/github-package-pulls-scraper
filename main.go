package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	colly "github.com/gocolly/colly/v2"
)

type stats struct {
	Date    string
	Package string
	Version string
	Count   string
}

const (
	URLBase = "https://github.com/orgs/<name>/packages?visibility=public"
)

var regVersion, regCount regexp.Regexp
var results []stats
var organization string

func init() {
	regVersion = *regexp.MustCompile(`(v)?[0-999]+\.[0-999]+\.[0-999]+(-.*)?`)
	regCount = *regexp.MustCompile(`[0-999]+(\,[0-999]+)?`)
}

func main() {
	orga := flag.String("p", "", "Your organization name")
	outputFolder := flag.String("o", "./outputs", "Destination folder for .csv")
	renderFolder := flag.String("r", "./renders", "Destination folder for the graphs")
	flag.Parse()

	if *orga == "" {
		fmt.Println("Usage:")
		flag.Usage()
		os.Exit(1)
	}

	organization = *orga

	*outputFolder += "/" + organization
	*renderFolder += "/" + organization

	if err := os.Mkdir("outputs", os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
		log.Printf("Folder '%v' already exists\n", *outputFolder)
	}

	scrape()
	writeCSV(outputFolder)
	renderChart(*outputFolder, *renderFolder)
}

func writeCSV(folder *string) {
	log.Printf("Writing of the .csv in '%v'\n", *folder)
	for _, i := range results {
		filename := fmt.Sprintf("%v/%v.csv", *folder, strings.ReplaceAll(i.Package, "/", "_"))
		_, errExist := os.Stat(filename)

		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if os.IsNotExist(errExist) {
			if _, err := f.WriteString("Date,Package,Version,Count\n"); err != nil {
				log.Fatal(err)
			}
		}
		defer f.Close()

		if _, err := f.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", i.Date, i.Package, i.Version, i.Count)); err != nil {
			return
		}
	}
}

func scrape() {
	var nbPackages int
	url := strings.ReplaceAll(URLBase, "<name>", organization)

	log.Printf("Start scrapping of '%v'\n", url)
	c := colly.NewCollector(colly.Async(true))
	c.Limit(&colly.LimitRule{Parallelism: 15})

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			log.Fatal("Status:", r.StatusCode)
		}
	})

	c.OnHTML(".paginate-container,.pagination", func(h *colly.HTMLElement) {
		h.ForEach("em", func(_ int, el *colly.HTMLElement) {
			if el.Text != el.Attr("data-total-pages") {
				page, _ := strconv.Atoi(el.Text)
				page++
				c.Visit(fmt.Sprintf("%v&page=%v", h.Request.URL.String(), page))
			}
		})
	})

	c.OnHTML(".Box-row", func(h *colly.HTMLElement) {
		h.ForEach("a", func(_ int, el *colly.HTMLElement) {
			if strings.Contains(el.Attr("class"), "Link--primary") {
				a := el.Attr("title")
				b := strings.Split(a, "/")
				u := fmt.Sprintf("https://github.com/%v/%v/pkgs/container/%v/versions?filters%%5Bversion_type%%5D=tagged", organization, b[0], a)
				log.Printf("Scrape pulls count for package '%v'\n", a)
				nbPackages++
				c.Visit(u)
			}
		})

		r := stats{}
		a := strings.Split(h.Request.URL.String(), "/")
		b := a[len(a)-3:]
		r.Date = time.Now().Format("2006-01-02")
		r.Package = fmt.Sprintf("%v/%v", b[0], b[1])
		h.ForEach("a", func(_ int, el *colly.HTMLElement) {
			if regVersion.Match([]byte(el.Text)) {
				r.Version = fmt.Sprintf("%v", el.Text)
			}
		})

		h.ForEach("span", func(_ int, el *colly.HTMLElement) {
			if el.Attr("class") == "d-flex flex-items-center gap-1 color-fg-muted overflow-hidden f6 mr-3" {
				c := string(regCount.FindAll([]byte(el.Text), -1)[0])
				c = strings.TrimSpace(c)
				c = strings.Trim(c, "\n")
				c = strings.ReplaceAll(c, ",", "")
				r.Count = fmt.Sprintf("%v", c)
			}
		})

		results = append(results, r)
	})

	c.Visit(url)
	c.Wait()
	results = removeIncomplete(results)
	log.Printf("%v package(s) found\n", nbPackages)
}

func removeIncomplete(r []stats) []stats {
	var s []stats
	for _, i := range results {
		if i.Package != "" && i.Version != "" && i.Count != "" {
			s = append(s, i)
		}
	}
	return s
}

func renderChart(dataFolder, renderFolder string) {
	log.Printf("Writing of the .html in '%v/'\n", renderFolder)

	files, err := os.ReadDir(dataFolder)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		p := strings.TrimSuffix(f.Name(), ".csv")

		file, err := os.Open(dataFolder + "/" + f.Name())
		if err != nil {
			log.Fatal(err)
		}

		dates := make(map[string]bool)
		versions := make(map[string]bool)
		count := make(map[string]map[string]string)

		fileScanner := bufio.NewScanner(file)
		fileScanner.Scan() // skip first list with headers
		for fileScanner.Scan() {
			s := fileScanner.Text()
			if count[strings.Split(s, ",")[0]] == nil {
				d := make(map[string]string)
				count[strings.Split(s, ",")[0]] = d
			}
			count[strings.Split(s, ",")[0]][strings.Split(s, ",")[2]] = strings.Split(s, ",")[3]
			dates[strings.Split(s, ",")[0]] = true
			versions[strings.Split(s, ",")[2]] = true
		}
		if err := fileScanner.Err(); err != nil {
			log.Fatal(err)
		}

		xData := make([]string, 0)
		for i := range dates {
			xData = append(xData, i)
		}
		sort.Strings(xData)

		line := charts.NewLine()
		line.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{
				PageTitle: strings.ReplaceAll(p, "_", "/"),
				Width:     "100%",
				Height:    "95vh"}),
			charts.WithTitleOpts(opts.Title{
				Title: strings.ReplaceAll(p, "_", "/"),
			}),
			charts.WithLegendOpts(opts.Legend{
				SelectedMode: "multiple",
			}),
			charts.WithTooltipOpts(opts.Tooltip{
				Trigger: "axis",
				AxisPointer: &opts.AxisPointer{
					Type: "cross",
				},
			}),
			charts.WithYAxisOpts(opts.YAxis{
				Name: "# pulls",
				Type: "value",
			}),
		)

		line.SetXAxis(xData)

		for i := range versions {
			yData := make([]opts.LineData, 0)
			for _, j := range xData {
				yData = append(yData, opts.LineData{Value: count[j][i]})
			}
			line.AddSeries(i, yData)
		}

		line.SetSeriesOptions(charts.WithMarkLineNameXAxisItemOpts(
			opts.MarkLineNameXAxisItem{
				XAxis: "2024-07-09",
			},
		))

		o, err := os.Create(fmt.Sprintf("%v/%v.html", renderFolder, p))
		if err != nil {
			log.Fatal(err)
		}
		line.Render(o)
	}
}
