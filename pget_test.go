package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"strconv"
)

func TestNumberPattern(t *testing.T) {
	pattern, _ := FindPattern("http://www.site.com/path/pic_23.jpg")
	if pattern.match != "23" {
		t.Fail()
	}
}

func TestNumberPatternWithQuery(t *testing.T) {
	pattern, _ := FindPattern("http://www.site.com/path/pic_23.jpg?param=123")
	if pattern.match != "23" {
		t.Fail()
	}
}

func TestPrecedenceForFile(t *testing.T) {
	pattern, _ := FindPattern("http://www.site.com/path/1/2.rar?param=3")
	if pattern.match != "2" {
		println(pattern.match)
		t.Fail()
	}
}

func TestPrecedenceForQuery(t *testing.T) {
	pattern, _ := FindPattern("http://www.site.com/path/1/a.zip?a=b&param=3")
	if pattern.match != "3" {
		println(pattern.match)
		t.Fail()
	}
}

func TestPrecedenceForPath(t *testing.T) {
	pattern, _ := FindPattern("http://www.site.com/path/1/a.zip?a=b")
	if pattern.match != "1" {
		println(pattern.match)
		t.Fail()
	}
}

func TestPatternFinderIntegrityInFile(t *testing.T) {
	testUrl := "http://www.site.com/path/1.zip"
	pattern, _ := FindPattern(testUrl)
	result := fmt.Sprintf("%s%s%s", pattern.prefix, pattern.match, pattern.suffix)
	if result != testUrl {
		println(result)
		t.Fail()
	}
}

func TestPatternFinderIntegrityInQuery(t *testing.T) {
	testUrl := "http://www.site.com/path/a.zip?a=b&x=1&z=q"
	pattern, _ := FindPattern(testUrl)
	result := fmt.Sprintf("%s%s%s", pattern.prefix, pattern.match, pattern.suffix)
	if result != testUrl {
		println(result)
		t.Fail()
	}
}

func TestPatternFinderIntegrityInPath(t *testing.T) {
	testUrl := "http://www.site.com/path/1/a.zip"
	pattern, _ := FindPattern(testUrl)
	result := fmt.Sprintf("%s%s%s", pattern.prefix, pattern.match, pattern.suffix)
	if result != testUrl {
		println(result)
		t.Fail()
	}
}

func TestNonMatchingUrl(t *testing.T) {
	pattern, _ := FindPattern("http://www.site.com/path/pic_XX.jpg")
	if pattern != nil {
		t.Fail()
	}
}

func TestFindPatternWithPadding(t *testing.T) {
	number, format, _ := ParseIndexAndFormat(&Pattern{match: "00321"})
	if number != 321 || format != "%05d" {
		t.Fail()
	}
}

func TestFindPatternWithNoPadding(t *testing.T) {
	number, format, _ := ParseIndexAndFormat(&Pattern{match: "123"})
	if number != 123 || format != "%d" {
		t.Fail()
	}
}

func TestFindPatternWithZero(t *testing.T) {
	number, format, _ := ParseIndexAndFormat(&Pattern{match: "0"})
	if number != 0 || format != "%d" {
		t.Fail()
	}
}

func TestFindPatternWithNonNumber(t *testing.T) {
	_, _, err := ParseIndexAndFormat(&Pattern{match: "XYZ"})
	if err == nil {
		t.Fail()
	}
}

func loopbacKServer() *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusNotFound
		number := 0
		n, _ := fmt.Sscanf(r.URL.Path, "/file/%d.ext", &number)
		if n == 1 && number > 0 && number < 20 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
	}))
	return s
}

func mkPaddedUrl(s *httptest.Server, number int, padding int) string {
	format := fmt.Sprintf("%%0%dd", padding)
	return fmt.Sprintf(fmt.Sprintf("%s/file/%s.ext", s.URL, format), number)
}

func mkUrl(s *httptest.Server, number int) string {
	return fmt.Sprintf("%s/file/%d.ext", s.URL, number)
}

func TestSuccessfulLoopback(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	res, _ := http.Get(mkUrl(s, 3))
	if res.StatusCode != 200 {
		t.Fail()
	}
}

func TestFailingLoopback(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	res, _ := http.Get(mkUrl(s, 100))
	if res.StatusCode != 404 {
		t.Fail()
	}
}

func TestPaddingProbe(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	pattern, _ := FindPattern(mkUrl(s, 10))
	value, _ := strconv.Atoi(pattern.match)
	if TestPadding(pattern.prefix, pattern.suffix, value) != true {
		t.Fail()
	}
}

func decrement(index int) int { return index - 1 }

func decrementCrawl(number int, format string, pattern *Pattern) int {
	c := make(chan int)
	go Crawler(number, format, pattern, decrement, c)
	return <-c
}

func TestSuccessfulCrawl(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	pattern, _ := FindPattern(mkUrl(s, 10))
	number, format, _ := ParseIndexAndFormat(pattern)

	res, err := StartCrawl(number, format, pattern, decrementCrawl)
	if res != 10 || err != nil {
		t.Fail()
	}
}

func TestUnsuccessfulCrawl(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	pattern, _ := FindPattern(mkUrl(s, 100))
	number, format, _ := ParseIndexAndFormat(pattern)

	_, err := StartCrawl(number, format, pattern, decrementCrawl)
	if err == nil {
		t.Fail()
	}
}

func TestSuccessfulPaddedCrawl(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	pattern, _ := FindPattern(mkPaddedUrl(s, 10, 5))
	number, format, _ := ParseIndexAndFormat(pattern)

	res, err := StartCrawl(number, format, pattern, decrementCrawl)
	if res != 10 || err != nil {
		t.Fail()
	}
}

func TestUnsuccessfulPaddedCrawl(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	pattern, _ := FindPattern(mkPaddedUrl(s, 100, 5))
	number, format, _ := ParseIndexAndFormat(pattern)

	_, err := StartCrawl(number, format, pattern, decrementCrawl)
	if err == nil {
		t.Fail()
	}
}

func TestShorterLessThanTen(t *testing.T) {
	if ClosestShorterInt(5) != -1 {
		t.Fail()
	}
}

func TestShorterTen(t *testing.T) {
	if ClosestShorterInt(10) != 9 {
		t.Fail()
	}
}

func TestShorterTenThousand(t *testing.T) {
	if ClosestShorterInt(10000) != 9999 {
		t.Fail()
	}
}

func assert(t *testing.T, a int, b int) {
	if a != b {
		fmt.Printf("%d not equal to %d\n", a, b)
		t.Fail()
	}
}

func TestIntLengthZero(t *testing.T) {
	assert(t, IntLen(0), 1)
}

func TestIntLengthTen(t *testing.T) {
	assert(t, IntLen(10), 2)
}

func TestIntLengthOffByOne(t *testing.T) {
	assert(t, IntLen(999), 3)
}

func TestIntLengthNegative(t *testing.T) {
	assert(t, IntLen(-10), 2)
}
