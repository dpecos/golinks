package main

import (
	"flag"

	. "github.com/dpecos/golinks/common"
	"github.com/dpecos/golinks/workers"
)

var (
	nWorkers     int
	mdPath       string
	domain       string
	onlyFailures bool
)

func main() {

	flag.IntVar(&nWorkers, "workers", 10, "Number of workers")
	flag.StringVar(&mdPath, "path", ".", "Path with text files to check")
	flag.StringVar(&domain, "domain", "", "Domain to use for relative links")
	flag.BoolVar(&onlyFailures, "only-ko", false, "Show only failed URLs")
	flag.Parse()

	linkJobs := make(chan Link, nWorkers)
	results := make(chan Link)

	wg := workers.Start(nWorkers, linkJobs, results)
	wgRes := PrintResults(results, onlyFailures)

	CheckLinksInPath(mdPath, domain, linkJobs)

	workers.Stop(linkJobs, results, wg, wgRes)
}
