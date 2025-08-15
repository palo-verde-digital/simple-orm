package orm

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type Condition interface {
	build() (string, []any)
}

func fixArgs(expr string) string {
	found := 0

	for i := range expr {
		if expr[i] == '$' {
			found = found + 1
			start := expr[:i] + fmt.Sprintf("$%d", found)

			offset := 2
			for j := 0; expr[i+j] >= '0' && expr[i+j] <= '9'; {
				offset = offset + j
			}

			expr = start + expr[i+offset:]
		}
	}

	return expr
}

type comparison struct {
	op, column string
	operand    any
}

func (c comparison) build() (string, []any) {
	return c.column + " " + c.op + " $1", []any{c.operand}
}

func Eq(col string, operand any) Condition {
	return comparison{"=", col, operand}
}

func NotEq(col string, operand any) Condition {
	return comparison{"<>", col, operand}
}

func Less(col string, operand any) Condition {
	return comparison{"<", col, operand}
}

func LessEq(col string, operand any) Condition {
	return comparison{"<=", col, operand}
}

func Greater(col string, operand any) Condition {
	return comparison{">", col, operand}
}

func GreaterEq(col string, operand any) Condition {
	return comparison{">=", col, operand}
}

type logical struct {
	op     string
	c1, c2 Condition
}

func (l logical) build() (string, []any) {
	condition1, arg1 := l.c1.build()
	condition2, arg2 := l.c2.build()

	queryStr := fmt.Sprintf("(%s) %s (%s)", condition1, l.op, condition2)
	queryStr = fixArgs(queryStr)

	return queryStr, append(arg1, arg2...)
}

func And(c1, c2 Condition) Condition {
	return logical{op: "AND", c1: c1, c2: c2}
}

func Or(c1, c2 Condition) Condition {
	return logical{op: "OR", c1: c1, c2: c2}
}

type allRows struct{}

func (a allRows) build() (string, []any) {
	return "", nil
}

func All() Condition {
	return allRows{}
}

func ValueOf(v reflect.Value) (interface{}, error) {
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid reflect.Value")
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Array, reflect.Slice:
		if u, isUUID := v.Interface().(uuid.UUID); isUUID {
			return u.String(), nil
		}

		if v.Type().Elem().Kind() == reflect.Uint8 {
			return v.Bytes(), nil
		}

		return nil, fmt.Errorf("unsupported array/slice type: %s", v.Type())
	case reflect.Pointer:
		if v.IsNil() {
			return nil, nil
		}

		return ValueOf(v.Elem())
	case reflect.Struct:
		if t, isTime := v.Interface().(time.Time); isTime {
			return t, nil
		}

		if u, isUUID := v.Interface().(uuid.UUID); isUUID {
			return u.String(), nil
		}

		return nil, fmt.Errorf("unsupported struct type: %s", v.Type())
	case reflect.Interface:
		if v.IsNil() {
			return nil, nil
		}

		return ValueOf(v.Elem())
	default:
		return nil, fmt.Errorf("unsupported type: %s", v.Kind())
	}
}
