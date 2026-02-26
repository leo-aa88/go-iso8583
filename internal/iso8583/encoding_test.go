package iso8583_test

import (
	"testing"

	"github.com/leo-aa88/go-iso8583/internal/iso8583"
)

func TestEncoding_ASCII_WritesLiteralBytes(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: numericFixed(4),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{4: "1234"}))
	// Field starts at offset 12; ASCII means raw digit bytes.
	if got := string(wire[12:16]); got != "1234" {
		t.Errorf("ASCII encoding: want %q got %q", "1234", got)
	}
}

func TestEncoding_BCD_EvenLength(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: {LengthType: iso8583.Fixed, MaxLength: 4, ContentType: iso8583.Numeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadLeft, PadChar: '0'},
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{4: "1234"}))
	// "1234" → [0x12, 0x34] (2 wire bytes for 4 digits).
	if wire[12] != 0x12 || wire[13] != 0x34 {
		t.Errorf("BCD even-length: want [0x12 0x34] got [0x%02X 0x%02X]", wire[12], wire[13])
	}
}

func TestEncoding_BCD_OddLength(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: {LengthType: iso8583.Fixed, MaxLength: 3, ContentType: iso8583.Numeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadLeft, PadChar: '0'},
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{4: "123"}))
	// Odd digits: "123" left-padded to "0123" → [0x01, 0x23].
	if wire[12] != 0x01 || wire[13] != 0x23 {
		t.Errorf("BCD odd-length: want [0x01 0x23] got [0x%02X 0x%02X]", wire[12], wire[13])
	}
}

func TestEncoding_BCD_RoundTrip(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: {LengthType: iso8583.Fixed, MaxLength: 4, ContentType: iso8583.Numeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadLeft, PadChar: '0'},
	})

	cases := []string{"1234", "0001", "9999"}
	for _, value := range cases {
		t.Run(value, func(t *testing.T) {
			wire := mustBuild(t, spec, buildMsg("0200", map[int]string{4: value}))
			parsed := mustParse(t, spec, wire)
			v, _ := parsed.Get(4)
			if v != value {
				t.Errorf("BCD round-trip: want %q got %q", value, v)
			}
		})
	}
}

func TestEncoding_BCD_OddLengthRoundTrip(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: {LengthType: iso8583.Fixed, MaxLength: 3, ContentType: iso8583.Numeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadLeft, PadChar: '0'},
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{4: "123"}))
	parsed := mustParse(t, spec, wire)
	v, _ := parsed.Get(4)
	// Leading zero nibble added during encode must be stripped on decode.
	if v != "123" {
		t.Errorf("BCD odd round-trip: want %q got %q", "123", v)
	}
}

func TestEncoding_BCD_InvalidNonDigit(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: {LengthType: iso8583.Fixed, MaxLength: 4, ContentType: iso8583.Numeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadLeft, PadChar: '0'},
	})

	_, err := iso8583.Build(spec, buildMsg("0200", map[int]string{4: "12X4"}))
	if err == nil {
		t.Error("expected error for non-digit character with BCD encoding")
	}
}

func TestEncoding_BCD_OnlyValidForNumericContentType(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		43: {LengthType: iso8583.Fixed, MaxLength: 4, ContentType: iso8583.AlphaNumeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadRight, PadChar: ' '},
	})

	_, err := iso8583.Build(spec, buildMsg("0200", map[int]string{43: "ABCD"}))
	if err == nil {
		t.Error("expected error for BCD encoding combined with non-Numeric content type")
	}
}

func TestEncoding_ContentType_Numeric_RejectsAlpha(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: numericFixed(12),
	})

	_, err := iso8583.Build(spec, buildMsg("0200", map[int]string{4: "ABCDEFGHIJKL"}))
	if err == nil {
		t.Error("expected error for alpha value in Numeric field")
	}
}

func TestEncoding_ContentType_Alpha_RejectsDigits(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		43: {LengthType: iso8583.Fixed, MaxLength: 8, ContentType: iso8583.Alpha, Encoding: iso8583.ASCII, PadDirection: iso8583.PadRight, PadChar: ' '},
	})

	_, err := iso8583.Build(spec, buildMsg("0200", map[int]string{43: "MERCH4NT"}))
	if err == nil {
		t.Error("expected error for digit character in Alpha field")
	}
}

func TestEncoding_ContentType_AlphaNumeric_AcceptsLettersAndDigits(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		37: alphaNumericFixed(12),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{37: "REF123456789"}))
	parsed := mustParse(t, spec, wire)
	v, _ := parsed.Get(37)
	if v != "REF123456789" {
		t.Errorf("alphanumeric round-trip: want %q got %q", "REF123456789", v)
	}
}

func TestEncoding_ContentType_Binary_AcceptsAnyByte(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		55: {LengthType: iso8583.LLLVAR, MaxLength: 10, ContentType: iso8583.Binary, Encoding: iso8583.ASCII, PadChar: 0x00},
	})

	msg := iso8583.NewMessage("0200")
	msg.SetBytes(55, []byte{0x00, 0x01, 0xFF, 0xFE})
	wire := mustBuild(t, spec, msg)
	parsed := mustParse(t, spec, wire)
	b, _ := parsed.GetBytes(55)
	if len(b) != 4 || b[0] != 0x00 || b[2] != 0xFF {
		t.Errorf("binary round-trip: unexpected bytes %v", b)
	}
}
