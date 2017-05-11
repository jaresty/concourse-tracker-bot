package parser

import (
	"fmt"
	"strings"
)

func makeGroupRegex(regexes []string) string {
	wrappedRegexes := make([]string, len(regexes))
	for i, regex := range regexes {
		wrappedRegexes[i] = fmt.Sprintf("(%s)", regex)
	}

	return strings.Join(wrappedRegexes, "|")
}

func Parse(inputMap map[string][]string) map[string]string {
	outputMap := make(map[string]string)

	for group, regexes := range inputMap {
		outputMap[makeGroupRegex(regexes)] = group
	}
	return outputMap
}
