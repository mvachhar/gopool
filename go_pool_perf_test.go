package gopool

import "testing"

func doIt(i int, ch chan int) {
	ch <- i
}

func eat(length int, ch chan int) (doneCh chan int) {
	doneCh = make(chan int)
	go func() {
		for i := 0; i < length; i++ {
			<-ch
		}
		doneCh <- length
	}()
	return doneCh
}

func goRoutineBench(gp GoPool) {
	length := 100000
	ch := make(chan int)

	doneCh := eat(length, ch)

	for i := 0; i < length; i++ {
		gp.Go(func() { doIt(i, ch) })
	}

	<-doneCh
}

func parallelBench(b *testing.B, poolSize int) {
	gp, err := New(poolSize)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		goRoutineBench(gp)
	}
	b.Logf("Requests %d, Misses %d, Miss Rate %g",
		gp.Requests(),
		gp.Misses(),
		float64(gp.Misses())/float64(gp.Requests()))
}

func BenchmarkPlainGoParallel(b *testing.B) {
	gp, err := NewNoPool()
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		goRoutineBench(gp)
	}
}

func BenchmarkGoPoolParallel10(b *testing.B) {
	parallelBench(b, 10)
}

func BenchmarkGoPoolParallel100(b *testing.B) {
	parallelBench(b, 100)
}

func BenchmarkGoPoolParallel1000(b *testing.B) {
	parallelBench(b, 1000)
}

func BenchmarkGoPoolParallel10000(b *testing.B) {
	parallelBench(b, 10000)
}

func BenchmarkGoPoolParallel100000(b *testing.B) {
	parallelBench(b, 100000)
}
