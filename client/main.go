package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	var (
		url      string
		duration time.Duration
		interval time.Duration
		fs       = pflag.NewFlagSet("labweek-client", pflag.ExitOnError)
	)

	fs.StringVar(&url, "url", "", "the URL to use")
	fs.DurationVarP(&duration, "duration", "d", 30*time.Second, "how long to run")
	fs.DurationVarP(&interval, "interval", "i", 50*time.Millisecond, "the interval between HTTP calls")
	fs.Parse(os.Args[1:])

	if len(url) == 0 {
		fmt.Fprintf(os.Stderr, "A url is required\n")
		os.Exit(1)
	}

	var (
		finished = new(sync.WaitGroup)
		done     = time.After(duration)
		ticker   = time.NewTicker(interval)
	)

	finished.Add(1)
	go func() {
		defer finished.Done()
		fmt.Printf("Invoking %s every %s for %s\n", url, interval, duration)

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				response, err := http.Get(url)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
				} else {
					io.Copy(ioutil.Discard, response.Body)
					response.Body.Close()
				}
			}
		}
	}()

	finished.Wait()
}
