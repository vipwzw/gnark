package crypto

import (
	"fmt"
	"testing"
)

// GF2mPolynomial 表示 GF(2^m) 中的多项式
type GF2mPolynomial struct {
	Coefficients []int
}

// NewGF2mPolynomial 创建一个新的 GF(2^m) 多项式
func NewGF2mPolynomial(coefficients []int) GF2mPolynomial {
	return GF2mPolynomial{Coefficients: coefficients}
}

// Add 多项式加法
func (p GF2mPolynomial) Add(q GF2mPolynomial) GF2mPolynomial {
	m := len(p.Coefficients)
	n := len(q.Coefficients)
	maxLen := m
	if n > m {
		maxLen = n
	}
	resultCoefficients := make([]int, maxLen)

	for i := 0; i < maxLen; i++ {
		if i < m {
			resultCoefficients[i] ^= p.Coefficients[i]
		}
		if i < n {
			resultCoefficients[i] ^= q.Coefficients[i]
		}
	}

	return NewGF2mPolynomial(resultCoefficients)
}

// Multiply 多项式乘法
func (p GF2mPolynomial) Multiply(q GF2mPolynomial, modPoly GF2mPolynomial) GF2mPolynomial {
	m := len(p.Coefficients)
	n := len(q.Coefficients)
	resultCoefficients := make([]int, m+n-1)

	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			resultCoefficients[i+j] ^= p.Coefficients[i] * q.Coefficients[j]
		}
	}

	result := NewGF2mPolynomial(resultCoefficients)
	return result.Mod(modPoly)
}

// Mod 对多项式进行模约简
func (p GF2mPolynomial) Mod(modPoly GF2mPolynomial) GF2mPolynomial {
	m := len(p.Coefficients)
	n := len(modPoly.Coefficients)
	if m < n {
		return p
	}

	for i := m - 1; i >= n-1; i-- {
		if p.Coefficients[i] != 0 {
			for j := 0; j < n; j++ {
				p.Coefficients[i-j] ^= modPoly.Coefficients[n-1-j]
			}
		}
	}

	return NewGF2mPolynomial(p.Coefficients[:n-1])
}

func TestGF2(t *testing.T) {
	// 定义有限域 GF(2^3) 中的多项式
	p := NewGF2mPolynomial([]int{1, 0, 1}) // x^2 + 1
	q := NewGF2mPolynomial([]int{1, 1})    // x + 1

	// 定义模多项式 x^3 + x + 1
	modPoly := NewGF2mPolynomial([]int{1, 1, 0, 1})

	// 进行多项式乘法并模约简
	result := p.Multiply(q, modPoly)
	fmt.Println("结果多项式的系数:", result.Coefficients)
}
