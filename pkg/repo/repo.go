package repo

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/palo-verde-digital/simple-orm/pkg/query"
)

type Repository[T any] struct {
	conn          *sqlx.DB
	schema, table string
	columns       []column

	ins, sel string
}

type column struct {
	name, field string
}

func NewRepository[T any](conn *sqlx.DB, schema, table string) *Repository[T] {
	t := reflect.TypeFor[T]()
	log.Printf("creating Repository[%s]", t.Name())

	if conn == nil || conn.Ping() != nil {
		log.Panicf("Invalid DB connection supplied to Repository[%s]", t.Name())
	}

	log.Printf("verified Repository[%s] conn", t.Name())

	columns := []column{}
	for i := range t.NumField() {
		if columnName := t.Field(i).Tag.Get("db"); columnName != "" {
			col := column{
				name:  columnName,
				field: t.Field(i).Name,
			}

			columns = append(columns, col)
		}
	}

	columnNames := make([]string, len(columns))
	for i, col := range columns {
		columnNames[i] = col.name
	}
	columnStr := strings.Join(columnNames, ", ")

	log.Printf("found %d columns for entity type %s", len(columns), t.Name())

	return &Repository[T]{
		conn:    conn,
		schema:  schema,
		table:   table,
		columns: columns,

		ins: fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (:%s);", schema, table, columnStr, strings.Join(columnNames, ", :")),
		sel: fmt.Sprintf("SELECT %s FROM %s.%s", columnStr, schema, table),
	}
}

func (r *Repository[T]) Create(entities ...T) error {
	log.Printf("executing: %s for %d %s(s)", r.ins, len(entities), reflect.TypeFor[T]().Name())

	_, err := r.conn.NamedExec(r.ins, entities)
	return err
}

func (r *Repository[T]) Read(filter query.Condition, limit int) ([]T, error) {
	sel := r.sel

	where, args := filter.Build()
	if where != "" {
		sel = sel + " WHERE " + where
	}

	if limit > 0 {
		sel = sel + fmt.Sprintf(" LIMIT %d", limit)
	}

	sel = sel + ";"

	log.Printf("executing: %s", sel)

	rows, err := r.conn.Queryx(sel, args...)
	if err != nil {
		return nil, err
	}

	var found []T
	for rows.Next() {
		var row T
		err := rows.StructScan(&row)
		if err != nil {
			return nil, err
		}
		found = append(found, row)
	}

	return found, nil
}
