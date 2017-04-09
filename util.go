package fjord

import (
	"bytes"
	"database/sql/driver"
	"reflect"
	"strings"
	"unicode"
)

// r is *Replacer for columnNameToAlias function.
var r = strings.NewReplacer(".", "__")

// columnNameToAlias converts a column name into alias name.
// e.g. "u.id" ==> "u__id"
func columnNameToAlias(column string) string {
	return r.Replace(column)
}

// camelCaseToSnakeCase converts a camelCaseString to a snake_case_string.
func camelCaseToSnakeCase(name string) string {
	buf := new(bytes.Buffer)

	runes := []rune(name)
	for i := 0; i < len(runes); i++ {
		buf.WriteRune(unicode.ToLower(runes[i]))
		if i != len(runes)-1 && unicode.IsUpper(runes[i+1]) &&
			(unicode.IsLower(runes[i]) || unicode.IsDigit(runes[i]) ||
				(i != len(runes)-2 && unicode.IsLower(runes[i+2]))) {
			buf.WriteRune('_')
		}
	}

	return buf.String()
}

// getColumnNameFromTag get a value from the db tag in a Struct field.
func getColumnNameFromTag(field reflect.StructField, ignorePrefix bool) (column string) {
	tag := field.Tag.Get("db")
	if tag == "-" {
		// Ignore the field that "-" tag is set.
		return ""
	}
	if tag == "" {
		// Convert based on default rules
		return camelCaseToSnakeCase(field.Name)
	}
	if strings.Contains(tag, ".") {
		if ignorePrefix {
			return strings.Split(tag, ".")[1]
		}
		return columnNameToAlias(tag)
	}

	return tag
}

func structMap(value reflect.Value, ignorePrefix bool) map[string]reflect.Value {
	m := make(map[string]reflect.Value)
	structValue(m, value, ignorePrefix)

	return m
}

var (
	typeValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
)

func structValue(m map[string]reflect.Value, value reflect.Value, ignorePrefix bool) {
	if value.Type().Implements(typeValuer) {
		return
	}

	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return
		}
		structValue(m, value.Elem(), ignorePrefix)
	case reflect.Struct:
		t := value.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				// unexported
				continue
			}
			column := getColumnNameFromTag(field, ignorePrefix)
			if column == "" {
				continue
			}

			fieldValue := value.Field(i)
			if _, ok := m[column]; !ok {
				m[column] = fieldValue
			}
			structValue(m, fieldValue, ignorePrefix)
		}
	}
}
