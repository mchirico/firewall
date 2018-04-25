package set

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestAdd(t *testing.T) {

	iprec := CreateIpRec("100.23.4.20", []int{22, 25, 80})
	iprec2 := CreateIpRec("100.23.4.20", []int{443, 22, 25})
	iprec3 := CreateIpRec("99.2.4.2", []int{443, 22, 25})
	s := CreateS()
	s2 := CreateS()
	s.Add(iprec)
	s.Add(iprec2)
	s2.Add(iprec3)

	s3 := s2.Union(s)

	result := s3.Values()
	expected := []int{22, 25, 80, 443}

	if !reflect.DeepEqual(result["100.23.4.20"], expected) {
		t.Errorf("Expected: %v Observed: %v\n", expected,
			result["100.23.4.20"])
	}
	log.Printf("result: %v\n", s3.Values())

}

func TestAddFromNil(t *testing.T) {

	iprec := CreateIpRec("100.23.4.20", []int{22, 25, 80})

	tmpSet := CreateS()
	tmpSet.Add(iprec)
	s := CreateS()
	s = tmpSet.Difference(s)
	fmt.Printf("values: %v\n", s.Values())
}

func TestDiff(t *testing.T) {
	a := []int{2, 3, 4, 9}
	b := []int{3, 4, 5}
	observed := Diff(a, b)
	expected := []int{2, 9}
	if !reflect.DeepEqual(observed, expected) {
		t.Errorf("Expected: %v Observed: %v\n", expected,
			observed)
	}

}

func TestDiffUnion(t *testing.T) {
	a := []int{2, 3, 4, 9}
	b := []int{3, 4, 5}
	observed := diffUnion(a, b)
	expected := map[int]int{2: 0, 3: 1, 4: 1, 5: -1, 9: 0}
	if !reflect.DeepEqual(observed, expected) {
		t.Errorf("Expected: %v Observed: %v\n", expected,
			observed)
	}
}

func TestSet_Difference(t *testing.T) {
	iprec := CreateIpRec("100.23.4.20", []int{22, 25, 80})
	iprec2 := CreateIpRec("100.23.4.20", []int{443, 22, 25})
	iprecExpected := CreateIpRec("100.23.4.20", []int{80})
	s := CreateS()
	s2 := CreateS()
	s.Add(iprec)
	s2.Add(iprec2)

	expected := CreateS()
	expected.Add(iprecExpected)

	observed := s.Difference(s2)

	fmt.Printf(" %v\n", s.Difference(s2))
	if !reflect.DeepEqual(observed, expected) {
		t.Errorf("Expected: %v Observed: %v\n", expected,
			observed)
	}

}

func TestSet_Copy(t *testing.T) {
	iprec := CreateIpRec("100.23.4.20", []int{22, 25, 80})
	iprec2 := CreateIpRec("100.23.4.20", []int{443, 22, 25})

	s := CreateS()
	s2 := s.Add(iprec).Copy()
	s2.Add(iprec2)

	if reflect.DeepEqual(s, s2) {
		t.Errorf("(Not Equal Test) Expected: %v Observed: %v\n", s,
			s2)
	}

}
