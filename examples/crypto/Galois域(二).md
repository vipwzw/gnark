# galois域(二)

## 计算机领域为啥需要 $2^n$ 的域

我们知道，计算机中的运算单位都是 $2^n$, 比如 char 类型的运算 就是 $2^8$ 范围的一个运算。普通的加减乘除运算规则会让运算范围超出char 类型能表达的范围，这个就是溢出。能不能设计一种运算规则，可以在char类型的范围内自由的做加减乘除，这篇文章要介绍这样的一个运算规则。

## $Z_2$ 是一个域

我们知道，$Z_p$ 是一个域, 这个在上一节中有证明。而且这个结论和拉格朗日定理有非常深刻的联系。正常情况下，只有素数的情况下才会构成域，而非常不幸的事情是，$2^n$ 在 $n>1$ 的情况下都不构成素数，所以，需要新的思路，让他变成域。所以，不能从一维的数字来考虑域的问题，要把这个问题变成二维的，三维的，n维的问题， 那么 $2^n$ 变成了一个扩域。我们最熟悉的一个扩域就是复数域。复数是一个二维的数字，所有的数字构成都是 $a+b*i$ 的形式，如果你从多项式的角度来看，这个是一个一次的多项式。依次类推，二次的多项式有三个维度， $a+b * i + b * i^2$ , 当然，这里会有一个问题，我们发现，复数相乘的时候会出现超过 一次的 $i^2$, 在复数中，我们通过定义 $i^2 = -1$ 来解决这个问题，这样复数乘积会变成封闭，对二次以上的多项式，一般用取模的概念来解决这个问题。

## $Z_{2^n}$ 域的运算如何定义

$Z_{2^n}$ 域实际上就是系数在 $Z_2$ 域上的多项式，$Z_2$ 域就是加减乘除都要mod 2 的域，我们可以总结一下这个域的运算规则:

在模为2的域（即 $Z_2$ 或 GF(2)）中，所有的运算都在模2的基础上进行。这意味着所有的加法、减法、乘法和除法运算的结果都要对2取模。以下是 $Z_2$ 域的运算规则：

### 加法和减法
在 $Z_2$ 域中，加法和减法是相同的，因为 $1 + 1 \equiv 0 \pmod{2}$。具体规则如下：
- $0 + 0 = 0$
- $0 + 1 = 1$
- $1 + 0 = 1$
- $1 + 1 = 0$

- $0 - 0 = 0$
- $0 - 1 = 1$
- $1 - 0 = 1$
- $1 - 1 = 0$
  
### 乘法
乘法规则如下：
- $0 \times 0 = 0$
- $0 \times 1 = 0$
- $1 \times 0 = 0$
- $1 \times 1 = 1$

### 除法
在 $Z_2$ 域中，除法与乘法是相同的，因为 $1$ 是其自身的逆元。具体规则如下：
- $0 / 1 = 0$
- $1 / 1 = 1$

### 例子
假设我们有两个元素 $a$ 和 $b$，它们的值可以是 $0$ 或 $1$。以下是一些运算示例：
- 加法：$a + b \equiv (a + b) \pmod{2}$
- 乘法：$a \times b \equiv (a \times b) \pmod{2}$

假设我们有两个多项式 $P(x) = x^2 + x + 1$ 和 $Q(x) = x + 1$，它们的系数在 $Z_2$ 域上。以下是一些运算示例：
- 多项式加法：$P(x) + Q(x) = (x^2 + x + 1) + (x + 1) = x^2 + 2x + 2 \equiv x^2 \pmod{2}$
- 多项式乘法：$P(x) \times Q(x) = (x^2 + x + 1) \times (x + 1) = x^3 + x^2 + x^2 + x + x + 1 = x^3 + 2x^2 + 2x + 1 \equiv x^3 + 1 \pmod{2}$

通过这些运算规则，我们可以在 $Z_2$ 域上进行各种代数运算，并将其扩展到 $Z_{2^n}$ 域。这里我们发现，乘法之后，出现了三次多项式，这样运算就不封闭了，构成不了域，下面就要引入本源多项式解决这个问题。

## 本源多项式

在 Galois 域 $GF(2^n)$ 中，运算规则不仅包括基本的加法和乘法，还涉及到使用本源多项式（irreducible polynomial）进行模运算。本源多项式是一个不可约多项式，用于定义扩展域 $GF(2^n)$ 的元素表示和运算规则。

### 运算规则

在 $GF(2^n)$ 域中，多项式的系数在 GF(2) 域上，运算规则如下：

- **多项式加法**：系数对应相加，并对2取模。
- **多项式乘法**：系数对应相乘，并对2取模。
- **多项式除法**：使用多项式的逆元进行运算，并对2取模。
- **模运算**：使用本源多项式对结果进行模运算，确保结果在 $GF(2^n)$ 域内。

### 什么是本源多项式

本源多项式是一个不可约多项式，用于定义 GF(2^n) 域。例如，对于 $GF(2^3)$，一个常用的本源多项式是 $x^3 + x + 1$。这个类似整数域中的素数。

### 带本源多项式运算中的例子

假设我们有两个多项式 $P(x) = x^2 + x + 1$ 和 $Q(x) = x + 1$，它们的系数在 GF(2) 域上，并且使用本源多项式 $M(x) = x^3 + x + 1$ 进行模运算。以下是一些运算示例：

#### 多项式加法

$$
P(x) + Q(x) = (x^2 + x + 1) + (x + 1) = x^2 + 2x + 2 \equiv x^2 \pmod{2}
$$

#### 多项式乘法

$$
P(x) \times Q(x) = (x^2 + x + 1) \times (x + 1) = x^3 + x^2 + x^2 + x + x + 1 = x^3 + 2x^2 + 2x + 1 \equiv x^3 + 1 \pmod{2}
$$

由于 $x^3 \equiv x + 1 \pmod{M(x)}$，所以：

$$
x^3 + 1 \equiv (x + 1) + 1 = x \pmod{M(x)}
$$

通过这些运算规则和本源多项式的使用，我们可以在 $GF(2^n)$ 域中进行各种代数运算，并确保结果在该域内。

定义了一种可以在 $2^n$ 的范围内可以封闭运算的规则后，接下来我们要看 $GF(2^n)$ 域在实践中的运用，主要介绍两个领域，一个领域是编码的领域，另外一个领域是密码学的领域。

# 如何选择适合的本源多项式来构建 GF(2^n) 域？

选择适合的本源多项式来构建 GF(2^n) 域是一个关键步骤，因为本源多项式定义了域的结构和运算规则。以下是选择本源多项式的一些指导原则和步骤：

### 1. 不可约性
本源多项式必须是不可约的，这意味着它不能被分解为更低次多项式的乘积。在 GF(2) 上，不可约多项式的判定可以通过试除法来实现。

### 2. 多项式的次数
本源多项式的次数必须等于 n，因为我们要构建 GF(2^n) 域。例如，要构建 GF(2^3) 域，本源多项式的次数必须为 3。

### 3. 选择常用的本源多项式
在实践中，通常会选择一些已知的、经过验证的本源多项式。这些多项式已经被广泛使用，并且在各种应用中表现良好。以下是一些常用的本源多项式：

- GF(2^2): $x^2 + x + 1$
- GF(2^3): $x^3 + x + 1$
- GF(2^4): $x^4 + x + 1$
- GF(2^5): $x^5 + x^2 + 1$
- GF(2^6): $x^6 + x + 1$
- GF(2^7): $x^7 + x^3 + 1$
- GF(2^8): $x^8 + x^4 + x^3 + x + 1$

在 GF(2^n) 域中，判断一个多项式是否为不可约多项式（即本源多项式）是一个关键步骤。不可约多项式是不能被分解为更低次多项式的乘积的多项式。以下是判断不可约多项式的详细步骤和算法描述。

### 判断不可约多项式的步骤

1. **输入多项式**：输入一个多项式 $P(x)$，其系数在 GF(2) 上。
2. **检查次数**：检查多项式的次数是否大于1。如果次数为1，则多项式是不可约的。
3. **生成所有可能的因子**：生成所有可能的低于 $P(x)$ 次数的一元多项式。
4. **多项式除法**：对每个可能的因子进行多项式除法，检查是否存在因子能够整除 $P(x)$。
5. **判断结果**：如果存在因子能够整除 $P(x)$，则 $P(x)$ 是可约的；否则，$P(x)$ 是不可约的。

### 详细算法描述

#### 1. 输入多项式
输入一个多项式 $P(x)$，其系数在 GF(2) 上。例如，$P(x) = x^3 + x + 1$。

#### 2. 检查次数
检查多项式的次数是否大于1。如果次数为1，则多项式是不可约的。

#### 3. 生成所有可能的因子
生成所有可能的低于 $P(x)$ 次数的一元多项式。例如，对于 $P(x) = x^3 + x + 1$，生成次数为1和2的所有多项式：
- $x$
- $x + 1$
- $x^2$
- $x^2 + 1$
- $x^2 + x$
- $x^2 + x + 1$

#### 4. 多项式除法
对每个可能的因子进行多项式除法，检查是否存在因子能够整除 $P(x)$。具体步骤如下：
- 对于每个因子 $Q(x)$，进行多项式除法 $P(x) \div Q(x)$。
- 如果余数为0，则 $Q(x)$ 是 $P(x)$ 的因子。

#### 5. 判断结果
如果存在因子能够整除 $P(x)$，则 $P(x)$ 是可约的；否则，$P(x)$ 是不可约的。

### 代码实现
以下是用 Go 语言实现的代码示例，用于判断一个多项式是否为不可约多项式：

```go
package main

import (
    "fmt"
)

// 多项式表示
type Polynomial struct {
    Coefficients []int
}

// 多项式加法
func (p Polynomial) Add(q Polynomial) Polynomial {
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

    return Polynomial{Coefficients: resultCoefficients}
}

// 多项式乘法
func (p Polynomial) Multiply(q Polynomial) Polynomial {
    m := len(p.Coefficients)
    n := len(q.Coefficients)
    resultCoefficients := make([]int, m+n-1)

    for i := 0; i < m; i++ {
        for j := 0; j < n; j++ {
            resultCoefficients[i+j] ^= p.Coefficients[i] * q.Coefficients[j]
        }
    }

    return Polynomial{Coefficients: resultCoefficients}
}

// 多项式除法
func (p Polynomial) Divide(q Polynomial) (Polynomial, Polynomial) {
    m := len(p.Coefficients)
    n := len(q.Coefficients)
    if m < n {
        return Polynomial{Coefficients: []int{0}}, p
    }

    resultCoefficients := make([]int, m-n+1)
    remainder := make([]int, m)
    copy(remainder, p.Coefficients)

    for i := 0; i <= m-n; i++ {
        if remainder[i] == 1 {
            resultCoefficients[i] = 1
            for j := 0; j < n; j++ {
                remainder[i+j] ^= q.Coefficients[j]
            }
        }
    }

    return Polynomial{Coefficients: resultCoefficients}, Polynomial{Coefficients: remainder}
}

// 生成指定次数的多项式
func generatePolynomials(degree int) []Polynomial {
    if degree == 0 {
        return []Polynomial{{Coefficients: []int{0}}, {Coefficients: []int{1}}}
    }
    smallerPolys := generatePolynomials(degree - 1)
    var polys []Polynomial
    for _, poly := range smallerPolys {
        polys = append(polys, Polynomial{Coefficients: append([]int{0}, poly.Coefficients...)})
        polys = append(polys, Polynomial{Coefficients: append([]int{1}, poly.Coefficients...)})
    }
    return polys
}

// 验证多项式是否为不可约多项式
func isIrreducible(poly Polynomial) bool {
    n := len(poly.Coefficients) - 1
    for i := 1; i <= n/2; i++ {
        for _, divisor := range generatePolynomials(i) {
            _, remainder := poly.Divide(divisor)
            if len(remainder.Coefficients) == 0 || remainder.Coefficients[0] == 0 {
                return false
            }
        }
    }
    return true
}

func main() {
    // 定义一个多项式 x^3 + x + 1
    poly := Polynomial{Coefficients: []int{1, 0, 1, 1}}

    // 验证多项式是否为不可约多项式
    if isIrreducible(poly) {
        fmt.Println("多项式是不可约的")
    } else {
        fmt.Println("多项式不是不可约的")
    }
}
```

### 解释
1. **多项式表示**：`Polynomial` 结构体表示一个多项式，包含一个系数切片。
2. **多项式加法和乘法**：实现多项式的加法和乘法运算。
3. **多项式除法**：实现多项式的除法运算，返回商和余数。
4. **生成多项式**：`generatePolynomials` 函数生成指定次数的所有多项式。
5. **验证不可约性**：`isIrreducible` 函数验证多项式是否为不可约多项式。

通过这些步骤和代码实现，可以判断一个多项式是否为不可约多项式，从而选择适合的本源多项式来构建 GF(2^n) 域。

### 结论
选择适合的本源多项式来构建 GF(2^n) 域需要确保多项式是不可约的，并且其次数等于 n。可以选择一些常用的本源多项式，并通过计算验证其不可约性和适用性。通过这些步骤，可以确保构建的 GF(2^n) 域具有正确的结构和运算规则。我们选择的验证方法算法效率并不是很高，对于 $2^64$ 规模大小的域，我们需要 $2^32$ 次运算。这个基本上还能接受，但是对于更大的区间，这个运算量就无法接受了。不过幸运的是，一般来说， $2^64$ 这个规模的域已经非常够用了。

## GF(2^n) 域在密码学中有什么应用

GF(2^n) 域在密码学中有广泛的应用，主要因为其数学性质和高效的运算特性。以下是一些主要的应用领域：

### 1. 对称加密算法

在对称加密算法中，GF(2^n) 域被广泛用于构建 S-盒（Substitution box）和其他非线性变换。例如：
- **AES（高级加密标准）**：AES 使用 GF(2^8) 域来构建其 S-盒。具体来说，AES 的 S-盒是通过 GF(2^8) 域上的逆元运算和仿射变换生成的。
- **Rijndael**：Rijndael 是 AES 的前身，也使用 GF(2^8) 域来构建其 S-盒。

### 2. 纠错码
在纠错码中，GF(2^n) 域用于构建高效的编码和解码算法。例如：
- **Reed-Solomon 码**：Reed-Solomon 码使用 GF(2^n) 域来构建编码和解码算法，广泛应用于数据存储和传输中的错误检测和纠正。
- **BCH 码**：BCH 码也是一种基于 GF(2^n) 域的纠错码，广泛应用于通信系统中。

### 3. 哈希函数
在某些哈希函数中，GF(2^n) 域用于构建高效的哈希算法。例如：
- **SHA-3**：SHA-3 使用 GF(2^n) 域来构建其内部的置换和混合操作。

### 4. 零知识证明
在零知识证明系统中，GF(2^n) 域用于构建高效的证明和验证算法。例如：
- **zk-SNARKs**：zk-SNARKs 使用 GF(2^n) 域来构建高效的证明和验证算法，广泛应用于区块链和隐私保护中。

在密码学中， GF(2^n) 的使用大同小异，我们主要描述他在零知识证明中的应用。

零知识证明（Zero-Knowledge Proof，简称 ZKP）是一种密码学技术，允许证明者向验证者证明某个陈述为真，而不泄露任何关于该陈述的额外信息。GF(2^n) 域在零知识证明中有广泛的应用，主要因为其高效的运算特性和良好的数学性质。以下是 GF(2^n) 域在零知识证明中的详细应用描述。

### GF(2^n) 域在零知识证明中的应用

#### 1. zk-SNARKs（零知识简洁非交互式知识论证）

zk-SNARKs 是一种高效的零知识证明系统，广泛应用于区块链和隐私保护中。GF(2^n) 域在 zk-SNARKs 中用于构建高效的证明和验证算法。具体应用如下：

- **多项式承诺**：zk-SNARKs 使用 GF(2^n) 域来构建多项式承诺方案。多项式承诺允许证明者承诺一个多项式，并在后续的交互中高效地证明多项式的某些性质。
- **椭圆曲线运算**：zk-SNARKs 使用 GF(2^n) 域来定义椭圆曲线，并在其上进行高效的点乘运算。椭圆曲线的选择和 GF(2^n) 域的运算特性使得 zk-SNARKs 具有高效性和安全性。

#### 2. zk-STARKs（零知识可扩展透明知识论证）

zk-STARKs 是另一种零知识证明系统，强调透明性和可扩展性。GF(2^n) 域在 zk-STARKs 中用于构建高效的证明和验证算法。具体应用如下：

- **多项式分解**：zk-STARKs 使用 GF(2^n) 域来进行多项式分解和插值运算。GF(2^n) 域的选择使得这些运算在大规模数据上仍然高效。
- **FFT（快速傅里叶变换）**：zk-STARKs 使用 GF(2^n) 域来实现快速傅里叶变换，用于高效地处理多项式和离散傅里叶变换。

#### 3. Bulletproofs

Bulletproofs 是一种高效的零知识证明系统，主要用于证明范围证明和其他加密货币相关的证明。GF(2^n) 域在 Bulletproofs 中用于构建高效的证明和验证算法。具体应用如下：

- **内积证明**：Bulletproofs 使用 GF(2^n) 域来构建内积证明方案。内积证明允许证明者高效地证明两个向量的内积等于某个值，而不泄露向量的具体值。
- **椭圆曲线运算**：Bulletproofs 使用 GF(2^n) 域来定义椭圆曲线，并在其上进行高效的点乘运算。GF(2^n) 域的选择使得这些运算具有高效性和安全性。

### 结论

GF(2^n) 域在零知识证明中有广泛的应用，主要因为其高效的运算特性和良好的数学性质。通过使用 GF(2^n) 域，可以构建高效、安全的零知识证明系统，如 zk-SNARKs、zk-STARKs 和 Bulletproofs。这些系统在区块链、隐私保护和加密货币等领域中起着至关重要的作用。