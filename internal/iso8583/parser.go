package iso8583

import (
	"fmt"
	"sort"
	"strconv"
)

// Parse decodes a raw ISO8583 byte slice into a Message using the provided Spec.
//
// Layout on the wire:
//
//	[MTI 4 bytes][Primary Bitmap 8 bytes][Secondary Bitmap 8 bytes, optional][Fields...]
//
// Returns a *Message on success, or a descriptive error (often *FieldError) on failure.
func Parse(spec *Spec, data []byte) (*Message, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec must not be nil")
	}

	pos := 0

	// --- MTI ---
	if len(data) < 4 {
		return nil, ErrTruncatedMTI
	}
	mti := string(data[pos : pos+4])
	pos += 4

	// --- Bitmaps ---
	bitmap, bitmapLen, err := parseBitmaps(data, pos)
	if err != nil {
		return nil, err
	}
	pos += bitmapLen

	msg := &Message{
		MTI:    mti,
		Fields: make(map[int][]byte),
	}

	// --- Fields ---
	// Iterate fields in numeric order (skip field 1 = secondary bitmap indicator).
	fieldNums := make([]int, 0, 128)
	for i := 2; i <= 128; i++ {
		if bitmap[i-1] {
			fieldNums = append(fieldNums, i)
		}
	}
	sort.Ints(fieldNums)

	for _, fieldNum := range fieldNums {
		fs, defined := spec.Fields[fieldNum]
		if !defined {
			return nil, &FieldError{Field: fieldNum, Err: fmt.Errorf("field present in bitmap but not defined in spec")}
		}

		var wireBytes []byte
		var wireLen int

		switch fs.LengthType {
		case Fixed:
			wl, err := encodedLength(fieldNum, fs.MaxLength, fs)
			if err != nil {
				return nil, err
			}
			if pos+wl > len(data) {
				return nil, fieldErr(fieldNum, "insufficient data: need %d bytes, have %d", wl, len(data)-pos)
			}
			wireBytes = data[pos : pos+wl]
			wireLen = wl

		case LLVAR:
			prefixLen := 2
			if pos+prefixLen > len(data) {
				return nil, fieldErr(fieldNum, "insufficient data for LLVAR length prefix")
			}
			n, err := strconv.Atoi(string(data[pos : pos+prefixLen]))
			if err != nil || n < 0 {
				return nil, fieldErr(fieldNum, "invalid LLVAR length prefix %q", string(data[pos:pos+prefixLen]))
			}
			if n > fs.MaxLength {
				return nil, fieldErr(fieldNum, "LLVAR declared length %d exceeds MaxLength %d", n, fs.MaxLength)
			}
			pos += prefixLen
			wl, err := encodedLength(fieldNum, n, fs)
			if err != nil {
				return nil, err
			}
			if pos+wl > len(data) {
				return nil, fieldErr(fieldNum, "insufficient data for LLVAR value: need %d bytes, have %d", wl, len(data)-pos)
			}
			wireBytes = data[pos : pos+wl]
			wireLen = wl

		case LLLVAR:
			prefixLen := 3
			if pos+prefixLen > len(data) {
				return nil, fieldErr(fieldNum, "insufficient data for LLLVAR length prefix")
			}
			n, err := strconv.Atoi(string(data[pos : pos+prefixLen]))
			if err != nil || n < 0 {
				return nil, fieldErr(fieldNum, "invalid LLLVAR length prefix %q", string(data[pos:pos+prefixLen]))
			}
			if n > fs.MaxLength {
				return nil, fieldErr(fieldNum, "LLLVAR declared length %d exceeds MaxLength %d", n, fs.MaxLength)
			}
			pos += prefixLen
			wl, err := encodedLength(fieldNum, n, fs)
			if err != nil {
				return nil, err
			}
			if pos+wl > len(data) {
				return nil, fieldErr(fieldNum, "insufficient data for LLLVAR value: need %d bytes, have %d", wl, len(data)-pos)
			}
			wireBytes = data[pos : pos+wl]
			wireLen = wl

		default:
			return nil, fieldErr(fieldNum, "unknown LengthType %d", fs.LengthType)
		}

		decoded, err := decodeField(fieldNum, wireBytes, fs)
		if err != nil {
			return nil, err
		}

		if err := validateContent(fieldNum, decoded, fs.ContentType); err != nil {
			return nil, err
		}

		msg.Fields[fieldNum] = decoded
		pos += wireLen
	}

	return msg, nil
}
