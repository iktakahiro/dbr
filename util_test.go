package fjord

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestColumnNameToAlias(t *testing.T) {
	assert.Equal(t, "user__id", columnNameToAlias("user.id"))
}

type StructForGetColumnNameFromTag struct {
	NonPrefixField int64  `db:"non_prefix_field"`
	PrefixField    string `db:"prefix.prefix_field"`
	IgnoreField    string `db:"-"`
	NonTagField    string
}

func TestGetColumnNameFromTag(t *testing.T) {

	testStruct := &StructForGetColumnNameFromTag{}
	rt := reflect.TypeOf(*testStruct)

	NonPrefixField := rt.Field(0)
	assert.Equal(t, "non_prefix_field", getColumnNameFromTag(NonPrefixField, false))

	PrefixField := rt.Field(1)
	assert.Equal(t, "prefix__prefix_field", getColumnNameFromTag(PrefixField, false))
	// When ignorePrefix flag is true, ignore string before ".":
	assert.Equal(t, "prefix_field", getColumnNameFromTag(PrefixField, true))

	IgnoreField := rt.Field(2)
	assert.Equal(t, "", getColumnNameFromTag(IgnoreField, false))

	NonTagField := rt.Field(3)
	assert.Equal(t, "non_tag_field", getColumnNameFromTag(NonTagField, false))
}

func TestSnakeCase(t *testing.T) {
	for _, test := range []struct {
		in   string
		want string
	}{
		{
			in:   "",
			want: "",
		},
		{
			in:   "IsDigit",
			want: "is_digit",
		},
		{
			in:   "Is",
			want: "is",
		},
		{
			in:   "IsID",
			want: "is_id",
		},
		{
			in:   "IsSQL",
			want: "is_sql",
		},
		{
			in:   "LongSQL",
			want: "long_sql",
		},
		{
			in:   "Float64Val",
			want: "float64_val",
		},
		{
			in:   "XMLName",
			want: "xml_name",
		},
	} {
		assert.Equal(t, test.want, camelCaseToSnakeCase(test.in))
	}
}

func TestStructMap(t *testing.T) {
	for _, test := range []struct {
		in  interface{}
		ok  []string
		bad []string
	}{
		{
			in: struct {
				CreatedAt time.Time
			}{},
			ok: []string{"created_at"},
		},
		{
			in: struct {
				intVal int
			}{},
			bad: []string{"int_val"},
		},
		{
			in: struct {
				IntVal int `db:"test"`
			}{},
			ok:  []string{"test"},
			bad: []string{"int_val"},
		},
		{
			in: struct {
				IntVal int `db:"-"`
			}{},
			bad: []string{"int_val"},
		},
		{
			in: struct {
				Test1 struct {
					Test2 int
				}
			}{},
			ok: []string{"test2"},
		},
	} {
		m := structMap(reflect.ValueOf(test.in), false)
		for _, c := range test.ok {
			_, ok := m[c]
			assert.True(t, ok)
		}
		for _, c := range test.bad {
			_, ok := m[c]
			assert.False(t, ok)
		}
	}
}
