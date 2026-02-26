package iso8583

import (
	"fmt"
	"unicode"
)

// encodeField encodes a raw field value according to the FieldSpec.
// It returns the wire bytes to be written.
func encodeField(field int, value []byte, fs FieldSpec) ([]byte, error) {
	switch fs.Encoding {
	case ASCII:
		return value, nil
	case BCD:
		if fs.ContentType != Numeric {
			return nil, fieldErr(field, "BCD encoding is only valid for Numeric fields")
		}
		return encodeBCD(field, value)
	default:
		return nil, fieldErr(field, "unknown encoding %d", fs.Encoding)
	}
}

// decodeField decodes wire bytes for a field back to the canonical string form.
func decodeField(field int, wire []byte, fs FieldSpec) ([]byte, error) {
	switch fs.Encoding {
	case ASCII:
		return wire, nil
	case BCD:
		if fs.ContentType != Numeric {
			return nil, fieldErr(field, "BCD encoding is only valid for Numeric fields")
		}
		return decodeBCD(field, wire, fs.MaxLength)
	default:
		return nil, fieldErr(field, "unknown encoding %d", fs.Encoding)
	}
}

// encodedLength returns the number of wire bytes a value will occupy after encoding.
func encodedLength(field int, valueLen int, fs FieldSpec) (int, error) {
	switch fs.Encoding {
	case ASCII:
		return valueLen, nil
	case BCD:
		return (valueLen + 1) / 2, nil
	default:
		return 0, fieldErr(field, "unknown encoding %d", fs.Encoding)
	}
}

// encodeBCD converts a decimal-digit string (as []byte) to packed BCD bytes.
// If the number of digits is odd, a leading zero nibble is prepended.
func encodeBCD(field int, digits []byte) ([]byte, error) {
	for _, b := range digits {
		if b < '0' || b > '9' {
			return nil, fieldErr(field, "BCD encoding: non-digit character %q", rune(b))
		}
	}
	length := (len(digits) + 1) / 2
	out := make([]byte, length)
	src := digits
	// If odd number of digits, left-pad with '0'.
	if len(digits)%2 != 0 {
		src = make([]byte, len(digits)+1)
		src[0] = '0'
		copy(src[1:], digits)
	}
	for i := 0; i < length; i++ {
		hi := src[i*2] - '0'
		lo := src[i*2+1] - '0'
		out[i] = (hi << 4) | lo
	}
	return out, nil
}

// decodeBCD converts packed BCD bytes back to a decimal-digit string.
// maxLength is used to trim any leading zero that was added during odd-length encoding.
func decodeBCD(field int, bcdBytes []byte, maxLength int) ([]byte, error) {
	out := make([]byte, len(bcdBytes)*2)
	for i, b := range bcdBytes {
		hi := (b >> 4) & 0x0F
		lo := b & 0x0F
		if hi > 9 || lo > 9 {
			return nil, fieldErr(field, "BCD decoding: invalid nibble in byte at position %d", i)
		}
		out[i*2] = '0' + hi
		out[i*2+1] = '0' + lo
	}
	// Trim leading zero added for odd-length values.
	if len(out) > maxLength && out[0] == '0' {
		out = out[1:]
	}
	return out, nil
}

// validateContent checks that value bytes conform to the ContentType constraint.
func validateContent(field int, value []byte, ct ContentType) error {
	switch ct {
	case Numeric:
		for _, b := range value {
			if b < '0' || b > '9' {
				return fieldErr(field, "numeric content violation: character %q is not a digit", rune(b))
			}
		}
	case Alpha:
		for _, b := range value {
			if !unicode.IsLetter(rune(b)) && b != ' ' {
				return fieldErr(field, "alpha content violation: character %q is not a letter or space", rune(b))
			}
		}
	case AlphaNumeric:
		// No restriction beyond printable ASCII for wire safety.
		for _, b := range value {
			if b < 0x20 || b > 0x7E {
				return fieldErr(field, "alphanumeric content violation: non-printable byte 0x%02X", b)
			}
		}
	case Binary:
		// Any byte is valid.
	default:
		return fieldErr(field, "unknown content type %d", ct)
	}
	return nil
}

// applyPadding returns value padded to exactly length using the spec's pad settings.
func applyPadding(value []byte, length int, fs FieldSpec) []byte {
	if len(value) >= length {
		return value
	}
	pad := make([]byte, length-len(value))
	for i := range pad {
		pad[i] = fs.PadChar
	}
	if fs.PadDirection == PadLeft {
		return append(pad, value...)
	}
	return append(value, pad...)
}

// formatLengthPrefix returns the ASCII length prefix string for LLVAR/LLLVAR fields.
func formatLengthPrefix(field int, length int, lt LengthType) ([]byte, error) {
	switch lt {
	case LLVAR:
		if length > 99 {
			return nil, fieldErr(field, "LLVAR length %d exceeds maximum representable value 99", length)
		}
		return []byte(fmt.Sprintf("%02d", length)), nil
	case LLLVAR:
		if length > 999 {
			return nil, fieldErr(field, "LLLVAR length %d exceeds maximum representable value 999", length)
		}
		return []byte(fmt.Sprintf("%03d", length)), nil
	default:
		return nil, fieldErr(field, "formatLengthPrefix called for non-variable field")
	}
}
