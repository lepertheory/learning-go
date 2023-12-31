package getopt_long

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LetFunc[T any] func() T

var seed = flag.String("seed", "", "Set the rng seed.")

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

type GetoptLongSuite struct {
	suite.Suite
	Rng *rand.Rand
}

func (suite *GetoptLongSuite) randomString(length int) string {
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteRune(rune(suite.Rng.Int()))
	}
	return sb.String()
}

func TestGetoptLongSuite(t *testing.T) {
	suite.Run(t, new(GetoptLongSuite))
}

func (suite *GetoptLongSuite) SetupSuite() {
	var n int64
	var err error
	if len(*seed) > 0 {
		n, err = strconv.ParseInt(*seed, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Error parsing seed %s: %s", *seed, err))
		}
	} else {
		n = time.Now().UnixNano()
	}
	suite.Rng = rand.New(rand.NewSource(n))
	suite.T().Logf("-seed=%d", n)
}

func (suite *GetoptLongSuite) TestGetCString_Basic() {
	testValue := let(func() string {
		retval := suite.randomString(suite.Rng.Intn(1024 * 1024))
		return retval
	})
	testActual, testFree := getCString(v(testValue))
	defer testFree()
	assert.Equal(suite.T(), getGoString(testActual), v(testValue))
}

func (suite *GetoptLongSuite) TestToOptstring() {
	assert.Equal(suite.T(), ArgNotAllowed.toOptstring(), "")
	assert.Equal(suite.T(), ArgOptional.toOptstring(), "::")
	assert.Equal(suite.T(), ArgRequired.toOptstring(), ":")
	var bloop ArgRequirement
	bloop = 7
	assert.PanicsWithErrorf(
		suite.T(),
		fmt.Sprintf("Unexpected ArgRequirement value: %d", bloop),
		func() { bloop.toOptstring() },
		"Unexpected ArgRequirement value: %d",
		bloop,
	)
}
