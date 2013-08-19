package main

import "testing"

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
	number, maxPadding, paddingFound, _ := ParseIndexString(&Pattern{match: "00321"})
	if number != 321 || maxPadding != 5 || paddingFound != true {
		t.Fail()
	}
}

func TestFindPatternWithNoPadding(t *testing.T) {
	number, maxPadding, paddingFound, _ := ParseIndexString(&Pattern{match: "123"})
	if number != 123 || maxPadding != 3 || paddingFound != false {
		t.Fail()
	}
}

func TestFindPatternWithZero(t *testing.T) {
	number, maxPadding, paddingFound, _ := ParseIndexString(&Pattern{match: "0"})
	if number != 0 || maxPadding != 1 || paddingFound != false {
		t.Fail()
	}
}

func TestFindPatternWithNonNumber(t *testing.T) {
	_, _, _, err := ParseIndexString(&Pattern{match: "XYZ"})
	if err == nil {
		t.Fail()
	}
}

func decrement(index int) int { return index - 1 }

func TestSuccessfulCrawl(t *testing.T) {
	c := make(chan int)
	go Crawler(2, &Pattern{urlPrefix: "<", match: "5", urlSuffix: ">"}, decrement, c)
	if <-c != 0 {
		t.Fail()
	}
}

func TestUnsuccessfulCrawl(t *testing.T) {
	c := make(chan int)
	go Crawler(100, &Pattern{urlPrefix: "<", match: "100", urlSuffix: ">"}, decrement, c)
	if <-c != 100 {
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