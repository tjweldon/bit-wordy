package util

func Pow(base, power int) int {
	exp := 1
	for i := 0; i < power; i++ {
		exp *= base
	}

	return exp
}

// Rev is a generic slice reverse
func Rev[S ~[]E, E any](s S) S {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}
