package repo_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/palo-verde-digital/simple-orm/pkg/repo"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	dbName     = "pvd"
	dbUser     = "pvdAdmin"
	dbPassword = "dbAssword"
)

var (
	pgContainer, pgConn = newPg()
)

type User struct {
	id       uuid.UUID `column:"id" relation:"PK"`
	username string    `column:"username"`
	created  time.Time `column:"created"`
	lastSeen time.Time `column:"last_seen"`
}

func newPg() (*postgres.PostgresContainer, *pgx.Conn) {
	pg, err := postgres.Run(context.Background(), "postgres:16-alpine",
		postgres.WithInitScripts("../../../sql/pvd.sql"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
	)

	if err != nil {
		log.Panicf("Unable to create postgres container: %s", err.Error())
	}

	dbHost, err := pg.Host(context.Background())
	if err != nil {
		log.Panicf("Unable to get postgres container host: %s", err.Error())
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Panicf("Unable to connect to postgres container via %s: %s", connStr,
			err.Error())
	}

	return pg, conn
}

func Test_NewRepository(t *testing.T) {
	repo.NewRepository[User](pgConn, "pvd_test", "user")
}

func Test_NewRepository_PanicOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	repo.NewRepository[User](nil, "pvd_test", "user")
}
