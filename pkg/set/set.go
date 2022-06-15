package set

import "strings"

type Set struct {
	m map[string]struct{}
}

func New() *Set {
	return &Set{
		m: make(map[string]struct{}),
	}
}

func (s *Set) Add(v string) {
	s.m[v] = struct{}{}
}

func (s *Set) Has(v string) bool {
	_, ok := s.m[v]
	return ok
}

func (s *Set) String() string {
	str := []string{}
	for k := range s.m {
		str = append(str, k)
	}
	return strings.Join(str, ",")
}
