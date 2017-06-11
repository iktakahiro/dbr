package fjord

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iktakahiro/fjord/dialect"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

//
// Test helpers
//

var (
	currID int64 = 256
)

// create id
func nextID() int64 {
	currID++
	return currID
}

const (
	mysqlDSN    = "root@unix(/tmp/mysql.sock)/fj_test?charset=utf8"
	postgresDSN = "postgres://postgres@localhost:5432/fj_test?sslmode=disable"
)

func createSession(driver, dsn string) *Session {
	var testDSN string
	switch driver {
	case "mysql":
		testDSN = os.Getenv("FJORD_TEST_MYSQL_DSN")
	case "postgres":
		testDSN = os.Getenv("FJORD_TEST_POSTGRES_DSN")
	}
	if testDSN != "" {
		dsn = testDSN
	}
	conn, err := Open(driver, dsn, nil)
	if err != nil {
		log.Fatal(err)
	}
	sess := conn.NewSession(nil)
	reset(sess)
	return sess
}

func getConnection(driver, dsn string) *Connection {
	var testDSN string
	switch driver {
	case "mysql":
		testDSN = os.Getenv("FJORD_TEST_MYSQL_DSN")
	case "postgres":
		testDSN = os.Getenv("FJORD_TEST_POSTGRES_DSN")
	}
	if testDSN != "" {
		dsn = testDSN
	}
	conn, err := Open(driver, dsn, nil)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

var (
	mysqlSession          = createSession("mysql", mysqlDSN)
	postgresSession       = createSession("postgres", postgresDSN)
	postgresBinarySession = createSession("postgres", postgresDSN+"&binary_parameters=yes")

	// all test sessions should be here
	testSession = []*Session{mysqlSession, postgresSession}
)

type Person struct {
	ID    int64
	Name  string
	Email string
}

type nullTypedRecord struct {
	Id         int64
	StringVal  NullString
	Int64Val   NullInt64
	Float64Val NullFloat64
	TimeVal    NullTime
	BoolVal    NullBool
}

func reset(sess *Session) {
	var autoIncrementType string
	switch sess.Dialect {
	case dialect.MySQL:
		autoIncrementType = "serial PRIMARY KEY"
	case dialect.PostgreSQL:
		autoIncrementType = "serial PRIMARY KEY"
	}
	for _, v := range []string{
		`DROP TABLE IF EXISTS person`,
		fmt.Sprintf(`CREATE TABLE person (
			id %s,
			name varchar(255) NOT NULL,
			email varchar(255)
		)`, autoIncrementType),

		`DROP TABLE IF EXISTS person2`,
		`CREATE TABLE person2 (
			id BIGINT,
			name varchar(255) NOT NULL
		)`,

		`DROP TABLE IF EXISTS role`,
		`CREATE TABLE role (
			person_id BIGINT,
			name VARCHAR(255) NOT NULL
		)`,

		`DROP TABLE IF EXISTS null_types`,
		fmt.Sprintf(`CREATE TABLE null_types (
			id %s,
			string_val varchar(255) NULL,
			int64_val integer NULL,
			float64_val float NULL,
			time_val timestamp NULL ,
			bool_val bool NULL
		)`, autoIncrementType),
	} {
		_, err := sess.Exec(v)
		if err != nil {
			log.Fatalf("Failed to execute statement: %s, Got error: %s", v, err)
		}
	}
}

func BenchmarkByteaNoBinaryEncode(b *testing.B) {
	benchmarkBytea(b, postgresSession)
}

func BenchmarkByteaBinaryEncode(b *testing.B) {
	benchmarkBytea(b, postgresBinarySession)
}

func benchmarkBytea(b *testing.B, sess *Session) {
	data := bytes.Repeat([]byte("0123456789"), 1000)
	for _, v := range []string{
		`DROP TABLE IF EXISTS bytea_table`,
		`CREATE TABLE bytea_table (
			val bytea
		)`,
	} {
		_, err := sess.Exec(v)
		assert.NoError(b, err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := sess.InsertInto("bytea_table").Pair("val", data).Exec()
		assert.NoError(b, err)
	}
}

func TestBasicCRUD(t *testing.T) {
	for _, sess := range testSession {
		person := Person{
			Name:  "John Titor",
			Email: "john@example.com",
		}
		insertColumns := []string{"name", "email"}
		if sess.Dialect == dialect.PostgreSQL {
			person.ID = nextID()
			insertColumns = []string{"id", "name", "email"}
		}
		// insert
		result, err := sess.InsertInto("person").Columns(insertColumns...).Record(&person).Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		assert.True(t, person.ID > 0)
		// select
		var persons []Person
		count, err := sess.Select("*").From("person").Where(Eq("id", person.ID)).Load(&persons)
		assert.NoError(t, err)
		if assert.Equal(t, 1, count) {
			assert.Equal(t, person.ID, persons[0].ID)
			assert.Equal(t, person.Name, persons[0].Name)
			assert.Equal(t, person.Email, persons[0].Email)
		}

		// update
		result, err = sess.Update("person").Where(Eq("id", person.ID)).Set("name", "John Tailor").Exec()
		assert.NoError(t, err)

		rowsAffected, err = result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		var n NullInt64
		sess.Select("count(*)").From("person").Where("name = ?", "John Tailor").Load(&n)
		assert.EqualValues(t, 1, n.Int64)

		// delete
		result, err = sess.DeleteFrom("person").Where(Eq("id", person.ID)).Exec()
		assert.NoError(t, err)

		rowsAffected, err = result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)
	}
}

type PersonWithTag struct {
	ID   int    `db:"p.id"`
	Name string `db:"p.name"`
}

type RoleWithTag struct {
	PersonID int    `db:"r.person_id"`
	Name     string `db:"r.name"`
}

type PersonForJoin struct {
	PersonWithTag
	RoleWithTag
}

func TestJoin(t *testing.T) {
	for _, sess := range testSession {
		person := &PersonWithTag{
			ID:   2036,
			Name: "John Titor",
		}
		// insert - person
		_, err := sess.InsertInto("person2").Columns("id", "name").Record(person).Exec()
		assert.NoError(t, err)

		role := &RoleWithTag{
			PersonID: 2036,
			Name:     "Time Traveler",
		}
		//insert - role
		_, err = sess.InsertInto("role").Columns("person_id", "name").Record(role).Exec()
		assert.NoError(t, err)

		// select

		personForJoin := new(PersonForJoin)
		_, err = sess.Select(I("p.id"), I("p.name"), I("r.person_id"), I("r.name")).
			From(I("person2").As("p")).
			LeftJoin(I("role").As("r"), "p.id = r.person_id").
			Where("p.id = ?", 2036).
			Load(personForJoin)

		assert.NoError(t, err)
		assert.Equal(t, personForJoin.PersonWithTag.ID, 2036)
		assert.Equal(t, personForJoin.PersonWithTag.Name, "John Titor")
		assert.Equal(t, personForJoin.RoleWithTag.PersonID, 2036)
		assert.Equal(t, personForJoin.RoleWithTag.Name, "Time Traveler")
	}
}

func TestContextCancel(t *testing.T) {

	for _, sess := range testSession {
		sess.Close()
	}

	mysqlConnection := getConnection("mysql", mysqlDSN)
	postgresConnection := getConnection("postgres", postgresDSN)
	for _, connection := range []*Connection{mysqlConnection, postgresConnection} {
		checkSessionContext(t, connection)
		checkTxQueryContext(t, connection)
		checkTxExecContext(t, connection)
	}
}

func checkSessionContext(t *testing.T, conn *Connection) {
	ctx, cancel := context.WithCancel(context.Background())
	sess := conn.NewSessionContext(ctx, nil)

	cancel()

	var one int
	_, err := sess.SelectBySql("SELECT 1").Load(&one)
	if err.Error() != "context canceled" {
		t.Errorf("context was not canceled: %v", err)
	}

	_, err = sess.Update("person").Where(Eq("id", 1)).Set("name", "john Titor").Exec()
	if err.Error() != "context canceled" {
		t.Errorf("context was not canceled: %v", err)
	}

	_, err = sess.BeginTx()
	if err.Error() != "context canceled" {
		t.Errorf("context was not canceled: %v", err)
	}
}

func checkTxQueryContext(t *testing.T, conn *Connection) {
	ctx, cancel := context.WithCancel(context.Background())
	sess := conn.NewSessionContext(ctx, nil)
	tx, err := sess.BeginTx()

	if err != nil {
		cancel()
		t.Errorf("transaction was not begun: %v", err)
	}
	cancel()

	var one int
	_, err = tx.SelectBySql("SELECT 1").Load(&one)

	if err.Error() != "context canceled" {
		t.Errorf("context was not canceled: %v", err)
	}

	tx.RollbackUnlessCommitted()
}

func checkTxExecContext(t *testing.T, conn *Connection) {
	ctx, cancel := context.WithCancel(context.Background())
	sess := conn.NewSessionContext(ctx, nil)
	tx, err := sess.BeginTx()
	if !assert.NoError(t, err) {
		cancel()
		return
	}
	_, err = tx.Update("person").Where(Eq("id", 1)).Set("name", "john Titor").Exec()
	if err != nil {
		t.Errorf("failed to update database: %v", err)
	}
	cancel()

	if tx.Commit().Error() != "context canceled" {
		t.Errorf("context was not canceled: %v", err)
	}
}
