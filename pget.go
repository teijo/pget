package main

import (
	"os"
	"fmt"
	"math"
	"net/url"
	"net/http"
	"io/ioutil"
	"regexp"
	"strings"
	"strconv"
)

type Pattern struct {
	urlPrefix string
	match     string
	urlSuffix string
}

func fileName(u *url.URL) (string, string) {
	var urlSegments []string = strings.Split(u.Path, "/")
	last := len(urlSegments)-1
	return strings.Join(urlSegments[:last], "/"), urlSegments[last]
}

func extractIndex(str string) []string {
	return regexp.MustCompile("(^[^\\d]*)([\\d]+)([^\\d]+.*)$").FindStringSubmatch(str)
}

func FindPattern(urlString string) *Pattern {
	u, err := url.Parse(urlString)
	if (err != nil) {
		println(err)
	}
	filePath, file := fileName(u)
	var fileMatch []string = extractIndex(file)

	hasFileMatch := len(fileMatch) > 0 && len(fileMatch[2]) > 0
	if hasFileMatch {
		return &Pattern{fmt.Sprintf("http://%s%s/%s", u.Host, filePath, fileMatch[1]), fileMatch[2], fmt.Sprintf("%s?%s", fileMatch[3], u.RawQuery)}
	}

	var queryMatch []string = extractIndex(u.RawQuery)
	hasQueryMatch := len(queryMatch) > 0 && len(queryMatch[2]) > 0
	if hasQueryMatch {
		return &Pattern{fmt.Sprintf("http://%s%s%s", u.Host, u.Path, queryMatch[1]), queryMatch[2], queryMatch[3]}
	}

	var pathMatch []string = extractIndex(u.Path)
	hasPathMatch := len(pathMatch) > 0 && len(pathMatch[2]) > 0
	if hasPathMatch {
		return &Pattern{fmt.Sprintf("http://%s%s", u.Host, pathMatch[1]), pathMatch[2], fmt.Sprintf("%s?%s", pathMatch[3], u.RawQuery)}
	}

	return nil
}

func ParseIndexAndFormat(pattern *Pattern) (number int, format string, err error) {
	maxPadding := len(pattern.match)
	if maxPadding > 1 && (pattern.match[0] == '0') {
		format = fmt.Sprintf("%%0%dd", maxPadding)
	} else {
		format = "%d"
	}

	var i64 int64 = 0
	i64, err = strconv.ParseInt(pattern.match, 10, 16)
	return int(i64), format, err
}

func ProbeUrlResource(urlString string) bool {
	res, err := http.Get(urlString)

	if (err != nil) {
		fmt.Printf("GET FAILED: %s\n", err)
		return false
	}

	defer res.Body.Close()
	fmt.Printf("Probing %s ... %d\n", urlString, res.StatusCode)

	if res.StatusCode != 200 {
		return false
	}

	urlSegments := strings.Split(urlString, "/")
	filename := urlSegments[len(urlSegments)-1]
	body, err := ioutil.ReadAll(res.Body)
	if (err != nil) {
		fmt.Printf("Reading FAILED: %s\n", err)
		return false
	}

	if (len(body) == 0) {
		return true
	}

	err = ioutil.WriteFile(filename, body, os.ModePerm)
	if (err != nil) {
		fmt.Printf("Writing FAILED: %s\n", err)
		return false
	}

	return true
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

func BuildUrl(scan int, format string, pattern *Pattern) string {
	printFmt := fmt.Sprintf("%%s%s%%s", format)
	return fmt.Sprintf(printFmt, pattern.urlPrefix, scan, pattern.urlSuffix)
}

func Crawl(scan int, format string, pattern *Pattern, channel chan bool) {
	channel <- ProbeUrlResource(BuildUrl(scan, format, pattern))
}

func Crawler(scan int, format string, pattern *Pattern, next func (int) int, done chan int) {
	channel := make(chan bool)
	for {
		go Crawl(scan, format, pattern, channel)
		if !<-channel {
			break
		}
		scan = next(scan)
	}
	done <- scan
}

func printUsage() {
	println("pget will try to detect a pattern in given URL and download files from similar URLs")
	println("")
	println("usage: pget <url>")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	url := os.Args[1]
	pattern := FindPattern(url)
	if pattern == nil {
		fmt.Printf("No pattern found in \"%s\"\n", url)
		os.Exit(1)
	}
	number, format, _ := ParseIndexAndFormat(pattern)
	fmt.Printf("Parse results: number %d, format %s\n", number, format)

	chanA, chanB := make(chan int), make(chan int)
	go Crawler(number, format, pattern, func(index int) int { return index - 1 }, chanA)
	go Crawler(number + 1, format, pattern, func(index int) int { return index + 1 }, chanB)
	fmt.Printf("Crawler 1 stopped at %d, crawler 2 stopped at %d\n",
		<-chanA, <-chanB)
	os.Exit(0)
}
