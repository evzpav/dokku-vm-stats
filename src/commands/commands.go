package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mackerelio/go-osstat/memory"
	columnize "github.com/ryanuber/columnize"
)

const (
	helpHeader = `Usage: dokku stats[:COMMAND]

Runs commands that interact with the app's repo

Additional commands:`

	helpContent = `
    stats:start CLIENT_URL, start collecting VM data
`
)

type stats struct {
	url             string
	intervalSeconds int
	ticker          *time.Ticker
}

func newStats() *stats {
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	return &stats{
		ticker: ticker,
		// intervalSeconds: 1,
	}
}

// func (s *stats) SetInterval(i int) {
// 	s.intervalSeconds = i
// }

func (s *stats) SetClientURL(url string) {
	s.url = url
}

func (s *stats) sendStats(body interface{}) {
	fmt.Println(s.url)
	respBs, err := baseRequest(http.MethodPost, s.url, body)
	if err != nil {
		fmt.Printf("failed to post stats: %v\n", err)
		return
	}

	fmt.Printf("resp: %+v\n", respBs)
}

func main() {
	done := make(chan bool)
	stats := newStats()
	flag.Usage = usage
	flag.Parse()

	mainCommand := "stats"
	subcommand := ""

	cmd := flag.Arg(0)

	if cmd == mainCommand {
		usage()
		return
	}

	if strings.Contains(cmd, ":") {
		cmds := strings.Split(cmd, ":")
		if cmds[0] != mainCommand {
			fmt.Println("invalid command")
			return
		}
		if cmds[1] == "" {
			usage()
			return
		}

		subcommand = cmds[1]
	}

	switch subcommand {
	case "help":
		usage()
	case "start":
		url := flag.Arg(1)
		if url == "" {
			fmt.Println("invalid url")
			break
		}

		stats.SetClientURL(url)
		fmt.Println("client url set:", stats.url)

		stats.startCollectingData(done)
	default:
		fmt.Println("invalid stats subcommand: ", subcommand)
		os.Exit(1)
	}
}

func usage() {
	config := columnize.DefaultConfig()
	config.Delim = ","
	config.Prefix = "\t"
	config.Empty = ""
	content := strings.Split(helpContent, "\n")[1:]
	fmt.Println(helpHeader)
	fmt.Println(columnize.Format(content, config))
}

func (s *stats) startCollectingData(done chan bool) {
	fmt.Println("start collecting data")

	// statsData := make(chan interface{})

	// go func() {
	for {
		select {
		case <-done:
			return
		case t := <-s.ticker.C:
			fmt.Println("Tick at", t)
			m := getMemoryStats()
			s.sendStats(m)
			// statsData <- m
		}
	}
	// }()

	// for d := range statsData {
	// 	s.sendStats(d)
	// }

}

func baseRequest(method, url string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader

	if body != nil {
		bodyBs, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error to marshal body %v", err)
		}
		bodyReader = bytes.NewBuffer(bodyBs)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error in create request to [%s]: %v", url, err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	httpClient := http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error to complete request to [%s]: %v", url, err)
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error of [%s] reading body: %v", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request status code [%d] [%s]: %v", resp.StatusCode, url, string(bs))
	}

	return bs, nil
}

type Memory struct {
	Total  int64
	Used   int64
	Cached int64
	Free   int64
}

func getMemoryStats() Memory {
	m, err := memory.Get()
	if err != nil {
		fmt.Printf("%s\n", err)
		return Memory{}
	}

	memory := Memory{
		Total:  toMegaBytes(m.Total),
		Used:   toMegaBytes(m.Used),
		Cached: toMegaBytes(m.Cached),
		Free:   toMegaBytes(m.Free),
	}

	fmt.Printf("memory: %+v Mb\n", memory)

	return memory
}

func toMegaBytes(m uint64) int64 {
	return int64(m) / int64(math.Pow10(6))
}
