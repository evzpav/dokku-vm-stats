package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
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

func main() {
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
		baseURL := flag.Arg(1)
		u, err := url.Parse(baseURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			fmt.Printf("invalid url: %v\n", err)
			break
		}

		everyMin := "*/1 * * * *"

		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("could not get working directory: %v\n", err)
			break
		}

		cronCmd := fmt.Sprintf("%s %s/%s --url=%s", everyMin, dir, "collectstats", baseURL)

		_, err = AddCronJob(cronCmd)
		if err != nil {
			fmt.Printf("failed to add cron command: %v\n", err)
			return
		}
	default:
		dokkuNotImplementExitCode, err := strconv.Atoi(os.Getenv("DOKKU_NOT_IMPLEMENTED_EXIT"))
		if err != nil {
			fmt.Println("failed to retrieve DOKKU_NOT_IMPLEMENTED_EXIT environment variable")
			dokkuNotImplementExitCode = 10
		}
		os.Exit(dokkuNotImplementExitCode)
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

	for {
		select {
		case <-done:
			return
		case t := <-s.ticker.C:
			fmt.Println("Tick at", t)
			m := getMemoryStats()
			s.sendStats(m)
		}
	}

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

func createEndpointFile(fileName, url string) error {
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println("failed to create stats url file")
		return err
	}

	_, err = f.Write([]byte(url))
	if err != nil {
		fmt.Println("failed to write stats url to file")
		return err
	}
	return nil
}

func readEndpointFile() (string, error) {
	f, err := os.Open("./stats_url.txt")
	if err != nil {
		fmt.Println("failed to open stats url file")
		return "", err
	}
	defer f.Close()

	b := bufio.NewReader(f)
	line, _, err := b.ReadLine()
	if err != nil {
		fmt.Println("failed to read stats url file")
		return "", err
	}

	// fmt.Printf("Url %v", string(line))

	return string(line), nil
}
