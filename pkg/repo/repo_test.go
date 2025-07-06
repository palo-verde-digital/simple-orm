package repo_test

import (
	"context"
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
	dbPassword = "pvdAssword"

	testSchema = "palo_verde"
	testTable  = "user"
)

var (
	pgContainer, pgConn = newPg()
)

type User struct {
	Id       uuid.UUID `column:"id" relation:"PK"`
	Username string    `column:"username"`
	Created  time.Time `column:"created"`
	LastSeen time.Time `column:"last_seen"`
}

func newPg() (*postgres.PostgresContainer, *pgx.Conn) {
	ctx := context.Background()

	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithInitScripts("../../sql/pvd.sql"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		log.Panicf("unable to create postgres container: %s", err.Error())
	}

	connStr, err := pg.ConnectionString(ctx)
	if err != nil {
		log.Panicf("unable to get postgres conn str: %s", err.Error())
	}

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Panicf("unable to connect to postgres container via %s: %s", connStr,
			err.Error())
	}

	return pg, conn
}

func count() int {
	var count int
	err := pgConn.QueryRow(context.Background(), "SELECT COUNT(*) FROM palo_verde.user").Scan(&count)

	if err != nil {
		log.Panicf("unable to get user count: %s", err.Error())
	}

	return count
}

func Test_NewRepository(t *testing.T) {
	repo.NewRepository[User](pgConn, testSchema, testTable)
}

func Test_NewRepository_PanicOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	repo.NewRepository[User](nil, testSchema, "user")
}

func Test_Create(t *testing.T) {
	if count() != 0 {
		t.Errorf("expected 0, got %d", count())
	}

	r := repo.NewRepository[User](pgConn, testSchema, "user")
	u := User{
		Id:       uuid.New(),
		Username: "TEST",
		Created:  time.Now(),
		LastSeen: time.Now(),
	}

	err := r.Create(u)
	if err != nil {
		t.Errorf("error occured: %s", err.Error())
	}

	if count() != 1 {
		t.Errorf("expected 1, got %d", count())
	}
}
