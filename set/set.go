package set

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"sync"
)

// diffUnion --
func diffUnion(a []int, b []int) map[int]int {

	m := map[int]int{}

	for _, v := range a {
		m[v] = 0
	}
	for _, v := range b {
		_, found := m[v]
		if found {
			m[v] = 1
		} else {
			m[v] = -1
		}
	}
	return m
}

// Diff -- wrapper to diff
func Diff(a []int, b []int) []int {
	t := []int{}
	result := diffUnion(a, b)
	for k, v := range result {
		if v == 0 {
			t = append(t, k)
		}
	}
	sort.Ints(t)
	return t
}

// USappend -- maybe name change
func USappend(x []int, y []int) []int {

	m := map[int]bool{}
	t := []int{}
	a := []int{}

	a = append(a, x...)
	a = append(a, y...)

	for _, i := range a {
		m[i] = true
	}
	for k := range m {
		t = append(t, k)
	}
	sort.Ints(t)
	return t
}

// SetKV -- add documentation
type SetKV interface {
	Key() string
	Val() []int
}

// IpRec -- making methods off this...
type IpRec struct {
	IP    string
	Count int
	Ports []int
}

// Set
type Set struct {
	sync.Mutex
	set map[string][]int
}

// CreateS -- never lock this
func CreateS() *Set {
	return &Set{set: map[string][]int{}}
}

// Add --
func (s *Set) Add(kv SetKV) *Set {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	setVal, found := s.set[kv.Key()]
	if found {
		val := kv.Val()
		s.set[kv.Key()] = USappend(val, setVal)
	} else {
		s.set[kv.Key()] = kv.Val()
	}
	return s
}

// Union --
func (s *Set) Union(s2 *Set) *Set {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	t := CreateS()
	for k, v := range s2.set {
		t.set[k] = v
	}
	for k, v := range s.set {

		setVal, found := t.set[k]
		if found {
			t.set[k] = USappend(v, setVal)
		} else {
			t.set[k] = v
		}
	}
	return t
}

// Difference -- new set with elements in s but not in s2
func (s *Set) Difference(s2 *Set) *Set {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	t := CreateS()
	for k, v := range s.set {
		t.set[k] = v
	}
	for k, v := range s2.set {
		val, found := t.set[k]
		if found {
			diff := Diff(val, v)
			if len(diff) == 0 {
				delete(t.set, k)
			} else {
				t.set[k] = diff
			}

		}
	}
	return t
}

// In -- only compares key
func (set *Set) In(s string) bool {
	set.Mutex.Lock()
	defer set.Mutex.Unlock()
	_, found := set.set[s]

	return found
}

// Values --
func (s *Set) Values() map[string][]int {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	// Want Copy
	t := map[string][]int{}
	for k, v := range s.set {
		t[k] = v
	}
	return t
}

// Copy -- copy and not reference
func (s *Set) Copy() *Set {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	t := &Set{set: map[string][]int{}}
	for k, v := range s.set {
		t.set[k] = v
	}
	return t
}

// Clear -- empties set
func (s *Set) Clear() *Set {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	t := CreateS()
	s.set = t.set
	return s
}

// WriteToFile --
func (s *Set) WriteToFile(file string) *Set {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	os.Remove(file)
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("writeRecs OpenFile error %v", err)
	}
	defer f.Close()

	setJson, _ := json.Marshal(s.set)
	f.Write(setJson)

	return s
}

// LoadFromFile -- appends values
func (s *Set) LoadFromFile(file string) *Set {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("writeRecs OpenFile error %v", err)
		return s
	}

	err = json.Unmarshal(f, &s.set)
	if err != nil {
		log.Printf("\nLoadFromFile: %v\n", err)
	}

	for k, v := range s.set {
		log.Printf("values: %v:%v", k, v)
	}

	return s
}

// CreateIpRec --
func CreateIpRec(ip string, ports []int) *IpRec {
	sort.Ints(ports)
	t := &IpRec{IP: ip, Count: 0, Ports: ports}
	return t
}

func (iprec *IpRec) Key() string {
	return iprec.IP
}

func (iprec *IpRec) Val() []int {
	return append(iprec.Ports)
}
