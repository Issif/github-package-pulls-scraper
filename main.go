package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

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

func init() {
	regVersion = *regexp.MustCompile(`[0-999]\.[0-999]\.[0-999](-.*)?`)
	regCount = *regexp.MustCompile(`[0-999]+(\,[0-999])?`)
}

func main() {
	profile := flag.String("p", "", "Your profile or organization name")
	output := flag.String("o", "./outputs", "Destination folder for .csv")
	flag.Parse()
	if *profile == "" {
		fmt.Println("Usage:")
		flag.Usage()
		os.Exit(1)
	}

	if err := os.Mkdir("outputs", os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
		log.Printf("Folder '%v' already exists\n", *output)
	}

	scrape(profile)
	writeCSV(output)
}

func writeCSV(folder *string) {
	log.Printf("Writing of the .csv in '%v'\n", *folder)
	for _, i := range results {
		filename := fmt.Sprintf("%v/%v_%v.csv", *folder, strings.ReplaceAll(i.Package, "/", "_"), i.Version)
		_, errorExist := os.Stat(filename)
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer f.Close()

		if os.IsNotExist(errorExist) {
			if _, err := f.WriteString("Date,Package,Version,Count\n"); err != nil {
				log.Fatal(err)
				return
			}
		}

		if _, err := f.WriteString(fmt.Sprintf("%v,%v,%v,%v\n", i.Date, i.Package, i.Version, i.Count)); err != nil {
			log.Fatal(err)
			return
		}
	}
}

func scrape(profile *string) {
	var nbPackages int
	url := strings.ReplaceAll(URLBase, "<name>", *profile)

	log.Printf("Start scrapping of '%v'\n", url)
	c := colly.NewCollector(colly.Async(true))
	c.Limit(&colly.LimitRule{Parallelism: 15})

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			log.Fatal("Status:", r.StatusCode)
		}
	})

	c.OnHTML(".Box-row", func(h *colly.HTMLElement) {
		h.ForEach("a", func(_ int, el *colly.HTMLElement) {
			if strings.Contains(el.Attr("class"), "Link--primary") {
				a := el.Attr("title")
				b := strings.Split(a, "/")
				u := fmt.Sprintf(`https://github.com/%v/%v/pkgs/container/%v/versions?filters%%5Bversion_type%%5D=tagged`, *profile, b[0], a)
				log.Printf("Scrape pulls count for package '%v'\n", a)
				nbPackages++
				c.Visit(u)
			}
		})
		r := stats{}
		a := strings.Split(h.Request.URL.String(), "/")
		b := a[len(a)-3:]
		r.Date = time.Now().Format(time.RFC3339)
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
