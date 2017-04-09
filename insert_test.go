package fjord

import (
	"testing"

	"github.com/iktakahiro/fjord/dialect"
	"github.com/stretchr/testify/assert"
)

type insertTest struct {
	A int
	C string `db:"b"`
	E string `db:"ignore_prefix.e"`
}

func TestInsertStmt(t *testing.T) {
	buf := NewBuffer()
	builder := InsertInto("table").Columns("a", "b", "e").Values(1, "one", "eins").Record(&insertTest{
		A: 2,
		C: "two",
		E: "zwei",
	})
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO `table` (`a`,`b`,`e`) VALUES (?,?,?), (?,?,?)", buf.String())
	assert.Equal(t, []interface{}{1, "one", "eins", 2, "two", "zwei"}, buf.Value())
}

func BenchmarkInsertValuesSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		InsertInto("table").Columns("a", "b", "e").Values(1, "one", "eins").Build(dialect.MySQL, buf)
	}
}

func BenchmarkInsertRecordSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		InsertInto("table").Columns("a", "b", "e").Record(&insertTest{
			A: 2,
			C: "two",
			E: "zwei",
		}).Build(dialect.MySQL, buf)
	}
}
