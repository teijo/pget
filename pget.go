package main

import (
	"os"
	"fmt"
	"math"
	"time"
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

type DownloadResult struct {
	error error
	bytes int64
	count int
	duration time.Duration
}

func (d DownloadResult) Ok(bytes int64, count int, duration time.Duration) DownloadResult {
	d.bytes = bytes
	d.count = count
	d.duration = duration
	return d
}

func (d DownloadResult) Fail(err error) DownloadResult {
	d.error = err
	return d
}

func (d* DownloadResult) Add(d2 DownloadResult) {
	d.count += d2.count
	d.bytes += d2.bytes
	d.duration += d2.duration
}

func (d DownloadResult) toString() string {
	return fmt.Sprintf("[%d] %.02fkBps", d.bytes, (float64(d.bytes) / 1024.0 / d.duration.Seconds()))
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

	return nil, fmt.Errorf("No pattern detected in \"%s\"", urlString)
}

func probeExistence(url string) (bool, error) {
	res, err := http.Head(url)
	if err != nil {
		return false, err
	}
	return (res.StatusCode == http.StatusOK), nil
}

func TestPadding(urlPrefix string, urlSuffix string, testIndex int) bool {
	format := fmt.Sprintf("%%0%dd", IntLen(testIndex))
	closest := ClosestShorterInt(testIndex)
	paddedIndexString := fmt.Sprintf(format, closest)
	found, _ := probeExistence(fmt.Sprintf("%s%s%s", urlPrefix, paddedIndexString, urlSuffix))
	return found
}

func ParseIndexAndFormat(pattern *Pattern) (number int, format string, err error) {
	number, err = strconv.Atoi(pattern.match)
	if err != nil {
		return 0, "", err
	}
	maxPadding := len(pattern.match)
	if maxPadding > 1 && (pattern.match[0] == '0') {
		format = fmt.Sprintf("%%0%dd", maxPadding)
	} else {
		format = "%d"
	}

	return number, format, err
}

func downloadUrl(u *url.URL, filename string) DownloadResult {
	result := DownloadResult{}
	start := time.Now()
	res, err := http.Get(u.String())
	if err != nil {
		return result.Fail(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return result.Fail(fmt.Errorf("Status %d", res.StatusCode))
	}

	body, err := ioutil.ReadAll(res.Body)
	downloaded := 1
	if (err != nil) {
		return result.Fail(err)
	} else if (len(body) == 0) {
		downloaded = 0
	} else if err = ioutil.WriteFile(filename, body, os.ModePerm); err != nil {
		return result.Fail(err)
	}
	end := time.Now()

	return result.Ok(res.ContentLength, downloaded, end.Sub(start))
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

func buildUrl(scan int, format string, pattern *Pattern) *url.URL {
	printFmt := fmt.Sprintf("%%s%s%%s", format)
	u, _ := url.Parse(fmt.Sprintf(printFmt, pattern.prefix, scan, pattern.suffix))
	return u
}

func Crawler(scan int, format string, pattern *Pattern, next func (int) int, done chan DownloadResult) {
	channel := make(chan DownloadResult)
	total := DownloadResult{}
	for {
		u := buildUrl(scan, format, pattern)
		_, filename := fileName(u)
		go func() { channel <- downloadUrl(u, filename) }()
		result := <-channel
		if result.error != nil {
			fmt.Printf("GET %s -> [%s]\n", u.String(), result.error)
			break
		}
		total.Add(result)
		fmt.Printf("GET %s %s -> %s\n", u.String(), result.toString(), filename)
		scan = next(scan)
	}
	done <- total
}

func printUsage() {
	println("pget will try to detect a pattern in given URL and download files from similar URLs")
	println("")
	println("usage: pget <url>")
}

func dualCrawl(number int, format string, pattern *Pattern) DownloadResult {
	result := DownloadResult{}
	desc, asc := make(chan DownloadResult), make(chan DownloadResult)
	go Crawler(number, format, pattern, func(index int) int { return index - 1 }, desc)
	go Crawler(number + 1, format, pattern, func(index int) int { return index + 1 }, asc)
	result.Add(<- desc)
	result.Add(<- asc)
	return result
}

func StartCrawl(number int, format string, pattern *Pattern, fn (func (int, string, *Pattern)DownloadResult)) (DownloadResult, error) {
	urlString := fmt.Sprintf("%s%s%s", pattern.prefix, pattern.match, pattern.suffix)
	found, err := probeExistence(urlString)
	if err != nil {
		return DownloadResult{}, err
	} else if !found {
		return DownloadResult{}, fmt.Errorf("Resource \"%s\" not found", urlString)
	}
	return fn(number, format, pattern), nil
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
	fmt.Printf("Detected pattern %s starting from index %d\n", format, number)

	result, err := StartCrawl(number, format, pattern, dualCrawl)
	if err != nil {
		fmt.Printf("Crawling failed: %s\n", err)
	}
	fmt.Printf("Downloaded %d files at %s\n", result.count, result.toString())
	os.Exit(0)
}
