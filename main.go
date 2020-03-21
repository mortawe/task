package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type urlInfo struct {
	url   string
	err   error
	count int
}

func printResult(results <-chan urlInfo, done chan<- bool) {
	total := 0
	for data := range results {
		if data.err != nil {
			log.Printf("error for %s : %s \n", data.url, data.err)
			continue
		}
		total += data.count
		fmt.Printf("count for %s : %d \n", data.url, data.count)
	}
	fmt.Printf("total : %d\n", total)
	done <- true
}

func scanInput(input chan<- string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := scanner.Text()
		input <- url
	}
	close(input)
}

func getURL(url, word string) urlInfo {
	data := urlInfo{
		url:   url,
		err:   nil,
		count: 0,
	}
	resp, err := http.Get(url)
	if err != nil {
		data.err = errors.New("can't get url")
		return data
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	data.count = strings.Count(string(bytes), word)

	return data
}

func count(input <-chan string, result chan<- urlInfo, goroutines chan bool, searchWord string) {
	wg := sync.WaitGroup{}

	for url := range input {
		wg.Add(1)
		goroutines <- true
		go func(str string) {
			defer wg.Done()
			data := getURL(str, searchWord)
			result <- data
			<-goroutines
		}(url)
	}

	wg.Wait()
	close(result)
}

func main() {
	var maxNConcurrentGoroutines int
	var searchWord string
	flag.IntVar(&maxNConcurrentGoroutines, "k", 5, "maximum goroutine count")
	flag.StringVar(&searchWord, "q", "Go", "search word")
	flag.Parse()

	done := make(chan bool)
	result := make(chan urlInfo)
	concurrentGoroutines := make(chan bool, maxNConcurrentGoroutines)
	input := make(chan string)

	go scanInput(input)
	go count(input, result, concurrentGoroutines, searchWord)
	go printResult(result, done)
	<-done
}
