package router

import (
	"testing"
	"fmt"
)

func TestAddAndContainsWithManyItems(t *testing.T) {
	test := Test{t}
	upperBound := 100 * 1000

	set := Set{}
	for i := 0; i < upperBound; i++ {
		if i % 2 == 0 {
			set.Add(i)
		}
	}

	for i := 0; i < upperBound; i++ {
		expected := i % 2 == 0
		actual := set.Contains(i)
		test.AssertEquals(fmt.Sprintf("error at item %s", i), expected, actual)
	}
}

