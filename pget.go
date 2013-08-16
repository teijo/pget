package main

import (
	"os"
	"fmt"
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

func ParseIndexString(index string) (number int64, maxPadding int, paddingFound bool, err error) {
	maxPadding = len(index)
	paddingFound = maxPadding > 1 &&(index[0] == '0')
	number, err = strconv.ParseInt(index, 10, 16)
	return number, maxPadding, paddingFound, err
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
	fmt.Printf("number %d, maxPadding %d, paddingFound %t\n", number, maxPadding, paddingFound)
	os.Exit(0)
}
