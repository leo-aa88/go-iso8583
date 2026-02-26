package iso8583

import "fmt"

// FieldError wraps a lower-level error with the field number that caused it.
// All parsing and building errors are returned as *FieldError so callers can
// identify exactly which field failed and why.
type FieldError struct {
	Field int
	Err   error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("field %d: %v", e.Field, e.Err)
}

func (e *FieldError) Unwrap() error {
	return e.Err
}

// fieldErr is a convenience constructor used throughout the package.
func fieldErr(field int, format string, args ...any) error {
	return &FieldError{Field: field, Err: fmt.Errorf(format, args...)}
}

// ErrTruncatedMTI is returned when there are fewer than 4 bytes available for the MTI.
var ErrTruncatedMTI = fmt.Errorf("truncated MTI: need 4 bytes")

// ErrTruncatedBitmap is returned when the bitmap bytes are not fully available.
var ErrTruncatedBitmap = fmt.Errorf("truncated bitmap")

// ErrUnknownField is returned during Build when a field is present in the message
// but has no entry in the Spec.
func ErrUnknownField(field int) error {
	return &FieldError{Field: field, Err: fmt.Errorf("not defined in spec")}
}
