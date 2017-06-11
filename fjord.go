package fjord

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/iktakahiro/fjord/dialect"
)

// Open instantiates a Connection for a given database/sql connection
// and event receiver
func Open(driver, dsn string, log EventReceiver) (*Connection, error) {
	if log == nil {
		log = nullReceiver
	}
	conn, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	var d Dialect
	switch driver {
	case "mysql":
		d = dialect.MySQL
	case "postgres":
		d = dialect.PostgreSQL
	default:
		return nil, ErrNotSupported
	}
	return &Connection{DB: conn, EventReceiver: log, Dialect: d}, nil
}

const (
	placeholder = "?"
)

// Connection is a connection to the database with an EventReceiver
// to send events, errors, and timings to
type Connection struct {
	*sql.DB
	Dialect Dialect
	EventReceiver
}

// Session represents a business unit of execution for some connection
type Session struct {
	*Connection
	EventReceiver
	ctx context.Context
}

// NewSession instantiates a Session for the Connection
func (conn *Connection) NewSession(log EventReceiver) *Session {
	if log == nil {
		log = conn.EventReceiver // Use parent instrumentation
	}
	return conn.NewSessionContext(context.Background(), log)
	// return &Session{Connection: conn, EventReceiver: log}
}

func (conn *Connection) NewSessionContext(ctx context.Context, log EventReceiver) *Session {
	if log == nil {
		log = conn.EventReceiver
	}
	return &Session{Connection: conn, EventReceiver: log, ctx: ctx}
}

// Ensure that tx and session are session runner
var (
	_ SessionRunner = (*Tx)(nil)
	_ SessionRunner = (*Session)(nil)
)

// SessionRunner can do anything that a Session can except start a transaction.
type SessionRunner interface {
	Select(column ...interface{}) *SelectBuilder
	SelectBySql(query string, value ...interface{}) *SelectBuilder

	InsertInto(table string) *InsertBuilder
	InsertBySql(query string, value ...interface{}) *InsertBuilder

	Update(table string) *UpdateBuilder
	UpdateBySql(query string, value ...interface{}) *UpdateBuilder

	DeleteFrom(table string) *DeleteBuilder
	DeleteBySql(query string, value ...interface{}) *DeleteBuilder
}

type runner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (sess *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	return sess.ExecContext(sess.ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (sess *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return sess.QueryContext(sess.ctx, query, args...)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.ExecContext(tx.ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.QueryContext(tx.ctx, query, args...)
}

func exec(runner runner, log EventReceiver, builder Builder, d Dialect) (sql.Result, error) {
	i := interpolator{
		Buffer:       NewBuffer(),
		Dialect:      d,
		IgnoreBinary: true,
	}
	err := i.interpolate(placeholder, []interface{}{builder})
	query, value := i.String(), i.Value()
	if err != nil {
		return nil, log.EventErrKv("fjord.exec.interpolate", err, kvs{
			"sql":  query,
			"args": fmt.Sprint(value),
		})
	}

	startTime := time.Now()
	defer func() {
		log.TimingKv("fjord.exec", time.Since(startTime).Nanoseconds(), kvs{
			"sql": query,
		})
	}()

	result, err := runner.Exec(query, value...)
	if err != nil {
		return result, log.EventErrKv("fjord.exec.exec", err, kvs{
			"sql": query,
		})
	}
	return result, nil
}

func query(ctx context.Context, runner runner, log EventReceiver, builder Builder, d Dialect, dest interface{}) (int, error) {
	i := interpolator{
		Buffer:       NewBuffer(),
		Dialect:      d,
		IgnoreBinary: true,
	}
	err := i.interpolate(placeholder, []interface{}{builder})
	query, value := i.String(), i.Value()
	if err != nil {
		return 0, log.EventErrKv("fjord.select.interpolate", err, kvs{
			"sql":  query,
			"args": fmt.Sprint(value),
		})
	}

	startTime := time.Now()
	defer func() {
		log.TimingKv("fjord.select", time.Since(startTime).Nanoseconds(), kvs{
			"sql": query,
		})
	}()

	rows, err := runner.Query(query, value...)
	if err != nil {
		return 0, log.EventErrKv("fjord.select.load.query", err, kvs{
			"sql": query,
		})
	}

	count, err := load(rows, dest)
	if err != nil {
		return 0, log.EventErrKv("fjord.select.load.scan", err, kvs{
			"sql": query,
		})
	}
	return count, nil
}
