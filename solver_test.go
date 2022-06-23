package main

import (
	"testing"
)

// func BenchmarkSolveOne(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		solveOne()
// 	}
// }

func BenchmarkIterate_Run(b *testing.B) {
	iter := Iterate{b.N, false}
	err := iter.Run(nil)
	if err != nil {
		b.Errorf("%s", err)
	}
}
