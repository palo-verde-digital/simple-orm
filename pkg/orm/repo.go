package orm

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository[T any] struct {
	conn          *sqlx.DB
	schema, table string
	columns       []column

	ins, sel, upd, del string
}

type column struct {
	name, field string
}

func NewRepository[T any](conn *sqlx.DB, schema, table string) (*Repository[T], error) {
	t := reflect.TypeFor[T]()
	log.Printf("creating Repository[%s]", t.Name())

	if conn == nil || conn.Ping() != nil {
		return nil, fmt.Errorf("Invalid DB connection supplied to Repository[%s]", t.Name())
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

	placeholders := make([]string, len(columnNames))
	for i, name := range columnNames {
		placeholders[i] = ":" + name
	}

	log.Printf("found %d columns for entity type %s", len(columns), t.Name())

	return &Repository[T]{
		conn:    conn,
		schema:  schema,
		table:   table,
		columns: columns,

		ins: fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s);", schema, table, strings.Join(columnNames, ", "), strings.Join(placeholders, ", ")),
		sel: fmt.Sprintf("SELECT %s FROM %s.%s", strings.Join(columnNames, ", "), schema, table),
		upd: fmt.Sprintf("UPDATE %s.%s SET ", schema, table),
		del: fmt.Sprintf("DELETE FROM %s.%s", schema, table),
	}, nil
}

func (r *Repository[T]) Create(entities ...T) error {
	log.Printf("executing: %s for %d %s(s)", r.ins, len(entities), reflect.TypeFor[T]().Name())

	_, err := r.conn.NamedExec(r.ins, entities)
	return err
}

func (r *Repository[T]) Read(filter Condition, limit int) ([]T, error) {
	sel := r.sel

	where, args := filter.build()
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

func (r *Repository[T]) Update(updates map[string]any, filter Condition) error {
	upd := r.upd

	i, expressions, args := 1, []string{}, []any{}
	for col, arg := range updates {
		expr := fmt.Sprintf("%s = $%d", col, i)
		i = i + 1

		expressions, args = append(expressions, expr), append(args, arg)
	}

	upd = upd + strings.Join(expressions, ", ")

	where, filterArgs := filter.build()
	if where != "" {
		args = append(args, filterArgs...)
		upd = fixArgs(upd + " WHERE " + where)
	}

	upd = upd + ";"
	log.Printf("executing: %s", upd)

	_, err := r.conn.Exec(upd, args...)
	return err
}

func (r *Repository[T]) Delete(filter Condition) error {
	del := r.del

	where, args := filter.build()
	if where != "" {
		del = del + " WHERE " + where
	}

	del = del + ";"
	log.Printf("executing: %s", del)

	_, err := r.conn.Exec(del, args...)
	return err
}
