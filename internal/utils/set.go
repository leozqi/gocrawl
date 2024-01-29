// Copyright (C) 2024 Leo Qi <leo@leozqi.com>
//
// Use of this source code is governed by the Apache-2.0 License.
// Full text can be found in the LICENSE file

package utils

type Set struct {
    list map[string]struct{} // empty structs occupy 0 memory
}

func (s *Set) Has(v string) bool {
    _, ok := s.list[v]
    return ok
}

func (s *Set) Add(v string) {
    s.list[v] = struct{}{}
}

func (s *Set) Remove(v string) {
    delete(s.list, v)
}

func (s *Set) Clear() {
    s.list = make(map[string]struct{})
}

func (s *Set) Size() int {
    return len(s.list)
}

func NewSet() *Set {
    s := &Set{}
    s.list = make(map[string]struct{})
    return s
}

//optional functionalities

// AddMulti Add multiple values in the set
func (s *Set) AddMulti(list ...string) {
    for _, v := range list {
        s.Add(v)
    }
}

func (s *Set) Union(s2 *Set) *Set {
    res := NewSet()
    for v := range s.list {
        res.Add(v)
    }

    for v := range s2.list {
        res.Add(v)
    }
    return res
}

func (s *Set) Intersect(s2 *Set) *Set {
    res := NewSet()
    for v := range s.list {
        if s2.Has(v) == false {
            continue
        }
        res.Add(v)
    }
    return res
}

// Difference returns the subset from s, that doesn't exists in s2 (param)
func (s *Set) Difference(s2 *Set) *Set {
    res := NewSet()
    for v := range s.list {
        if s2.Has(v) {
            continue
        }
        res.Add(v)
    }
    return res
}

func (s *Set) Slice() []string {
    v := make([]string, 0, len(s.list))
    for value, _ := range s.list {
        v = append(v, value)
    }
    return v
}

