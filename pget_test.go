package main

import "testing"

func TestNumberPattern(t *testing.T) {
	if FindPattern("http://www.site.com/path/pic_23.jpg") != "23" {
		t.Fail()
	}
}

func TestNumberPatternWithQuery(t *testing.T) {
	if FindPattern("http://www.site.com/path/pic_23.jpg?param=123") != "23" {
		t.Fail()
	}
}

func TestNonMatchingUrl(t *testing.T) {
	if FindPattern("http://www.site.com/path/pic_XX.jpg") != "" {
		t.Fail()
	}
}

func TestFindPatternWithPadding(t *testing.T) {
	number, maxPadding, paddingFound, _ := ParseIndexString("00321")
	if number != 321 || maxPadding != 5 || paddingFound != true {
		t.Fail()
	}
}

func TestFindPatternWithNoPadding(t *testing.T) {
	number, maxPadding, paddingFound, _ := ParseIndexString("123")
	if number != 123 || maxPadding != 3 || paddingFound != false {
		t.Fail()
	}
}

func TestFindPatternWithZero(t *testing.T) {
	number, maxPadding, paddingFound, _ := ParseIndexString("0")
	if number != 0 || maxPadding != 1 || paddingFound != false {
		t.Fail()
	}
}

func TestFindPatternWithNonNumber(t *testing.T) {
	_, _, _, err := ParseIndexString("XYZ")
	if err == nil {
		t.Fail()
	}
}
