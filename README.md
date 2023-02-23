# Github packages pulls scraper

Github doesn't provide an API endpoint to track the number of pulls of the OCI packages it hosts. This script offers an hacky solution by scraping the packages page to extract the list then iterating to get the count of pulls per version.

It generates one comma separated `.csv` file per package, with the columns: `Date, Package, Version, Count`. Up to you to run the script every X times to add a new line in the files.

## Build

```shell
go build
```

### Usage

```shell
Usage of github-package-pulls-scraper:
  -o string
        Destination folder for .csv (default "./outputs")
  -p string
        Your profile or organization name
```

### Results

Log:
```
❯ go run . -p falcosecurity
2023/02/23 16:54:27 Start scrapping of 'https://github.com/orgs/falcosecurity/packages?visibility=public'
2023/02/23 16:54:27 Scrape pulls count for package 'rules/falco-rules'
2023/02/23 16:54:28 Scrape pulls count for package 'plugins/ruleset/k8saudit'
2023/02/23 16:54:30 Scrape pulls count for package 'plugins/plugin/k8saudit-eks'
2023/02/23 16:54:32 Scrape pulls count for package 'plugins/plugin/json'
2023/02/23 16:54:33 Scrape pulls count for package 'plugins/plugin/cloudtrail'
2023/02/23 16:54:34 Scrape pulls count for package 'plugins/plugin/k8saudit'
2023/02/23 16:54:35 Scrape pulls count for package 'plugins/ruleset/cloudtrail'
2023/02/23 16:54:37 Scrape pulls count for package 'plugins/plugin/dummy'
2023/02/23 16:54:38 Scrape pulls count for package 'plugins/plugin/okta'
2023/02/23 16:54:39 Scrape pulls count for package 'plugins/ruleset/okta'
2023/02/23 16:54:40 Scrape pulls count for package 'plugins/plugin/github'
2023/02/23 16:54:42 Scrape pulls count for package 'plugins/ruleset/github'
2023/02/23 16:54:43 Scrape pulls count for package 'rules/application-rules'
2023/02/23 16:54:44 Scrape pulls count for package 'event-generator/run-on-arch-falcosecurity-event-generator-ci-build-aarch64-alpine-latest'
2023/02/23 16:54:45 Scrape pulls count for package 'plugins/plugin/dummy_c'
2023/02/23 16:54:47 15 package(s) found
2023/02/23 16:54:47 Writing of the .csv in './outputs'
```

Example of a `.csv` file:
```csv
2023-02-22T16:52:54+01:00,ruleset/k8saudit,0.5.0,19390
2023-02-22T16:54:30+01:00,ruleset/k8saudit,0.5.0,19392
2023-02-22T17:27:12+01:00,ruleset/k8saudit,0.5.0,19417
```

## Author

Thomas Labarussias (https://github.com/Issif)
