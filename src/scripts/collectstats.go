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
	"net/url"
	"time"

	"github.com/mackerelio/go-osstat/memory"
)

func main() {
	baseURL := flag.String("url", "", "url to send data needed")
	flag.Parse()

	if *baseURL == "" {
		fmt.Println("url is needed. Use flag --url= ", *baseURL)
		return
	}

	u, err := url.Parse(*baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		fmt.Printf("invalid url: %s\n", *baseURL)
		return
	}

	m, err := getMemoryStats()
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := sendStats(u.String(), m)
	if err != nil {
		fmt.Printf("failed to post stats: %v\n", err)
		return
	}

	fmt.Printf("Success%+v\n", string(resp))

}

func sendStats(url string, body interface{}) ([]byte, error) {
	return baseRequest(http.MethodPost, url, body)

}

type Memory struct {
	Total  int64
	Used   int64
	Cached int64
	Free   int64
}

func getMemoryStats() (Memory, error) {
	m, err := memory.Get()
	if err != nil {
		fmt.Printf("%s\n", err)
		return Memory{}, err
	}

	memory := Memory{
		Total:  toMegaBytes(m.Total),
		Used:   toMegaBytes(m.Used),
		Cached: toMegaBytes(m.Cached),
		Free:   toMegaBytes(m.Free),
	}

	fmt.Printf("memory: %+v Mb\n", memory)

	return memory, nil
}

func toMegaBytes(m uint64) int64 {
	return int64(m) / int64(math.Pow10(6))
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
