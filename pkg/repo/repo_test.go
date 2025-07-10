package repo_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/palo-verde-digital/simple-orm/pkg/query"
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
	Id       uuid.UUID `db:"id" relation:"PK"`
	Username string    `db:"username"`
	Created  time.Time `db:"created"`
	LastSeen time.Time `db:"last_seen"`
}

func newPg() (*postgres.PostgresContainer, *sqlx.DB) {
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

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Panicf("unable to get postgres conn str: %s", err.Error())
	}

	conn, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Panicf("unable to connect to postgres container via %s: %s", connStr,
			err.Error())
	}

	return pg, conn
}

func count() int {
	var count int
	err := pgConn.QueryRow("SELECT COUNT(*) FROM palo_verde.user;").Scan(&count)

	if err != nil {
		log.Panicf("unable to get user count: %s", err.Error())
	}

	return count
}

func cleanup() {
	_, err := pgConn.Exec("DELETE FROM palo_verde.user;")
	if err != nil {
		log.Panicf("unable to clean up table: %s", err.Error())
	}
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

	r := repo.NewRepository[User](pgConn, testSchema, testTable)
	u1 := User{Id: uuid.New(), Username: "TEST_USER_1", Created: time.Now(), LastSeen: time.Now()}

	err := r.Create(u1)
	if err != nil {
		t.Errorf("error occured: %s", err.Error())
	}

	if count() != 1 {
		t.Errorf("expected 1, got %d", count())
	}

	users := []User{
		User{Id: uuid.New(), Username: "TEST_USER_2", Created: time.Now(), LastSeen: time.Now()},
		User{Id: uuid.New(), Username: "TEST_USER_3", Created: time.Now(), LastSeen: time.Now()},
	}

	err = r.Create(users...)
	if err != nil {
		t.Errorf("error occured: %s", err.Error())
	}

	if count() != 3 {
		t.Errorf("expected 3, got %d", count())
	}

	if cleanup(); count() != 0 {
		t.Errorf("expected 0, got %d", count())
	}
}

func Test_Read(t *testing.T) {
	if count() != 0 {
		t.Errorf("expected 0, got %d", count())
	}

	r := repo.NewRepository[User](pgConn, testSchema, testTable)
	users := []User{
		User{Id: uuid.New(), Username: "TEST_USER_1", Created: time.Now(), LastSeen: time.Now()},
		User{Id: uuid.New(), Username: "TEST_USER_2", Created: time.Now(), LastSeen: time.Now()},
		User{Id: uuid.New(), Username: "TEST_USER_3", Created: time.Now(), LastSeen: time.Now()},
	}

	err := r.Create(users...)
	if err != nil {
		t.Errorf("error occured: %s", err.Error())
	}

	result, err := r.Read(query.All(), 1)
	if err != nil {
		t.Errorf("error occured: %s", err.Error())
	}

	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}

	result, err = r.Read(query.Eq("id", users[0].Id), 0)
	if err != nil {
		t.Errorf("error occured: %s", err.Error())
	}

	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}

	where := query.Or(query.Eq("username", "TEST_USER_2"), query.Eq("username", "TEST_USER_3"))
	result, err = r.Read(where, 0)
	if err != nil {
		t.Errorf("error occured: %s", err.Error())
	}

	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}
