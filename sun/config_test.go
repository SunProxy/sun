package sun

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkGenKey(b *testing.B) {
	Chars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i < b.N; i++ {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		var key = make([]rune, 25)
		for ind := range key {
			key[ind] = rune(Chars[rnd.Intn(len(Chars))])
		}
	}
}
