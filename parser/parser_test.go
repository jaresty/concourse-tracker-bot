package parser_test

import (
	"fmt"

	. "github.com/jaresty/concourse-tracker-bot/parser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	It("Prepares a data structure useful for grouping by name", func() {
		regexGroupa1String := ".*-groupa"
		regexGroupa2String := "groupa-.*"
		regexGroupbString := "groupb-.*-groupb"
		inputMap := make(map[string][]string)
		inputMap["groupa"] = []string{regexGroupa1String, regexGroupa2String}
		inputMap["groupb"] = []string{regexGroupbString}
		outputMap := Parse(inputMap)
		expectedOutputMap := make(map[string]string)
		expectedOutputMap[fmt.Sprintf("(%s)|(%s)", regexGroupa1String, regexGroupa2String)] = "groupa"
		expectedOutputMap[fmt.Sprintf("(%s)", regexGroupbString)] = "groupb"
		Expect(outputMap).To(Equal(expectedOutputMap))
	})
})
