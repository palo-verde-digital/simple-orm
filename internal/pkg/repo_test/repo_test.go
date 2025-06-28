package repo_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/palo-verde-digital/simple-orm/pkg/repo"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	dbName     = "pvd"
	dbUser     = "pvdAdmin"
	dbPassword = "dbAssword"
)

var (
	pgContainer, pgHost = newPgContainer()
	pgConn              = fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
		dbUser, dbPassword, pgHost, dbName)
)

type User struct {
	id       uuid.UUID `column:"id" relation:"PK"`
	username string    `column:"username"`
	created  time.Time `column:"created"`
	lastSeen time.Time `column:"last_seen"`
}

func newPgContainer() (*postgres.PostgresContainer, string) {
	ctx := context.Background()

	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.WithInitScripts(),
	)

	if err != nil {
		log.Panicf("Unable to create postgres container: %s", err.Error())
	}

	host, err := pg.Host(ctx)

	if err != nil {
		log.Panicf("Unable to get postgres container host: %s", err.Error())
	}

	return pg, host
}

func Test_NewRepository(t *testing.T) {
	conn, err := sqlx.Connect("pgx", pgConn)
	if err != nil {
		log.Panicf("Unable to connect to postgres via %s", pgConn)
	}

	repo.NewRepository[User](conn)
}

func Test_NewRepository_PanicOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	repo.NewRepository[User](nil)
}
