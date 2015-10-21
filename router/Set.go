package router
import (
	"os"
	"bufio"
)

type Set map[interface{}]struct{}

func (this *Set) IsEmpty() bool {
	return len(*this) == 0
}

func (this *Set) Add(item interface{}) {
	(*this)[item] = struct{}{}
}

func (this *Set) Contains(item interface{}) bool {
	_, exists := (*this)[item]
	return exists
}

func CreateSetFromFile(path string, processAndAddLine func (set *Set, line string)) (set Set, err error) {
	sourceFile, err := os.Open(path)
	if err != nil { return }
	defer sourceFile.Close()

	set = Set{}
	sourceScanner := bufio.NewScanner(sourceFile)
	for sourceScanner.Scan() {
		line := sourceScanner.Text()
		processAndAddLine(&set, line)
	}

	return set, sourceScanner.Err()
}

