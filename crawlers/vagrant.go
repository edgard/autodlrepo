package crawlers

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Vagrant scrapes from downloads page
func Vagrant() (string, string, error) {
	res, err := http.Get("https://www.vagrantup.com/downloads.html")
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", "", err
	}

	var (
		release string
		url     string
	)
	doc.Find("div.downloads a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		if strings.HasSuffix(link, "_x86_64.rpm") {
			r, _ := regexp.Compile("vagrant_(.*)_x86_64.rpm")
			release = strings.TrimSpace(r.FindStringSubmatch(link)[1])
			url = strings.TrimSpace(link)
		}
	})
	if release != "" && url != "" {
		return release, url, nil
	}
	return "", "", errors.New("unable to find release")
}
