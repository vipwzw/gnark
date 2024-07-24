package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func counOne1(n uint64) int {
	count := 0
	for n != 0 {
		count += int(n & 1)
		n >>= 1
	}
	return count
}

// 用分治的方法实现
func counOne2(x uint64) int {
	if x == 0 || x == 1 {
		return int(x)
	}
	if x == 2 || x == 3 {
		// x= 2 -> 1 , x = 3 -> 2
		return int(x - 1)
	}
	if x >= 1<<32 {
		return counOne2(x>>32) + counOne2(x&0xffffffff)
	} else if x >= 1<<16 {
		return counOne2(x>>16) + counOne2(x&0xffff)
	} else if x >= 1<<8 {
		return counOne2(x>>8) + counOne2(x&0xff)
	} else if x >= 1<<4 {
		return counOne2(x>>4) + counOne2(x&0xf)
	} else {
		return counOne2(x>>2) + counOne2(x&0x3)
	}
}

func counOne3(x uint64) int {
	x = (x & 0x5555555555555555) + ((x >> 1) & 0x5555555555555555)
	x = (x & 0x3333333333333333) + ((x >> 2) & 0x3333333333333333)
	x = (x & 0x0f0f0f0f0f0f0f0f) + ((x >> 4) & 0x0f0f0f0f0f0f0f0f)
	x = (x & 0x00ff00ff00ff00ff) + ((x >> 8) & 0x00ff00ff00ff00ff)
	x = (x & 0x0000ffff0000ffff) + ((x >> 16) & 0x0000ffff0000ffff)
	x = (x & 0x00000000ffffffff) + ((x >> 32) & 0x00000000ffffffff)
	return int(x)
}

func TestCountOne(t *testing.T) {
	assert.Equal(t, 1, counOne1(1))
	assert.Equal(t, 2, counOne1(10))
	assert.Equal(t, 64, counOne1(1<<64-1))

	assert.Equal(t, 1, counOne2(1))
	assert.Equal(t, 2, counOne2(10))
	assert.Equal(t, 64, counOne2(1<<64-1))

	assert.Equal(t, 1, counOne3(1))
	assert.Equal(t, 2, counOne3(10))
	assert.Equal(t, 64, counOne3(1<<64-1))
}

func BenchmarkCountOne1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		counOne1(1<<64 - 1)
	}
}

func BenchmarkCountOne2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		counOne2(1<<64 - 1)
	}
}

func BenchmarkCountOne3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		counOne3(1<<64 - 1)
	}
}
