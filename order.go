package fjord

type direction bool

// order by directions
// most databases by default use asc
const (
	asc  direction = false
	desc           = true
)

func order(column string, dir direction) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		// FIXME: no quote ident
		buf.WriteString(column)

		if dir {
			buf.WriteString(" DESC")
			return nil
		}
		buf.WriteString(" ASC")
		return nil
	})
}
