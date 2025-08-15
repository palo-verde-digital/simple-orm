package orm

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	tables = make(map[reflect.Type]*table)
)

type column struct {
	columnName    string
	isPk, isFk    bool
	foreignColumn string
	foreignTable  *table
}

type table struct {
	schemaName, tableName string
	primaryKey            *column
	orderedColumns        []string
	columns               map[string]*column

	ins, sel, upd, del string
}

func newTable(t reflect.Type, schemaName, tableName string) (*table, error) {
	if cachedTable := tables[t]; cachedTable != nil {
		return cachedTable, nil
	}

	return createTable(t, schemaName, tableName)
}

func createTable(t reflect.Type, schemaName, tableName string) (*table, error) {
	table := &table{
		schemaName:     schemaName,
		tableName:      tableName,
		primaryKey:     nil,
		orderedColumns: []string{},
		columns:        make(map[string]*column),
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.Tag.Get("db") != "" {
			column, err := createColumn(f, table, schemaName, tableName)
			if err != nil {
				return nil, err
			}

			table.columns[f.Name] = column
			table.orderedColumns = append(table.orderedColumns, column.columnName)
		}
	}

	generateQueries(table)

	tables[t] = table
	return table, nil
}

func createColumn(f reflect.StructField, t *table, schemaName, tableName string) (*column, error) {
	columnName := f.Tag.Get("db")
	rel := f.Tag.Get("relation")

	column := &column{columnName: columnName}

	if rel == "PK" {
		if t.primaryKey != nil {
			return nil, fmt.Errorf("multiple primary keys for %s: %s, %s",
				t.tableName, t.primaryKey.columnName, columnName)
		}

		column.isPk = true
		t.primaryKey = column
	}

	if rel == "FK" {
		fTable := f.Tag.Get("ft")
		if fTable == "" {
			return nil, fmt.Errorf("foreign table for entity %s is missing", t.tableName)
		}

		fCol := f.Tag.Get("fk")
		if fCol == "" {
			return nil, fmt.Errorf("foreign key for entity %s field %s is missing", t.tableName, f.Name)
		}

		if f.Type.Kind() != reflect.Pointer {
			return nil, fmt.Errorf("foreign key %s.%s is %s, expected %s",
				tableName, fCol, f.Type.Kind().String(), reflect.Pointer.String())
		}

		if f.Type.Elem().Kind() != reflect.Struct {
			return nil, fmt.Errorf("foreign key %s.%s element is %s, expected %s",
				tableName, fCol, f.Type.Elem().Kind().String(), reflect.Struct.String())
		}

		column.isFk = true
		column.foreignColumn = fCol

		foreignTable, err := newTable(f.Type.Elem(), schemaName, fTable)
		if err != nil {
			return nil, err
		}

		column.foreignTable = foreignTable
	}

	return column, nil
}

func generateQueries(t *table) {
	columnNames, placeholders := []string{}, []string{}

	for i := range t.orderedColumns {
		columnNames = append(columnNames, t.orderedColumns[i])
		placeholders = append(placeholders, ":"+t.orderedColumns[i])
	}

	t.ins = fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		t.schemaName,
		t.tableName,
		strings.Join(columnNames, ", "),
		strings.Join(placeholders, ", "),
	)

	t.sel = fmt.Sprintf("SELECT %s FROM %s.%s",
		strings.Join(columnNames, ", "),
		t.schemaName,
		t.tableName,
	)

	t.upd = fmt.Sprintf("UPDATE %s.%s SET ", t.schemaName, t.tableName)
	t.del = fmt.Sprintf("DELETE FROM %s.%s", t.schemaName, t.tableName)
}
