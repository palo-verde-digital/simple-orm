package repo

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Condition interface {
	build() string
}

type Comparison struct {
	op, operand1, operand2 string
}

func (c *Comparison) build() string {
	return c.operand1 + " " + c.op + " " + c.operand2
}

func Eq(op1, op2 string) Comparison {
	return Comparison{op: "=", operand1: op1, operand2: op2}
}

func NotEq(op1, op2 string) Comparison {
	return Comparison{op: "<>", operand1: op1, operand2: op2}
}

func Less(op1, op2 string) Comparison {
	return Comparison{op: "<", operand1: op1, operand2: op2}
}

func LessEq(op1, op2 string) Comparison {
	return Comparison{op: "<=", operand1: op1, operand2: op2}
}

func Greater(op1, op2 string) Comparison {
	return Comparison{op: ">", operand1: op1, operand2: op2}
}

func GreaterEq(op1, op2 string) Comparison {
	return Comparison{op: ">", operand1: op1, operand2: op2}
}

type Logical struct {
	op     string
	c1, c2 Condition
}

func (l *Logical) build() string {
	return fmt.Sprintf("(%s) %s (%s)", l.c1.build(), l.op, l.c2.build())
}

type Repository[T any] struct {
	conn          *pgx.Conn
	schema, table string
	columns       []column
}

type column struct {
	name, field string
}

func NewRepository[T any](conn *pgx.Conn, schema, table string) *Repository[T] {
	t := reflect.TypeFor[T]()
	log.Printf("creating new Repository[%s]", t.Name())

	if conn == nil || conn.Ping(context.Background()) != nil {
		log.Panicf("Invalid DB connection supplied to Repository[%s]", t.Name())
	}

	log.Printf("verified Repository[%s] conn", t.Name())
	columns := columnsFor[T]()

	return &Repository[T]{
		conn:    conn,
		schema:  schema,
		table:   table,
		columns: columns,
	}
}

func columnsFor[T any]() []column {
	t := reflect.TypeFor[T]()

	columns := []column{}
	for i := range t.NumField() {
		if columnName := t.Field(i).Tag.Get("column"); columnName != "" {
			column := column{
				name:  columnName,
				field: t.Field(i).Name,
			}

			columns = append(columns, column)
		}
	}

	log.Printf("found %d columns for entity type %s", len(columns), t.Name())
	return columns
}

func queryValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return fmt.Sprintf("'%s'", reflect.String)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Float())
	case reflect.Array, reflect.Slice:
		if u, isUUID := v.Interface().(uuid.UUID); isUUID {
			return fmt.Sprintf("'%s'", u.String())
		}
	case reflect.Pointer:
		return queryValue(v.Elem())
	case reflect.Struct:
		if t, isTime := v.Interface().(time.Time); isTime {
			return fmt.Sprintf("'%s'", t.Format(time.RFC3339))
		}
	}

	return ""
}

func (r *Repository[T]) columnStr() string {
	columnStr := ""

	for i := range r.columns {
		columnStr = columnStr + r.columns[i].name

		if i < len(r.columns)-1 {
			columnStr = columnStr + ", "
		}
	}

	return columnStr
}

func (r *Repository[T]) Create(entities ...T) error {
	insert := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES ", r.schema, r.table, r.columnStr())

	valueStrs := make([]string, len(entities))
	for i := range entities {
		e := reflect.ValueOf(entities[i])

		values := make([]string, len(r.columns))
		for j := range r.columns {
			v := e.FieldByName(r.columns[j].field)
			values[j] = queryValue(v)
		}

		valueStrs[i] = fmt.Sprintf("(%s)", strings.Join(values, ", "))
	}
	values := strings.Join(valueStrs, ", ")

	insert = insert + values + ";"
	log.Printf("executing: %s", insert)

	_, err := r.conn.Exec(context.Background(), insert)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository[T]) Read(where Condition) ([]T, error) {
	return nil, nil
}
