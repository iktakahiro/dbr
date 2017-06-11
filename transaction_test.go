package fjord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionCommit(t *testing.T) {
	for _, conn := range testConnections {
		sess := conn.NewSession(nil)
		tx, err := sess.Begin()
		assert.NoError(t, err)
		defer tx.RollbackUnlessCommitted()

		id := nextID()

		result, err := tx.InsertInto("person").Columns("id", "name", "email").Values(id, "Barack", "obama@whitehouse.gov").Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		err = tx.Commit()
		assert.NoError(t, err)

		var person Person
		_, err = tx.Select("*").From("person").Where(Eq("id", id)).Load(&person)
		assert.Error(t, err)
	}
}

func TestTransactionRollback(t *testing.T) {
	for _, conn := range testConnections {
		sess := conn.NewSession(nil)
		tx, err := sess.Begin()
		assert.NoError(t, err)
		defer tx.RollbackUnlessCommitted()

		id := nextID()

		result, err := tx.InsertInto("person").Columns("id", "name", "email").Values(id, "Barack", "obama@whitehouse.gov").Exec()
		assert.NoError(t, err)

		rowsAffected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, rowsAffected)

		err = tx.Rollback()
		assert.NoError(t, err)

		var person Person
		_, err = tx.Select("*").From("person").Where(Eq("id", id)).Load(&person)
		assert.Error(t, err)
	}
}
