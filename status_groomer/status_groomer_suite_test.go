package status_groomer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestStatusGroomer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Status Groomer Suite")
}
