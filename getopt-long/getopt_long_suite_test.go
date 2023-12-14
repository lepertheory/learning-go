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

var lets = make(map[*sync.Once]bool)
type LetFunc[T any] func() T
func let[T any](f LetFunc[T]) func() T {
	var once sync.Once
	var val T

	return func() T {
		once.Do(func() {
			lets[&once] = true
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

var _ = Describe("ArgRequirement", func() {
	Describe("toOptString", func() {
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
})

func pointer[T any](val T) *T {
	return &val
}

var _ = Describe("GetOpt", func() {
	var arguments LetFunc[[]string]
	var getopt LetFunc[*GetOpt]
	var optHelp LetFunc[Option]
	var optHelpArg LetFunc[ArgRequirement]
	var optHelpName LetFunc[*string]
	var optHelpRequired LetFunc[bool]
	var optHelpShort LetFunc[*string]
	var optList LetFunc[Option]
	var optListArg LetFunc[ArgRequirement]
	var optListName LetFunc[*string]
	var optListRequired LetFunc[bool]
	var optListShort LetFunc[*string]
	var options LetFunc[[]Option]
	var programName LetFunc[string]

	BeforeEach(func() {
		programName = let(func() string { return "program" })

		arguments = let(func() []string { return []string{v(programName), "--" + *v(optHelpName)} })

		optHelpName = let(func() *string { return pointer("help") })
		optHelpShort = let(func() *string { return pointer("h") })
		optHelpRequired = let(func() bool { return false })
		optHelpArg = let(func() ArgRequirement { return ArgNotAllowed })
		optHelp = let(func() Option {
			return Option{
				Name:     v(optHelpName),
				Short:    v(optHelpShort),
				Required: v(optHelpRequired),
				Arg:      v(optHelpArg),
			}
		})

		optListName = let(func() *string { return pointer("list") })
		optListShort = let(func() *string { return pointer("l") })
		optListRequired = let(func() bool { return false })
		optListArg = let(func() ArgRequirement { return ArgOptional })
		optList = let(func() Option {
			return Option{
				Name:     v(optListName),
				Short:    v(optListShort),
				Required: v(optListRequired),
				Arg:      v(optListArg),
			}
		})

		options = let(func() []Option {
			return []Option{
				v(optHelp),
				v(optList),
			}
		})

		getopt = let(func() *GetOpt {
			return pointer(GetOpt{
				Options: v(options),
				Arguments: v(arguments),
			})
		})
	})

	Describe("Process", func() {
		When("using a basic set of options", func() {
			It("works with --help", func() {
				v(getopt).Process()
				Expect(v(getopt).Results).To(Equal(map[Option]OptionResult{v(optHelp): {SetCount: 1}}))
			})
		})

		When("the help option is misspelled", func() {
			BeforeEach(func() {
				arguments = let(func() []string { return []string{v(programName), "--help"} })
				optHelpName = let(func() *string { return pointer("hlep") })
			})
			It("does something else", func() {
				// TODO: Test better.
				v(getopt).Process()
			})
		})
	})
})
