package main

import "testing"

func BenchmarkSolveOne(b *testing.B) {
	for i := 0; i < b.N; i++ {
		solveOne()
	}
}
