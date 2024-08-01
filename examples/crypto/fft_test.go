package crypto

import (
	"math"
	"math/cmplx"
	"testing"
)

// FFT 实现
func FFT(a []complex128, invert bool) {
	n := len(a)
	if n == 1 {
		return
	}

	// 将数组分成偶数和奇数部分
	half := n / 2
	even := make([]complex128, half)
	odd := make([]complex128, half)
	for i := 0; i < half; i++ {
		even[i] = a[i*2]
		odd[i] = a[i*2+1]
	}

	// 递归调用 FFT
	FFT(even, invert)
	FFT(odd, invert)

	// 计算 FFT
	angle := 2 * math.Pi / float64(n)
	if invert {
		angle = -angle
	}
	w := complex(1, 0)
	wn := cmplx.Exp(complex(0, angle))
	for i := 0; i < half; i++ {
		a[i] = even[i] + w*odd[i]
		a[i+half] = even[i] - w*odd[i]
		if invert {
			a[i] /= 2
			a[i+half] /= 2
		}
		w *= wn
	}
}

// 多项式乘法
func multiplyPolynomials(a, b []complex128) []complex128 {
	n := 1
	for n < len(a)+len(b) {
		n <<= 1
	}
	a = append(a, make([]complex128, n-len(a))...)
	b = append(b, make([]complex128, n-len(b))...)

	FFT(a, false)
	FFT(b, false)

	for i := range a {
		a[i] *= b[i]
	}

	FFT(a, true)
	return a
}

func TestFFT(t *testing.T) {
	a := []complex128{1, 2, 3, 4}
	b := []complex128{1, 2, 3, 4}
	result := multiplyPolynomials(a, b)
	for _, v := range result {
		t.Logf("%.0f ", real(v))
	}
}
