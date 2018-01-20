package common

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func CheckLinksInPath(path string, domain string, linkJobs chan Link) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			links, err := extractLinks(path, domain)
			if err != nil {
				fmt.Errorf("Error reading file %s", path)
			}
			for _, link := range links {
				linkJobs <- link
			}
		}
		return nil
	})
}

func extractLinks(path string, domain string) ([]Link, error) {
	var links []Link

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	re := regexp.MustCompile("http(s?)://[^'\"<>\\]]+|(src|href)=['\"]/[^'\"<>\\]]+['\"]")
	innerRe := regexp.MustCompile("['\"][^'\"]+['\"]")

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		for _, match := range re.FindAllString(word, -1) {
			link := innerRe.FindString(match)
			if link == "" {
				link = match
			} else {
				link = link[1 : len(link)-1]
			}
			if link[:1] == "/" {
				link = domain + link
			}
			links = append(links, Link{path, link, 0})
		}
	}

	return links, nil
}
