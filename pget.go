package main

import (
	"os"
	"fmt"
	"math"
	"regexp"
	"strconv"
)

func FindPattern(url string) string {
	var match []string = regexp.MustCompile("^[^\\d]*([\\d]+)[^\\d]+.*$").FindStringSubmatch(url)
	if len(match) > 0 {
		return match[1]
	}
	return ""
}

func ParseIndexString(index string) (number int, maxPadding int, paddingFound bool, err error) {
	maxPadding = len(index)
	paddingFound = maxPadding > 1 && (index[0] == '0')
	var i64 int64 = 0
	i64, err = strconv.ParseInt(index, 10, 16)
	return int(i64), maxPadding, paddingFound, err
}

func ProbeUrlResource(s int) bool {
	return s > 0 && s < 20
}

func IntLen(number int) int {
	return int(math.Floor(math.Log10(float64(number)))) + 1
}

func ClosestShorterInt(number int) int {
	if number < 10 {
		return -1
	}
	multiplier := (IntLen(number) - 1)*10
	return (number/multiplier)*multiplier - 1
}

func ProbeExistence(url string) bool {
	return true
}

func TestPadding(urlPrefix string, urlSuffix string, testIndex int) bool {
	format := fmt.Sprintf("%%0%dd", IntLen(testIndex))
	paddedIndexString := fmt.Sprintf(format, ClosestShorterInt(testIndex))
	return ProbeExistence(fmt.Sprintf("%s%s%s", urlPrefix, paddedIndexString, urlSuffix))
}

func Crawl(initial int, channel chan bool) {
	channel <- ProbeUrlResource(initial)
}

func Crawler(scan int, next func (int) int, done chan int) {
	channel := make(chan bool)
	for {
		go Crawl(scan, channel)
		if !<-channel {
			break
		}
		scan = next(scan)
	}
	done <- scan
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	url := os.Args[1]
	match := FindPattern(url)
	if len(match) == 0 {
		os.Exit(1)
	}
	number, maxPadding, paddingFound, _ := ParseIndexString(match)
	fmt.Printf("Parse results: number %d, maxPadding %d, paddingFound %t\n", number, maxPadding, paddingFound)

	chanA, chanB := make(chan int), make(chan int)
	go Crawler(number, func(index int) int { return index - 1 }, chanA)
	go Crawler(number + 1, func(index int) int { return index + 1 }, chanB)
	fmt.Printf("Crawler 1 stopped at %d, crawler 2 stopped at %d\n",
		<-chanA, <-chanB)
	os.Exit(0)
}
