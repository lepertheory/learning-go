package getopt_long

import (
	"math"
	"math/rand"
	"strings"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGetoptLong(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GetoptLong Suite")
}

type LetFunc[T any] func() T
func let[T any](f LetFunc[T]) func() T {
	var once sync.Once
	var val T

	return func() T {
		once.Do(func() {
			val = f()
		})
		return val
	}
}
func v[T any](f LetFunc[T]) T {
	return f()
}

func randomString(length int, rng *rand.Rand) string {
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteRune(rune(rng.Int()))
	}
	return sb.String()
}

var _ = Describe("getCString", func() {
	var inputLen int
	var input LetFunc[string]
	var rng *rand.Rand

	BeforeEach(func() {
		rng = rand.New(rand.NewSource(GinkgoRandomSeed()))
		inputLen = rng.Intn(20)
		input = let(func() string { return randomString(inputLen, rng) })
	})

	When("working on a sub-MB string", func() {
		BeforeEach(func() { inputLen = rng.Intn(1024 * 1024) })

		It("works", func() {
			Expect(toGoString(getCString(v(input)))).To(Equal(v(input)))
		})
	})

	When("working on an empty string", func() {
		BeforeEach(func() { inputLen = 0 })
		It("returns an pointer to a null byte.", func() {
			Expect(toGoString(getCString(v(input)))).To(BeEmpty())
		})
	})
})

var _ = Describe("toOptString", func() {
	DescribeTable(
		"valid options",
		func(input ArgRequirement, expected string) {
			Expect(input.toOptstring()).To(Equal(expected))
		},
		Entry("when an argument is not allowed", ArgNotAllowed, ""),
		Entry("when an argument is optional", ArgOptional, "::"),
		Entry("when an argument is required", ArgRequired, ":"),
	)

	When("an invalid value is used", func() {
		It("panics", func() {
			var invalid ArgRequirement
			invalid = math.MaxInt
			Expect(func() { invalid.toOptstring() }).To(Panic())
		})
	})
})
