package primitives

import "sort"

// By is the type of a "less" function that defines the ordering of its Result arguments.
type By func(p1, p2 *Result) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(rows []Result) {
	ps := &ResultSorter{
		rows: rows,
		by:   by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// ResultSorter joins a By function and a slice of results to be sorted.
type ResultSorter struct {
	rows []Result
	by   func(p1, p2 *Result) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *ResultSorter) Len() int {
	return len(s.rows)
}

// Swap is part of sort.Interface.
func (s *ResultSorter) Swap(i, j int) {
	s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *ResultSorter) Less(i, j int) bool {
	return s.by(&s.rows[i], &s.rows[j])
}
