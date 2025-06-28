package repo

import (
	"context"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5"
)

type Repository[T any] struct {
	conn   *pgx.Conn
	schema string
	table  string
}

func NewRepository[T any](conn *pgx.Conn, schema, table string) *Repository[T] {
	if conn == nil || conn.Ping(context.Background()) != nil {
		log.Panicf("Invalid DB connection supplied to %s repository",
			reflect.TypeFor[T]().Name())
	}

	return &Repository[T]{
		conn:   conn,
		schema: schema,
		table:  table,
	}
}
