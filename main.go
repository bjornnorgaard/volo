package main

import (
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	WEBSITE = "http://dnd5e.wikidot.com/"
	DOMAIN  = "dnd5e.wikidot.com"

	forumURL  = "http://dnd5e.wikidot.com/forum*"
	systemURL = "http://dnd5e.wikidot.com/system*"

	OUTPUT = "output"
	RACES  = "races"
)

func main() {
	start := time.Now()

	disallowed := []*regexp.Regexp{
		regexp.MustCompile(forumURL),
		regexp.MustCompile(systemURL),
	}

	c := colly.NewCollector(
		colly.AllowedDomains(DOMAIN),
		colly.CacheDir("./cache"),
		colly.DisallowedURLFilters(disallowed...),
	)

	rule := &colly.LimitRule{
		DomainGlob:   fmt.Sprintf("*%s*", DOMAIN),
		DomainRegexp: fmt.Sprintf(".*%s.*", DOMAIN),
		Delay:        time.Millisecond * 200,
		RandomDelay:  time.Millisecond * 100,
		Parallelism:  1,
	}

	err := c.Limit(rule)
	if err != nil {
		log.Fatalln(err)
	}

	createDir(OUTPUT)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)
		_ = c.Visit(absoluteURL)
	})

	c.OnHTML("#skrollr-body > div.container-wrap-wrap > div.container-wrap > main > div > div > div > div", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()

		urlPath := strings.Split(url, ".com/")
		if len(urlPath) < 2 {
			return
		}

		split := strings.Split(urlPath[1], ":")
		if len(split) < 2 {
			return
		}

		category := split[0]
		article := split[1]

		createDir(filepath.Join(OUTPUT, category))

		html, _ := e.DOM.Html()
		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(html)
		if err != nil {
			log.Fatalln(err)
		}

		data := []byte(markdown)
		name := fmt.Sprintf("%s.md", article)
		filename := filepath.Join(OUTPUT, category, name)
		if err = os.WriteFile(filename, data, 0644); err != nil {
			log.Fatalln(err)
		}
	})

	err = c.Visit(WEBSITE)
	if err != nil {
		log.Fatalln(err)
	}

	slog.Info("Cache built", slog.Duration("elapsed", time.Since(start)))
}

func createDir(dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalln(err)
	}
}
