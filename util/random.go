package util

import (
	"math/rand"
	"strings"
	"time"
)

var random *rand.Rand

const alphabet = "abcdefghijklmnopqrstuvxz"

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Random int between min and max
func RandomInt(min, max int64) int64 {
	return min + random.Int63n(max-min+1)
}

// Random string with len == n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[random.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// Random owner
func RandomOnwer() string {
	return RandomString(6)
}

// Random money amount
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// Random type of currency
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "BRL"}
	n := len(currencies)
	return currencies[random.Intn(n)]
}
