package orm

import (
	"fmt"
	"reflect"
	"strings"
)

type column struct {
	name, field string
	isPk, isFk  bool
}

type table[T any] struct {
	schema, name string
	pk           *column
	columns      []*column

	ins, sel, upd, del string
}

func newTable[T any](schema, name string) (*table[T], error) {
	t := reflect.TypeFor[T]()
	table := &table[T]{
		schema:  schema,
		name:    name,
		pk:      nil,
		columns: []*column{},
	}

	columnNames := []string{}
	columns := []*column{}

	for i := range t.NumField() {
		f := t.Field(i)

		if columnName, rel := f.Tag.Get("db"), f.Tag.Get("relation"); columnName != "" {
			column := &column{name: columnName, field: f.Name}

			if rel == "PK" {
				if table.pk != nil {
					return nil, fmt.Errorf("multiple primary keys for %s: %s, %s",
						t.Name(), table.pk.name, columnName)
				}

				column.isPk = true
				table.pk = column
			}

			columns = append(columns, column)
			columnNames = append(columnNames, columnName)
		}
	}

	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = ":" + columnNames[i]
	}

	table.ins = fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s);", schema, name,
		strings.Join(columnNames, ", "), strings.Join(placeholders, ", "))
	table.sel = fmt.Sprintf("SELECT %s FROM %s.%s", strings.Join(columnNames, ", "), schema, name)
	table.upd = fmt.Sprintf("UPDATE %s.%s SET ", schema, name)
	table.del = fmt.Sprintf("DELETE FROM %s.%s", schema, name)

	return table, nil
}
