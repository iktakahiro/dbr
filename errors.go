package fjord

import "errors"

// package errors
var (
	ErrNotFound           = errors.New("fjord: not found")
	ErrNotSupported       = errors.New("fjord: not supported")
	ErrTableNotSpecified  = errors.New("fjord: table not specified")
	ErrColumnNotSpecified = errors.New("fjord: column not specified")
	ErrInvalidPointer     = errors.New("fjord: attempt to load into an invalid pointer")
	ErrPlaceholderCount   = errors.New("fjord: wrong placeholder count")
	ErrInvalidSliceLength = errors.New("fjord: length of slice is 0. length must be >= 1")
	ErrCantConvertToTime  = errors.New("fjord: can't convert to time.Time")
	ErrInvalidTimestring  = errors.New("fjord: invalid time string")
)
