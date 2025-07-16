package orm

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository[T any] struct {
	conn  *sqlx.DB
	table *table[T]
}

func NewRepository[T any](conn *sqlx.DB, schemaName, tableName string) (*Repository[T], error) {
	t := reflect.TypeFor[T]()

	log.Printf("creating Repository[%s]", t.Name())

	if conn == nil || conn.Ping() != nil {
		return nil, fmt.Errorf("Invalid DB connection supplied to Repository[%s]", t.Name())
	}

	log.Printf("verified Repository[%s] conn", t.Name())

	table, err := newTable[T](schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("Invalid entity %s: %s", t.Name(), err.Error())
	}

	return &Repository[T]{
		conn:  conn,
		table: table,
	}, nil
}

func (r *Repository[T]) Create(entities ...T) error {
	log.Printf("executing: %s for %d %s(s)", r.table.ins, len(entities), reflect.TypeFor[T]().Name())

	_, err := r.conn.NamedExec(r.table.ins, entities)
	return err
}

func (r *Repository[T]) Read(filter Condition, limit int) ([]T, error) {
	sel := r.table.sel

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
	upd := r.table.upd

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
	del := r.table.del

	where, args := filter.build()
	if where != "" {
		del = del + " WHERE " + where
	}

	del = del + ";"
	log.Printf("executing: %s", del)

	_, err := r.conn.Exec(del, args...)
	return err
}
