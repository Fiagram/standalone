package logic_account_test

import (
	"math/rand"
	"strings"
	"time"
)

func RandomString(length uint) string {
	g := rand.New(rand.NewSource(time.Now().UnixNano()))
	const alphabet = "qazwsxedcrfvtgbyhnujmikolp"
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < int(length); i++ {
		c := alphabet[g.Intn(k)]
		sb.WriteByte(c)
	}
	return strings.TrimSpace(sb.String())
}
