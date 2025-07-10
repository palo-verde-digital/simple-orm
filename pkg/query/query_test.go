package query

import "testing"

func Test_Eq(t *testing.T) {
	expected := "id = $1"
	actual, values := Eq("id", 1).Build()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	if values[0] != 1 {
		t.Errorf("expected 1, got %v", values[0])
	}
}

func Test_NotEq(t *testing.T) {
	expected := "id <> $1"
	actual, values := NotEq("id", 1).Build()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	if values[0] != 1 {
		t.Errorf("expected 1, got %v", values[0])
	}
}

func Test_Less(t *testing.T) {
	expected := "id < $1"
	actual, values := Less("id", 1).Build()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	if values[0] != 1 {
		t.Errorf("expected 1, got %v", values[0])
	}
}

func Test_LessEq(t *testing.T) {
	expected := "id <= $1"
	actual, values := LessEq("id", 1).Build()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	if values[0] != 1 {
		t.Errorf("expected 1, got %v", values[0])
	}
}

func Test_Greater(t *testing.T) {
	expected := "id > $1"
	actual, values := Greater("id", 1).Build()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	if values[0] != 1 {
		t.Errorf("expected 1, got %v", values[0])
	}
}

func Test_GreaterEq(t *testing.T) {
	expected := "id >= $1"
	actual, values := GreaterEq("id", 1).Build()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	if values[0] != 1 {
		t.Errorf("expected 1, got %v", values[0])
	}
}

func Test_And(t *testing.T) {
	expected := "(id < $1) AND (age > $2)"
	actual, values := And(Less("id", 3), Greater("age", 25)).Build()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	if values[0] != 3 {
		t.Errorf("expected 3, got %v", values[0])
	}

	if values[1] != 25 {
		t.Errorf("expected 3, got %v", values[1])
	}
}

func Test_Or(t *testing.T) {
	expected := "(id < $1) OR (age > $2)"
	actual, values := Or(Less("id", 3), Greater("age", 25)).Build()

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	if values[0] != 3 {
		t.Errorf("expected 3, got %v", values[0])
	}

	if values[1] != 25 {
		t.Errorf("expected 3, got %v", values[1])
	}
}
