package orm_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/palo-verde-digital/simple-orm/pkg/orm"
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

	users = []User{
		User{Id: uuid.New(), Username: "TEST_USER_1", Logins: 18, Created: time.Now(), LastSeen: time.Now()},
		User{Id: uuid.New(), Username: "TEST_USER_2", Logins: 25, Created: time.Now(), LastSeen: time.Now()},
		User{Id: uuid.New(), Username: "TEST_USER_3", Logins: 40, Created: time.Now(), LastSeen: time.Now()},
	}
)

type User struct {
	Id       uuid.UUID `db:"id" relation:"PK"`
	Username string    `db:"username"`
	Logins   int       `db:"logins"`
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
	tests := map[string]struct {
		hasErr bool
		conn   *sqlx.DB
	}{
		"ok":  {hasErr: false, conn: pgConn},
		"err": {hasErr: true, conn: nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := orm.NewRepository[User](test.conn, testSchema, testTable)
			if err != nil && !test.hasErr {
				t.Fatalf("unexpected error occurred: %s", err.Error())
			}
		})
	}
}

func Test_Create(t *testing.T) {
	r, err := orm.NewRepository[User](pgConn, testSchema, testTable)
	if err != nil {
		t.Fatalf("unexpected error occurred: %s", err.Error())
	}

	err = r.Create(users...)
	if err != nil {
		t.Fatalf("unexpected error occurred: %s", err.Error())
	}

	if count() != 3 {
		t.Errorf("expected 3, got %d", count())
	}

	cleanup()
}

func Test_Read(t *testing.T) {
	tests := map[string]struct {
		c        orm.Condition
		lim, len int
	}{
		"limit":     {c: orm.All(), lim: 1, len: 1},
		"all":       {c: orm.All(), lim: 0, len: 3},
		"eq":        {c: orm.Eq("username", "TEST_USER_1"), lim: 0, len: 1},
		"notEq":     {c: orm.NotEq("logins", 18), lim: 0, len: 2},
		"less":      {c: orm.Less("logins", 18), lim: 0, len: 0},
		"lessEq":    {c: orm.LessEq("logins", 25), lim: 0, len: 2},
		"greater":   {c: orm.Greater("logins", 18), lim: 0, len: 2},
		"greaterEq": {c: orm.GreaterEq("logins", 18), lim: 0, len: 3},
	}

	r, err := orm.NewRepository[User](pgConn, testSchema, testTable)
	if err != nil {
		t.Fatalf("unexpected error occurred: %s", err.Error())
	}

	err = r.Create(users...)
	if err != nil {
		t.Fatalf("unexpected error occurred: %s", err.Error())
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := r.Read(test.c, test.lim)
			if err != nil {
				t.Fatalf("unexpected error occurred: %s", err.Error())
			}

			if len(result) != test.len {
				t.Errorf("expected %d, got %d", test.len, len(result))
			}
		})
	}

	cleanup()
}
