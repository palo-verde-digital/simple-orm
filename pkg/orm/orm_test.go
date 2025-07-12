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
		where    orm.Condition
		lim, len int
	}{
		"limit":     {where: orm.All(), lim: 1, len: 1},
		"all":       {where: orm.All(), lim: 0, len: 3},
		"eq":        {where: orm.Eq("username", "TEST_USER_1"), lim: 0, len: 1},
		"notEq":     {where: orm.NotEq("logins", 18), lim: 0, len: 2},
		"less":      {where: orm.Less("logins", 18), lim: 0, len: 0},
		"lessEq":    {where: orm.LessEq("logins", 25), lim: 0, len: 2},
		"greater":   {where: orm.Greater("logins", 18), lim: 0, len: 2},
		"greaterEq": {where: orm.GreaterEq("logins", 18), lim: 0, len: 3},
		"and":       {where: orm.And(orm.Greater("logins", 20), orm.Eq("id", users[1].Id)), lim: 0, len: 1},
		"or":        {where: orm.Or(orm.Eq("username", "TEST_USER_2"), orm.Eq("username", "TEST_USER_3")), lim: 0, len: 2},
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
			result, err := r.Read(test.where, test.lim)
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

func Test_Update(t *testing.T) {
	tests := map[string]struct {
		where, verify orm.Condition
		found         int
		updates       map[string]any
	}{
		"all": {
			where:  orm.All(),
			verify: orm.Eq("username", "TEST_USER"),
			found:  3,
			updates: map[string]any{
				"username": "TEST_USER",
			},
		},
		"eq": {
			where:  orm.Eq("username", "TEST_USER_1"),
			verify: orm.Eq("logins", 0),
			found:  1,
			updates: map[string]any{
				"logins": 0,
			},
		},
		"notEq": {
			where:  orm.NotEq("username", "TEST_USER_1"),
			verify: orm.Eq("logins", 0),
			found:  2,
			updates: map[string]any{
				"logins": 0,
			},
		},
		"less": {
			where:  orm.Less("logins", 20),
			verify: orm.Eq("logins", 0),
			found:  1,
			updates: map[string]any{
				"logins": 0,
			},
		},
		"lessEq": {
			where:  orm.LessEq("logins", 25),
			verify: orm.Eq("logins", 0),
			found:  2,
			updates: map[string]any{
				"logins": 0,
			},
		},
		"greater": {
			where:  orm.Greater("logins", 18),
			verify: orm.Eq("logins", 0),
			found:  2,
			updates: map[string]any{
				"logins": 0,
			},
		},
		"greaterEq": {
			where:  orm.GreaterEq("logins", 18),
			verify: orm.Eq("logins", 0),
			found:  3,
			updates: map[string]any{
				"logins": 0,
			},
		},
		"and": {
			where:  orm.And(orm.Greater("logins", 20), orm.Eq("id", users[1].Id)),
			verify: orm.Eq("username", "MODIFIED_USERNAME"),
			found:  1,
			updates: map[string]any{
				"username": "MODIFIED_USERNAME",
			},
		},
		"or": {
			where:  orm.Or(orm.Eq("username", "TEST_USER_2"), orm.Eq("username", "TEST_USER_3")),
			verify: orm.Eq("username", "MODIFIED_USERNAME"),
			found:  2,
			updates: map[string]any{
				"username": "MODIFIED_USERNAME",
			},
		},
	}

	r, err := orm.NewRepository[User](pgConn, testSchema, testTable)
	if err != nil {
		t.Fatalf("unexpected error occurred: %s", err.Error())
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err = r.Create(users...); err != nil {
				t.Fatalf("unexpected error occurred: %s", err.Error())
			}

			err := r.Update(test.updates, test.where)
			if err != nil {
				t.Fatalf("unexpected error occurred: %s", err.Error())
			}

			res, err := r.Read(test.verify, 0)
			if len(res) != test.found {
				t.Errorf("expected %d, got %d", test.found, len(res))
			}

			cleanup()
		})
	}
}

func Test_Delete(t *testing.T) {
	tests := map[string]struct {
		where orm.Condition
		rem   int
	}{
		"all":       {where: orm.All(), rem: 0},
		"eq":        {where: orm.Eq("username", "TEST_USER_1"), rem: 2},
		"notEq":     {where: orm.NotEq("logins", 18), rem: 1},
		"less":      {where: orm.Less("logins", 18), rem: 3},
		"lessEq":    {where: orm.LessEq("logins", 25), rem: 1},
		"greater":   {where: orm.Greater("logins", 18), rem: 1},
		"greaterEq": {where: orm.GreaterEq("logins", 18), rem: 0},
		"and":       {where: orm.And(orm.Greater("logins", 20), orm.Eq("id", users[1].Id)), rem: 2},
		"or":        {where: orm.Or(orm.Eq("username", "TEST_USER_2"), orm.Eq("username", "TEST_USER_3")), rem: 1},
	}

	r, err := orm.NewRepository[User](pgConn, testSchema, testTable)
	if err != nil {
		t.Fatalf("unexpected error occurred: %s", err.Error())
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err = r.Create(users...); err != nil {
				t.Fatalf("unexpected error occurred: %s", err.Error())
			}

			if err = r.Delete(test.where); err != nil {
				t.Fatalf("unexpected error occurred: %s", err.Error())
			}

			if where := count(); where != test.rem {
				t.Errorf("expected %d, got %d", test.rem, where)
			}

			cleanup()
		})
	}
}
