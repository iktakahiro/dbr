package fjord

import (
	"context"
	"database/sql"
)

// Tx is a transaction for the given Session
type Tx struct {
	EventReceiver
	Dialect Dialect
	*sql.Tx
	ctx context.Context
}

// BeginTx starts a transaction with context.
func (sess *Session) BeginTx() (*Tx, error) {
	tx, err := sess.Connection.BeginTx(sess.ctx, nil)
	if err != nil {
		return nil, sess.EventErr("fjord.begin.error", err)
	}
	sess.Event("fjord.begin")

	return &Tx{
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		Tx:            tx,
		ctx:           sess.ctx,
	}, nil
}

// Begin creates a transaction for the given session
func (sess *Session) Begin() (*Tx, error) {
	tx, err := sess.Connection.Begin()
	if err != nil {
		return nil, sess.EventErr("fjord.begin.error", err)
	}
	sess.Event("fjord.begin")

	return &Tx{
		EventReceiver: sess,
		Dialect:       sess.Dialect,
		Tx:            tx,
	}, nil
}

// Commit finishes the transaction
func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	if err != nil {
		return tx.EventErr("fjord.commit.error", err)
	}
	tx.Event("fjord.commit")
	return nil
}

// Rollback cancels the transaction
func (tx *Tx) Rollback() error {
	err := tx.Tx.Rollback()
	if err != nil {
		return tx.EventErr("fjord.rollback", err)
	}
	tx.Event("fjord.rollback")
	return nil
}

// RollbackUnlessCommitted rollsback the transaction unless it has already been committed or rolled back.
// Useful to defer tx.RollbackUnlessCommitted() -- so you don't have to handle N failure cases
// Keep in mind the only way to detect an error on the rollback is via the event log.
func (tx *Tx) RollbackUnlessCommitted() {
	err := tx.Tx.Rollback()
	if err == sql.ErrTxDone {
		// ok
	} else if err != nil {
		tx.EventErr("fjord.rollback_unless_committed", err)
	} else {
		tx.Event("fjord.rollback")
	}
}
