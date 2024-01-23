package sets

import (
	"fmt"
	"maps"
	"strings"
)

type Set[T comparable] struct {
	elements map[T]struct{}
}

func allocate[T comparable](size int) Set[T] {
	return Set[T]{
		elements: make(map[T]struct{}, size),
	}
}

func New[T comparable](elts ...T) Set[T] {
	res := allocate[T](len(elts))
	for _, elt := range elts {
		res.Add(elt)
	}

	return res
}

func Empty[T comparable]() Set[T] {
	return New[T]()
}

func (s *Set[T]) Size() int {
	return len(s.elements)
}

func (s *Set[T]) Add(elt T) {
	s.elements[elt] = struct{}{}
}

func (s *Set[T]) Remove(elt T) {
	delete(s.elements, elt)
}

func (s *Set[T]) Clear() {
	*s = Empty[T]()
}

func (s *Set[T]) Equals(s2 *Set[T]) bool {
	return maps.Equal(s.elements, s2.elements)
}

func (s *Set[T]) ForEach(f func(t T)) {
	for elt := range s.elements {
		f(elt)
	}
}

func (s *Set[T]) Contains(elt T) bool {
	_, exists := s.elements[elt]
	return exists
}

func (s *Set[T]) Copy() Set[T] {
	res := allocate[T](s.Size())
	s.ForEach(func(elt T) {
		res.Add(elt)
	})

	return res
}

func (s Set[T]) String() string {
	b := strings.Builder{}
	b.WriteString("{")

	count := 0
	size := s.Size()
	for elt := range s.elements {
		count += 1
		b.WriteString(fmt.Sprint(elt))
		if count < size {
			b.WriteString(", ")
		}
	}

	b.WriteString("}")
	return b.String()
}
