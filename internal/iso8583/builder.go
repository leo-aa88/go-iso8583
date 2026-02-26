package iso8583

import (
	"fmt"
	"sort"
)

// Build encodes a Message into a raw ISO8583 byte slice using the provided Spec.
//
// All fields present in msg.Fields are validated against their FieldSpec, padded,
// encoded, and written in numeric order.  The bitmap(s) are generated automatically.
//
// Returns the wire bytes on success, or a descriptive error on failure.
func Build(spec *Spec, msg *Message) ([]byte, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec must not be nil")
	}
	if len(msg.MTI) != 4 {
		return nil, fmt.Errorf("MTI must be exactly 4 characters, got %d", len(msg.MTI))
	}

	// Validate all fields before touching any output buffer.
	for fieldNum, value := range msg.Fields {
		if fieldNum < 2 || fieldNum > 128 {
			return nil, fieldErr(fieldNum, "field number out of range [2,128]")
		}
		fs, defined := spec.Fields[fieldNum]
		if !defined {
			return nil, ErrUnknownField(fieldNum)
		}
		if err := validateContent(fieldNum, value, fs.ContentType); err != nil {
			return nil, err
		}
		if len(value) > fs.MaxLength {
			return nil, fieldErr(fieldNum, "value length %d exceeds MaxLength %d", len(value), fs.MaxLength)
		}
	}

	// Build field bytes in sorted order so encoding is deterministic.
	fieldNums := make([]int, 0, len(msg.Fields))
	for k := range msg.Fields {
		fieldNums = append(fieldNums, k)
	}
	sort.Ints(fieldNums)

	// Pre-encode all field values to calculate sizes before writing.
	type encodedField struct {
		prefix []byte // nil for Fixed fields
		body   []byte
	}
	encoded := make(map[int]encodedField, len(fieldNums))

	for _, fieldNum := range fieldNums {
		value := msg.Fields[fieldNum]
		fs := spec.Fields[fieldNum]

		var ef encodedField

		switch fs.LengthType {
		case Fixed:
			padded := applyPadding(value, fs.MaxLength, fs)
			body, err := encodeField(fieldNum, padded, fs)
			if err != nil {
				return nil, err
			}
			ef.body = body

		case LLVAR, LLLVAR:
			// Variable fields: encode value as-is (no padding), prefix with length.
			body, err := encodeField(fieldNum, value, fs)
			if err != nil {
				return nil, err
			}
			// The length prefix represents the number of content characters, not wire bytes.
			prefix, err := formatLengthPrefix(fieldNum, len(value), fs.LengthType)
			if err != nil {
				return nil, err
			}
			ef.prefix = prefix
			ef.body = body

		default:
			return nil, fieldErr(fieldNum, "unknown LengthType %d", fs.LengthType)
		}

		encoded[fieldNum] = ef
	}

	// Generate bitmaps.
	bitmapBytes := buildBitmaps(msg.Fields)

	// Assemble output: MTI + bitmaps + fields.
	out := make([]byte, 0, 4+len(bitmapBytes)+256)
	out = append(out, []byte(msg.MTI)...)
	out = append(out, bitmapBytes...)

	for _, fieldNum := range fieldNums {
		ef := encoded[fieldNum]
		if ef.prefix != nil {
			out = append(out, ef.prefix...)
		}
		out = append(out, ef.body...)
	}

	return out, nil
}
