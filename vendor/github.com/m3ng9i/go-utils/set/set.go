// A simple set data structure.
package set

import "sync"
import "fmt"
import "bytes"

type any interface{}

type Set struct {
    m map[any]struct{}
    sync.RWMutex
}


// Determine if a variable could be a member of Set.
func IsLegal(v any) bool {
    switch v.(type) {
        // byte is alias of uint8, rune is alias of uint32,
        // so byte and rune are not in case clause
        case bool, error, string, complex64, complex128, float32, float64,
             int, int8, int16, int32, int64,
             uint, uint8, uint16, uint32, uint64, uintptr:
                return true
    }

    return false
}


// Create a new Set
func New() *Set {
    s := &Set{}
    s.m = make(map[any]struct{})
    return s
}


// Create a new Set and add some items.
func NewSet(items ...any) (s *Set, ok bool) {
    s = New()
    ok = s.Add(items...)
    return
}


// Create a new Set and add some items. Panic if not success.
func MustNew(items ...any) *Set {
    s, ok := NewSet(items...)
    if !ok {
        panic(fmt.Errorf("Create Set failed"))
    }
    return s
}


// Add item(s) to set.
// If there's an item that is not a legal type, return false.
// If success, return true.
func (s *Set) Add(items ...any) bool {
    s.Lock()
    defer s.Unlock()

    // If any item of items is not legal, return false
    for _, i := range items {
        if IsLegal(i) == false {
            return false
        }
    }

    for _, i := range items {
        // struct{}{} is a struct{} instance
        s.m[i] = struct{}{}
    }

    return true
}


// Add item(s) to set. if there's an item that is not a legal type, panic
func (s *Set) MustAdd(items ...any) {
    if s.Add(items...) == false {
        value := fmt.Sprintf("%v", items)

        // remove the outside square brackets
        if value[0] == '[' && value[len(value)-1] == ']' {
            value = value[1:len(value)-1]
        }
        panic(fmt.Sprintf("Value is not legal for adding to Set: %v\n", value))
    }
}


func (s *Set) Remove(items ...any) {
    s.Lock()
    s.Unlock()

    for _, i := range items {
        delete(s.m, i)
    }
}


func (s *Set) Has(item any) bool {
    s.RLock()
    defer s.RUnlock()
    _, ok := s.m[item]
    return ok
}


func (s *Set) Len() int {
    s.RLock()
    defer s.RUnlock()
    return len(s.m)
}


func (s *Set) Clear() {
    s.Lock()
    defer s.Unlock()
    s.m = make(map[any]struct{})
}


func (s *Set) IsEmpty() bool {
    return s.Len() == 0
}


func (s *Set) List() []any {
    s.RLock()
    defer s.RUnlock()

    l := make([]any, 0, s.Len())

    for i := range s.m {
        l = append(l, i)
    }
    return l
}


func (s *Set) String() string {
    s.RLock()
    defer s.RUnlock()

    var buf bytes.Buffer
    buf.WriteString("Set{")

    first := true
    for i := range s.m {
        if first {
            first = false
        } else {
            buf.WriteString(", ")
        }
        buf.WriteString(fmt.Sprintf("%v", i))
    }
    buf.WriteString("}")
    return buf.String()
}


func (s *Set) Clone() *Set {
    s.RLock()
    defer s.RUnlock()

    n, _ := NewSet(s.List()...)
    return n
}


func Equals(s1, s2 *Set) bool {

    if s1 == nil && s2 == nil {
        return true
    }

    if s1 == nil || s2 == nil {
        return false
    }

    s1.RLock()
    defer s1.RUnlock()
    s2.RLock()
    defer s2.RUnlock()

    if s1.Len() != s2.Len() {
        return false
    }

    for key := range s1.m {
        if !s2.Has(key) {
            return false
        }
    }

    return true
}


// Determine if Set s is superset of Set a
func IsSuperset(s1, s2 *Set) bool {

    if s1 == nil || s2 == nil {
        return false
    }

    s1.RLock()
    defer s1.RUnlock()
    s2.RLock()
    defer s2.RUnlock()

    s1Len := s1.Len()
    s2Len := s2.Len()

    if s1Len == 0 || s1Len == s2Len {
        return false
    }

    if s1Len > 0 && s2Len == 0 {
        return true
    }

    for _, v := range s2.List() {
        if !s1.Has(v) {
            return false
        }
    }

    return true
}


func Union(s ...*Set) *Set {

    n := New()

    for _, i := range s {
        if i == nil {
            continue
        }

        i.RLock()
        defer i.RUnlock()

        n.Add(i.List()...)
    }

    return n
}


func Intersect(s ...*Set) *Set {

    if len(s) == 0 {
        return New()
    }

    n := s[0]

    for _, i:= range s[1:] {
        if i == nil {
            return nil
        }

        i.RLock()
        defer i.RUnlock()

        t := New()
        for _, v := range i.List() {
            if n.Has(v) {
                t.Add(v)
            }
        }
        n = t
    }

    return n
}
