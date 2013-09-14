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
	prefix string
	match  string
	suffix string
}

func Dump(variable interface{}) {
	fmt.Printf("%+v\n", variable)
}

func fileName(u *url.URL) (string, string) {
	var urlSegments []string = strings.Split(u.Path, "/")
	last := len(urlSegments) - 1
	return strings.Join(urlSegments[:last], "/"), urlSegments[last]
}

func extractIndex(str string) (bool, *Pattern) {
	match := regexp.MustCompile("(^[^\\d]*)([\\d]+)([^\\d]*)$").FindStringSubmatch(str)
	success := len(match) > 0 && len(match[2]) > 0
	var pat *Pattern = nil
	if success {
		pat = &Pattern{match[1], match[2], match[3]}
	}
	return success, pat
}

func condQueryJoin(a string, u *url.URL) string {
	if len(u.RawQuery) > 0 {
		return fmt.Sprintf("%s?%s", a, u.RawQuery)
	}
	return fmt.Sprintf("%s", a)
}

func tryFindFile(u *url.URL) (bool, *Pattern) {
	filePath, file := fileName(u)
	if match, pat := extractIndex(file); match {
		return true, &Pattern{fmt.Sprintf("http://%s%s/%s", u.Host, filePath, pat.prefix), pat.match, condQueryJoin(pat.suffix, u)}
	}
	return false, nil
}

func tryFindQuery(u *url.URL) (bool, *Pattern) {
	if match, pat := extractIndex(u.RawQuery); match {
		return true, &Pattern{fmt.Sprintf("http://%s%s?%s", u.Host, u.Path, pat.prefix), pat.match, pat.suffix}
	}
	return false, nil
}

func tryFindPath(u *url.URL) (bool, *Pattern) {
	if match, pat := extractIndex(u.Path); match {
		return true, &Pattern{fmt.Sprintf("http://%s%s", u.Host, pat.prefix), pat.match, condQueryJoin(pat.suffix, u)}
	}
	return false, nil
}

func FindPattern(urlString string) (*Pattern, error) {
	u, err := url.Parse(urlString)
	if (err != nil) {
		return nil, err
	}

	fns := [](func (*url.URL)(bool, *Pattern)){tryFindFile, tryFindQuery, tryFindPath}
	for _, fn := range fns {
		if match, pat := fn(u); match {
			return pat, nil
		}
	}

	return nil, fmt.Errorf("No pattern found in \"%s\"\n", urlString)
}

func probeExistence(url string) (bool, error) {
	res, err := http.Head(url)
	if err != nil {
		return false, err
	}
	return (res.StatusCode == http.StatusOK), nil
}

func StrToI(number string) int {
	var i64 int64 = 0
	i64, _ = strconv.ParseInt(number, 10, 16)
	return int(i64)
}

func TestPadding(urlPrefix string, urlSuffix string, testIndex int) bool {
	format := fmt.Sprintf("%%0%dd", IntLen(testIndex))
	closest := ClosestShorterInt(testIndex)
	paddedIndexString := fmt.Sprintf(format, closest)
	found, _ := probeExistence(fmt.Sprintf("%s%s%s", urlPrefix, paddedIndexString, urlSuffix))
	return found
}

func ParseIndexAndFormat(pattern *Pattern) (number int, format string, err error) {
	number = StrToI(pattern.match)
	maxPadding := len(pattern.match)
	if maxPadding > 1 && (pattern.match[0] == '0') {
		format = fmt.Sprintf("%%0%dd", maxPadding)
	} else {
		format = "%d"
	}

	return number, format, err
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
	filename := urlSegments[len(urlSegments) - 1]
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
	if (number == 0) {
		return 1
	}
	return int(math.Floor(math.Log10(math.Abs(float64(number))))) + 1
}

func ClosestShorterInt(number int) int {
	if number < 10 {
		return -1
	}
	multiplier := (IntLen(number) - 1)*10
	return (number/multiplier)*multiplier - 1
}

func BuildUrl(scan int, format string, pattern *Pattern) string {
	printFmt := fmt.Sprintf("%%s%s%%s", format)
	return fmt.Sprintf(printFmt, pattern.prefix, scan, pattern.suffix)
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
	pattern, err := FindPattern(url)
	if err != nil {
		fmt.Printf("%s\n", err)
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
