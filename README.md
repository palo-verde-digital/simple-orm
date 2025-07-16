# simple-orm
simple-orm provides a CRUD repository abstraction for postgreSQL databases

## Prerequisites:
1. go >=1.23

## Add the module to your project:
```bash
go get github.com/palo-verde-digital/simple-orm
```
## Creating a repository:
simple-orm uses `sqlx` perform database operations. Instantiating a new repository requires:
1. `*sqlx.DB` connected to your postgres database
2. `string` entity schema name
3. `string` entity object name

### 
```go
package orm

func NewRepository[T any](conn *sqlx.DB, schemaName, tableName string) (*Repository[T], error)

...

package myPackage

userRepo, err := orm.NewRepository[User](myDb, mySchema, myTable)
if err != nil {
	log.Panicf("unexpected error occurred: %s", err.Error())
}
```

## Tagging entity structs:
Column names must be provided via the `db` tag. If a field is not tagged, it is excluded from the queries and entity metadata computed when `NewRepository` is called.
```go
type User struct {
	Id       uuid.UUID `db:"id"`
	Username string    `db:"username"`
	Created  time.Time `db:"created"`
	LastSeen time.Time `db:"last_seen"`
}
```

## Repository Features:
The repository provides basic CRUD functionalities and query-building for filtering.

### Create
Insert new records
```go
package orm

func (r *Repository[T]) Create(entities ...T) error

...

package myPackage

users := []User{
	User{Id: uuid.New(), Username: "TEST_USER_1", Created: time.Now(), LastSeen: time.Now()},
	User{Id: uuid.New(), Username: "TEST_USER_2", Created: time.Now(), LastSeen: time.Now()},
	User{Id: uuid.New(), Username: "TEST_USER_3", Created: time.Now(), LastSeen: time.Now()},
}

err := userRepo.Create(users...)
```
### Read (and filter)
Query for records
```go
package orm

// limit <= 0 will not be applied
func (r *Repository[T]) Read(filter Condition, limit int) ([]T, error)

// filter conditions are obtained by providing the column and operand
// to an operator function. conditions are composable via 'And' & 'Or' 

func Eq(col string, operand any) Condition
func NotEq(col string, operand any)
func Less(col string, operand any) Condition
func LessEq(col string, operand any) Condition
func Greater(col string, operand any) Condition
func GreaterEq(col string, operand any) Condition
func And(c1, c2 Condition) Condition
func Or(c1, c2 Condition) Condition
func All() Condition

...

package myPackage

allUsers, err := userRepo.Read(orm.All(), 0)

thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
returningUserFilter := orm.And(orm.Greater("created", thirtyDaysAgo),
    orm.LessEq("last_seen", thirtyDaysAgo))

returningUsers, err := userRepo.Read(returningUserFilter, 0)
```
### Update
Modify existing records
```go
package orm

func (r *Repository[T]) Update(updates map[string]any, filter Condition) error

...

package myPackage

thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
inactiveUserFilter := orm.And(orm.Greater("created", thirtyDaysAgo),
    orm.Greater("last_seen", thirtyDaysAgo))

inactiveUsernameUpdate := map[string]any{"username": "INACTIVE USER"} 

err := userRepo.Update(intactiveUsernameUpdate, inactiveUserFilter)
```
### Delete
Delete records
```go
package orm

func (r *Repository[T]) Delete(filter Condition) error

...

package myPackage

thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
inactiveUserFilter := orm.And(orm.Greater("created", thirtyDaysAgo),
    orm.Greater("last_seen", thirtyDaysAgo))

err := userRepo.Delete(inactiveUserFilter)
```