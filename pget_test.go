package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNumberPattern(t *testing.T) {
	pattern := FindPattern("http://www.site.com/path/pic_23.jpg")
	if pattern.match != "23" {
		t.Fail()
	}
}

func TestNumberPatternWithQuery(t *testing.T) {
	pattern := FindPattern("http://www.site.com/path/pic_23.jpg?param=123")
	if pattern.match != "23" {
		t.Fail()
	}
}

func TestNonMatchingUrl(t *testing.T) {
	pattern := FindPattern("http://www.site.com/path/pic_XX.jpg")
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
		if n ==  1 && number > 0 && number < 20 {
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

func decrement(index int) int { return index - 1 }

func TestSuccessfulCrawl(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	c := make(chan int)

	pattern := FindPattern(mkUrl(s, 10))
	number, format, _ := ParseIndexAndFormat(pattern)

	go Crawler(number, format, pattern, decrement, c)
	if <-c != 0 {
		t.Fail()
	}
}

func TestUnsuccessfulCrawl(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	c := make(chan int)

	pattern := FindPattern(mkUrl(s, 100))
	number, format, _ := ParseIndexAndFormat(pattern)

	go Crawler(number, format, pattern, decrement, c)
	if <-c == 0 {
		t.Fail()
	}
}

func TestSuccessfulPaddedCrawl(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	c := make(chan int)

	pattern := FindPattern(mkPaddedUrl(s, 10, 5))
	number, format, _ := ParseIndexAndFormat(pattern)

	go Crawler(number, format, pattern, decrement, c)
	if <-c != 0 {
		t.Fail()
	}
}

func TestUnsuccessfulPaddedCrawl(t *testing.T) {
	s := loopbacKServer()
	defer s.Close()

	c := make(chan int)

	pattern := FindPattern(mkPaddedUrl(s, 100, 5))
	number, format, _ := ParseIndexAndFormat(pattern)

	go Crawler(number, format, pattern, decrement, c)
	if <-c == 0 {
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
