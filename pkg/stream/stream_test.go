package stream_test

import (
	"strings"
	"testing"

	"github.com/palo-verde-digital/simple-orm/pkg/stream"
)

func Test_Filter(t *testing.T) {
	numbers := make(stream.Stream[string], 3)
	numbers[0], numbers[1], numbers[2] = "one", "two", "three"

	filter := func(s string) bool { return strings.Contains(s, "e") }
	filtered := numbers.Filter(filter)

	if len(filtered) != 2 {
		t.Errorf("expected 2, got %d", len(filtered))
	}
}

func Test_ForEach(t *testing.T) {
	numbers := make(stream.Stream[string], 3)
	numbers[0], numbers[1], numbers[2] = "one", "two", "three"

	firsts := []rune{}
	extractFirstChar := func(s string) { firsts = append(firsts, rune(s[0])) }
	numbers.ForEach(extractFirstChar)

	if len(firsts) != 3 {
		t.Errorf("expected 3, got %d", len(firsts))
	}

	if firsts[0] != 'o' {
		t.Errorf("expected 'o', got '%c'", firsts[0])
	} else if firsts[1] != 't' {
		t.Errorf("expected 't', got '%c'", firsts[1])
	} else if firsts[2] != 't' {
		t.Errorf("expected 't', got '%c'", firsts[2])
	}
}

func Test_Map(t *testing.T) {
	numbers := make(stream.Stream[string], 3)
	numbers[0], numbers[1], numbers[2] = "one", "two", "three"

	toInt := func(s string) any {
		switch s {
		case "one":
			return 1
		case "two":
			return 2
		case "three":
			return 3
		default:
			return 0
		}
	}

	ints := numbers.Map(toInt)
	if ints[0] != 1 {
		t.Errorf("expected 1, got %d", ints[0])
	} else if ints[1] != 2 {
		t.Errorf("expected 2, got %d", ints[1])
	} else if ints[2] != 3 {
		t.Errorf("expected 3, got %d", ints[2])
	}
}
