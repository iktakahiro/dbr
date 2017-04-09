package fjord

import (
	"testing"

	"github.com/iktakahiro/fjord/dialect"
	"github.com/stretchr/testify/assert"
)

func TestSelectStmt(t *testing.T) {
	buf := NewBuffer()
	builder := Select("a", "b").
		From(Select("a").From("table")).
		LeftJoin("table2", "table.a1 = table.a2").
		Distinct().
		Where(Eq("c", 1)).
		GroupBy("d").
		Having(Eq("e", 2)).
		OrderAsc("f").
		Limit(3).
		Offset(4)
	err := builder.Build(dialect.MySQL, buf)
	assert.NoError(t, err)
	assert.Equal(t, "SELECT DISTINCT a, b FROM ? LEFT JOIN `table2` ON table.a1 = table.a2 WHERE (`c` = ?) GROUP BY d HAVING (`e` = ?) ORDER BY f ASC LIMIT 3 OFFSET 4", buf.String())
	// two functions cannot be compared
	assert.Equal(t, 3, len(buf.Value()))
}

func TestSelectWithAliasStmt(t *testing.T) {

	for _, d := range []Dialect{dialect.MySQL, dialect.PostgreSQL} {
		buf := NewBuffer()
		builder := Select(I("t1.a"), I("t2.b").As("t2__b"), "c AS t2__c").
			From(I("table1").As("t1")).
			LeftJoin(I("table2").As("t2"), "t1.id = t2.id")
		err := builder.Build(d, buf)
		assert.NoError(t, err)

		switch d {
		case dialect.MySQL:
			assert.Equal(t, "SELECT `t1`.`a` AS t1__a, ?, c AS t2__c FROM ? LEFT JOIN ? ON t1.id = t2.id", buf.String())
		case dialect.PostgreSQL:
			assert.Equal(t, `SELECT "t1"."a" AS t1__a, ?, c AS t2__c FROM ? LEFT JOIN ? ON t1.id = t2.id`, buf.String())
		}
	}
}

func BenchmarkSelectSQL(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		Select("a", "b").From("table").Where(Eq("c", 1)).OrderAsc("d").Build(dialect.MySQL, buf)
	}
}
