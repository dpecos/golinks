package common

import (
	"fmt"
	"sync"

	"github.com/fatih/color"
)

type Link struct {
	File   string
	Url    string
	Status int
}

func PrintResults(results chan Link, onlyFailures bool) *sync.WaitGroup {
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
