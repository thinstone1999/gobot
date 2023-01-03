package util

import (
	"math/rand"
	"strings"
)

const (
	min   = 0x4E00
	max   = 0x9FA5
	Range = max - min + 1
)

func GenChineseName(nameLen int) string {
	var buf strings.Builder
	for i := 0; i < nameLen; i++ {
		char := rand.Int31n(Range) + min
		buf.WriteRune(char)
	}

	return buf.String()
}
