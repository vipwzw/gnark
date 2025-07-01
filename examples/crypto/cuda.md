# 在cpu中模拟cuda并行编程

cuda本质是一个多核计算的框架。通过软件和硬件的结合，让并行计算变的非常的简单。但是并行计算的模式和传统计算还是有一些区别，理解这些规则的最好方法是，用我们熟悉的cpu编程来实现cuda的一些基本功能。这样我们就能深入的理解cuda的模式。

## cuda 矢量加法

我们用cuda实现两个矢量加法:

```c++
__global__ void add( int *a, int *b, int *c ) {
    int tid = blockIdx.x;    // this thread handles the data at its thread id
    if (tid < N)
        c[tid] = a[tid] + b[tid];
}
```

调用:

```C++
add<<<N,1>>>( dev_a, dev_b, dev_c );
```

我们发现，矢量加法，我们不需要做for循环，gpu自动会调度，完成所有的加法。

下面我们用go语言实现一个类似的功能：

```go
package main

import (
    "fmt"
    "sync"
)

const N = 10

func add(a, b, c []int, wg *sync.WaitGroup, id int) {
    defer wg.Done()
    if id < N {
        c[id] = a[id] + b[id]
    }
}

func main() {
    a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    b := []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
    c := make([]int, N)

    var wg sync.WaitGroup

    for i := 0; i < N; i++ {
        wg.Add(1)
        go add(a, b, c, &wg, i)
    }

    wg.Wait()

    fmt.Println("Result:", c)
}
```

## cuda 任意长度矢量加法

当然，上面的代码没有考虑性能问题，当N非常大的时候，新开go线程的性能要远远大于计算的开销，当然，cuda中也有这样的问题，
线程虽然非常轻量级，但是也不是无限的。我们看一下，当N比较大的时候，一般怎么处理呢？cuda抽象了线程块 和 线程两个概念。
线程块里面会装固定数量的线程，一个线程块中，线程数量一般不能超过1024。对于任意数量的N，我们可以计算，如果一个线程块有
128个线程，那么我们需要 N/128 个线程块，但是，直接除有点问题 1-127 的情况下，这个数字为0，我们改为

```c++
gridDim.x = (N+127)/128
```

这样，我们可以认为，一共有 128 * gridDim.x 核心在运行。

```c++
__global__ void add( int *a, int *b, int *c ) {
    int tid = threadIdx.x + blockIdx.x * blockDim.x;
    while (tid < N) {
        c[tid] = a[tid] + b[tid];
        tid += blockDim.x *  gridDim.x;
    }
}
```

Go中，我们没有线程和线程块的概念，我们就是每次用M个协程来做运算。
下面是代码

```go
package main

import (
    "fmt"
    "sync"
)

const (
    N = 1000  // 向量长度
    M = 128   // 协程数量
)

func add(a, b, c []int, wg *sync.WaitGroup, id int) {
    defer wg.Done()
    tid := id
    for tid < N {
        c[tid] = a[tid] + b[tid]
        tid += M
    }
}

func main() {
    a := make([]int, N)
    b := make([]int, N)
    c := make([]int, N)

    // 初始化向量 a 和 b
    for i := 0; i < N; i++ {
        a[i] = i
        b[i] = N - i
    }

    var wg sync.WaitGroup

    for i := 0; i < M; i++ {
        wg.Add(1)
        go add(a, b, c, &wg, i)
    }

    wg.Wait()

    fmt.Println("Result:", c)
}
```

## cuda 点积算法

到这里，只是启动多个线程去做运算，当然，这些运算必须要有同步，要等所有线程运算完毕后，才能最终输出结果。
到现在为止，我们每个线程做的事情是一样的，但是有些时候，不一样才能发挥更大的作用，我们来看一个更加复杂的问题。
我们来看一个向量的点积问题。

```c++
__global__ void dot( float *a, float *b, float *c ) {
    __shared__ float cache[threadsPerBlock];
    int tid = threadIdx.x + blockIdx.x * blockDim.x;
    int cacheIndex = threadIdx.x;

    float   temp = 0;
    while (tid < N) {
        temp += a[tid] * b[tid];
        tid += blockDim.x * gridDim.x;
    }
    
    // set the cache values
    cache[cacheIndex] = temp;
    
    // synchronize threads in this block
    __syncthreads();

    // for reductions, threadsPerBlock must be a power of 2
    // because of the following code
    int i = blockDim.x/2;
    while (i != 0) {
        if (cacheIndex < i)
            cache[cacheIndex] += cache[cacheIndex + i];
        __syncthreads();
        i /= 2;
    }

    if (cacheIndex == 0)
        c[blockIdx.x] = cache[0];
}
```

上面的代码稍微复杂了一些，我们会根据 cacheIndex 的不同，产生不同的行为。而且，这里还有 __syncthreads() 这样的函数，意味着，我们要等待其他线程的完成。虽然上面的代码比较复杂，不熟悉cuda编程模式的人甚至看不懂，他做的事情实际上非常简单，就是求两个向量的点积。他这个求和过程，复杂度从 N 变成了 LogN。当然，cuda的同步足够轻量级别，要是在go中，他的同步过于重量级不能提高性能，除非进行非常复杂的运算，我们可以用这样的编程模式。不过为了便于理解，我们实现Go版本的模仿cuda的点积功能。

```go

```