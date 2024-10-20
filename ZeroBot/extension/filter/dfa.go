package filter

import (
	"bufio"
	"os"
	"strings"
)

type DFAFilter struct {
	keywordChains map[rune]map[rune]interface{}
	delimit       rune
}

func NewDFAFilter() *DFAFilter {
	return &DFAFilter{
		keywordChains: make(map[rune]map[rune]interface{}),
		delimit:       '\x00',
	}
}

func (f *DFAFilter) Add(keyword string) {
	chars := []rune(strings.TrimSpace(keyword))
	if len(chars) == 0 {
		return
	}

	level := f.keywordChains
	for i, char := range chars {
		if nextLevel, found := level[char]; found {
			level = nextLevel
		} else {
			for j := i; j < len(chars); j++ {
				if level[chars[j]] == nil {
					level[chars[j]] = make(map[rune]interface{})
				}
				level = level[chars[j]].(map[rune]interface{})
			}
			level[chars[len(chars)-1]] = map[rune]interface{}{f.delimit: nil}
			break
		}
	}
	level[f.delimit] = nil
}

func (f *DFAFilter) Parse(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f.Add(scanner.Text())
	}

	return scanner.Err()
}

func (f *DFAFilter) Filter(message string, repl string) string {
	ret := []rune{}
	start := 0

	for start < len(message) {
		level := f.keywordChains
		stepIns := 0

		for _, char := range message[start:] {
			if nextLevel, found := level[char]; found {
				stepIns++
				if _, isDelim := nextLevel[f.delimit]; !isDelim {
					level = nextLevel.(map[rune]interface{})
				} else {
					ret = append(ret, []rune(strings.Repeat(repl, stepIns))...)
					start += stepIns - 1
					break
				}
			} else {
				ret = append(ret, char)
				break
			}
		}

		if stepIns == 0 {
			ret = append(ret, rune(message[start]))
		}
		start++
	}

	return string(ret)
}
