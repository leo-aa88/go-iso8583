package iso8583

// LengthType defines how a field's length is encoded in the message.
type LengthType int

const (
	// Fixed means the field always occupies exactly MaxLength bytes.
	Fixed LengthType = iota
	// LLVAR means the field is prefixed with a 2-digit ASCII decimal length.
	LLVAR
	// LLLVAR means the field is prefixed with a 3-digit ASCII decimal length.
	LLLVAR
)

// ContentType defines what kind of data is valid for a field.
type ContentType int

const (
	Numeric      ContentType = iota // Digits 0-9 only.
	Alpha                           // Letters and spaces only.
	AlphaNumeric                    // Letters, digits, and special characters.
	Binary                          // Raw bytes; no content validation applied.
)

// EncodingType defines how the field value bytes are encoded on the wire.
type EncodingType int

const (
	// ASCII means each character is one byte.
	ASCII EncodingType = iota
	// BCD (Binary-Coded Decimal) packs two decimal digits per byte.
	// Used for Numeric fields; output length is ceil(n/2).
	BCD
)

// PadDirection indicates which side padding is applied to reach MaxLength.
type PadDirection int

const (
	PadLeft  PadDirection = iota // Pad on the left (e.g., numeric fields: "42" → "0042").
	PadRight                     // Pad on the right (e.g., alpha fields: "AB" → "AB  ").
)

// FieldSpec describes how a single ISO8583 field is parsed and built.
type FieldSpec struct {
	// LengthType determines how the field length is communicated in the wire format.
	LengthType LengthType
	// MaxLength is the maximum number of content characters/digits this field may hold.
	// For Fixed fields it is also the exact length.
	MaxLength int
	// ContentType constrains the allowed character set of the field value.
	ContentType ContentType
	// Encoding controls the wire encoding of the field value bytes.
	Encoding EncodingType
	// PadDirection is the side on which padding is applied for Fixed fields.
	PadDirection PadDirection
	// PadChar is the byte used for padding (e.g., '0' or ' ').
	PadChar byte
}

// Spec is the top-level specification that drives all parsing and building.
// Engineers create one Spec per ISO8583 variant (e.g., ISO87, ISO93, proprietary).
type Spec struct {
	// Fields maps field numbers (1-128) to their specifications.
	// Field 1 is the secondary bitmap and is handled automatically; do not add it.
	Fields map[int]FieldSpec
}
