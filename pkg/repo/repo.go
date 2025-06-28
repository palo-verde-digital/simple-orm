package repo

import (
	"log"
	"reflect"

	"github.com/jmoiron/sqlx"
)

type Repository[T any] struct {
	conn *sqlx.DB
}

func NewRepository[T any](conn *sqlx.DB) *Repository[T] {
	if conn == nil || conn.Ping() != nil {
		log.Panicf("Invalid DB connection supplied to %s repository",
			reflect.TypeFor[T]().Name())
	}

	return &Repository[T]{
		conn: conn,
	}
}
