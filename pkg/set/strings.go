package set

import "sort"

type Strings map[string]struct{}

func (s Strings) Add(values ...string) {
	for _, v := range values {
		s[v] = struct{}{}
	}
}

func (s Strings) Remove(values ...string) {
	for _, v := range values {
		delete(s, v)
	}
}

func (s Strings) Has(v string) bool {
	_, ok := s[v]
	return ok
}

func (s Strings) Slice() []string {
	values := make([]string, 0, len(s))
	for v := range s {
		values = append(values, v)
	}
	sort.Strings(values)
	return values
}

func (s Strings) Intersect(t Strings) Strings {
	intersect := make(Strings)
	for _, sv := range s.Slice() {
		if _, ok := t[sv]; ok {
			intersect[sv] = struct{}{}
		}
	}
	return intersect
}

func (s Strings) Union(t Strings) Strings {
	union := make(Strings)
	union.Add(s.Slice()...)
	union.Add(t.Slice()...)
	return union
}

func (s Strings) Subtract(t Strings) Strings {
	subtract := make(Strings)
	subtract.Add(s.Slice()...)
	subtract.Remove(t.Slice()...)
	return subtract
}
