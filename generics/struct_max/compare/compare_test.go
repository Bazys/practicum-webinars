package compare

import "testing"

type Comparator[T any] interface {
	Compare(a, b T) int
}

type IntComparator struct{}

func (ic IntComparator) Compare(a, b int) int {
	if a == b {
		return 0
	} else if a < b {
		return -1
	} else {
		return 1
	}
}

type StringLengthComparator struct{}

func (slc StringLengthComparator) Compare(a, b string) int {
	lenA := len(a)
	lenB := len(b)
	if lenA == lenB {
		return 0
	} else if lenA < lenB {
		return -1
	} else {
		return 1
	}
}

func CompareValues[T any](comp Comparator[T], a, b T) int {
	return comp.Compare(a, b)
}

func TestCompareValues(t *testing.T) {
	intComp := IntComparator{}
	result := CompareValues(intComp, 5, 10)
	if result != -1 {
		t.Errorf("Expected -1, got %v", result)
	}

	strComp := StringLengthComparator{}
	result = CompareValues(strComp, "short", "longer")
	if result != -1 {
		t.Errorf("Expected -1, got %v", result)
	}
	result = CompareValues(strComp, "short", "start")
	if result != 0 {
		t.Errorf("Expected -1, got %v", result)
	}
}
