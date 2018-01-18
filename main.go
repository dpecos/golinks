package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	nWorkers     int
	mdPath       string
	domain       string
	onlyFailures bool
)

type Link struct {
	File   string
	Url    string
	Status int
}

func main() {

	flag.IntVar(&nWorkers, "workers", 10, "Number of workers")
	flag.StringVar(&mdPath, "path", ".", "Path with text files to check")
	flag.StringVar(&domain, "domain", "", "Domain to use for relative links")
	flag.BoolVar(&onlyFailures, "only-ko", false, "Show only failed URLs")
	flag.Parse()

	linkJobs := make(chan Link, nWorkers)
	results := make(chan Link)

	wg := startWorkers(linkJobs, results)
	wgRes := printResults(results)

	checkLinksInPath(mdPath, linkJobs)

	closeWorkers(linkJobs, results, wg, wgRes)
}

func startWorkers(linkJobs chan Link, results chan Link) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(nWorkers)

	for i := 0; i < nWorkers; i++ {
		go linkChecker(linkJobs, results, &wg)
	}

	return &wg
}

func printResults(results chan Link) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		totals := make(map[int]int)

		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		blue := color.New(color.FgBlue).SprintFunc()

		for link := range results {
			totals[link.Status]++

			var status string
			failed := link.Status < 200 || link.Status >= 300

			if failed {
				status = red(link.Status)
			} else {
				status = green(link.Status)
			}

			if failed || !onlyFailures {
				fmt.Printf("%s - %s - %s\n", status, link.Url, blue(link.File))
			}
		}

		fmt.Printf(blue("\n  Totals\n-----------\n"))
		for status, total := range totals {
			st := red(status)
			if status >= 200 && status < 300 {
				st = green(status)
			}
			fmt.Printf("%s\t%d\n", st, total)
		}
	}()

	return &wg
}

func closeWorkers(linkJobs chan Link, results chan Link, wg *sync.WaitGroup, wgRes *sync.WaitGroup) {
	close(linkJobs)
	wg.Wait()

	close(results)
	wgRes.Wait()
}

func checkLinksInPath(path string, linkJobs chan Link) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			links, err := extractLinks(path)
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

func extractLinks(path string) ([]Link, error) {
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

func linkChecker(linkJobs chan Link, results chan Link, wg *sync.WaitGroup) {
	defer wg.Done()

	var netClient = &http.Client{
		Timeout: time.Second * 5,
	}

	for link := range linkJobs {
		resp, err := netClient.Get(link.Url)
		if err != nil || resp.StatusCode < 200 && resp.StatusCode >= 300 {
			if resp != nil {
				link.Status = resp.StatusCode
			} else {
				link.Status = -1
			}
		} else {
			link.Status = resp.StatusCode
		}

		results <- link
	}
}
