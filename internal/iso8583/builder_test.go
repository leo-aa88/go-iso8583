package iso8583_test

import (
	"errors"
	"testing"

	"github.com/leo-aa88/go-iso8583/internal/iso8583"
)

func TestBuilder_NilSpec(t *testing.T) {
	_, err := iso8583.Build(nil, iso8583.NewMessage("0200"))
	if err == nil {
		t.Error("expected error for nil spec")
	}
}

func TestBuilder_InvalidMTILength(t *testing.T) {
	cases := []struct {
		name string
		mti  string
	}{
		{"too short", "020"},
		{"too long", "02000"},
		{"empty", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := iso8583.Build(stdSpec(), iso8583.NewMessage(tc.mti))
			if err == nil {
				t.Errorf("expected error for MTI %q", tc.mti)
			}
		})
	}
}

func TestBuilder_UnknownField(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		2: numericFixed(2),
	})

	// Field 5 is not defined in the spec.
	msg := buildMsg("0200", map[int]string{2: "42", 5: "99"})
	_, err := iso8583.Build(spec, msg)
	if err == nil {
		t.Error("expected error for field not in spec")
	}
	var fe *iso8583.FieldError
	if !errors.As(err, &fe) || fe.Field != 5 {
		t.Errorf("expected FieldError for field 5, got %v", err)
	}
}

func TestBuilder_FieldLengthExceedsMax(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		3: numericFixed(6),
	})

	msg := buildMsg("0200", map[int]string{3: "1234567"}) // 7 digits, MaxLength is 6.
	_, err := iso8583.Build(spec, msg)
	if err == nil {
		t.Error("expected error for value exceeding MaxLength")
	}
	var fe *iso8583.FieldError
	if !errors.As(err, &fe) || fe.Field != 3 {
		t.Errorf("expected FieldError for field 3, got %v", err)
	}
}

func TestBuilder_InvalidContentNumericField(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		3: numericFixed(6),
	})

	msg := buildMsg("0200", map[int]string{3: "ABC123"})
	_, err := iso8583.Build(spec, msg)
	if err == nil {
		t.Error("expected error for alpha characters in numeric field")
	}
	var fe *iso8583.FieldError
	if !errors.As(err, &fe) || fe.Field != 3 {
		t.Errorf("expected FieldError for field 3, got %v", err)
	}
}

func TestBuilder_PaddingLeft(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		3: numericFixed(6),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{3: "42"}))
	// Field starts at offset 12 (4 MTI + 8 bitmap).
	if got := string(wire[12:18]); got != "000042" {
		t.Errorf("left-padded field: want %q got %q", "000042", got)
	}
}

func TestBuilder_PaddingRight(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		39: alphaNumericFixed(6),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{39: "AB"}))
	// Field starts at offset 12 (4 MTI + 8 bitmap).
	if got := string(wire[12:18]); got != "AB    " {
		t.Errorf("right-padded field: want %q got %q", "AB    ", got)
	}
}

func TestBuilder_LLVARLengthPrefix(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		2: numericLLVAR(19),
	})

	cases := []struct {
		value          string
		expectedPrefix string
	}{
		{"4", "01"},
		{"4111111111111111", "16"},
		{"1234567890123456789", "19"},
	}
	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			wire := mustBuild(t, spec, buildMsg("0200", map[int]string{2: tc.value}))
			if got := string(wire[12:14]); got != tc.expectedPrefix {
				t.Errorf("LLVAR prefix: want %q got %q", tc.expectedPrefix, got)
			}
		})
	}
}

func TestBuilder_LLLVARLengthPrefix(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		48: alphaNumericLLLVAR(999),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{48: "Hello ISO8583"}))
	// LLLVAR prefix = 3 digits; "Hello ISO8583" is 13 chars → "013".
	if got := string(wire[12:15]); got != "013" {
		t.Errorf("LLLVAR prefix: want %q got %q", "013", got)
	}
}

func TestBuilder_BCDEncoding(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: {LengthType: iso8583.Fixed, MaxLength: 4, ContentType: iso8583.Numeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadLeft, PadChar: '0'},
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{4: "1234"}))
	// BCD "1234" → [0x12, 0x34].
	if wire[12] != 0x12 || wire[13] != 0x34 {
		t.Errorf("BCD encoding: want [0x12 0x34] got [0x%02X 0x%02X]", wire[12], wire[13])
	}
}

func TestBuilder_BCDInvalidNonDigit(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		4: {LengthType: iso8583.Fixed, MaxLength: 4, ContentType: iso8583.Numeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadLeft, PadChar: '0'},
	})

	_, err := iso8583.Build(spec, buildMsg("0200", map[int]string{4: "12X4"}))
	if err == nil {
		t.Error("expected error for non-digit value with BCD encoding")
	}
}

func TestBuilder_BCDRequiresNumericContentType(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		43: {LengthType: iso8583.Fixed, MaxLength: 4, ContentType: iso8583.AlphaNumeric, Encoding: iso8583.BCD, PadDirection: iso8583.PadRight, PadChar: ' '},
	})

	_, err := iso8583.Build(spec, buildMsg("0200", map[int]string{43: "ABCD"}))
	if err == nil {
		t.Error("expected error for BCD encoding with non-Numeric content type")
	}
}

func TestBuilder_AlphaField_Valid(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		43: {LengthType: iso8583.Fixed, MaxLength: 8, ContentType: iso8583.Alpha, Encoding: iso8583.ASCII, PadDirection: iso8583.PadRight, PadChar: ' '},
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{43: "MERCHANT"}))
	parsed := mustParse(t, spec, wire)
	v, _ := parsed.Get(43)
	if v != "MERCHANT" {
		t.Errorf("alpha field round-trip: want %q got %q", "MERCHANT", v)
	}
}

func TestBuilder_AlphaField_InvalidDigit(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		43: {LengthType: iso8583.Fixed, MaxLength: 8, ContentType: iso8583.Alpha, Encoding: iso8583.ASCII, PadDirection: iso8583.PadRight, PadChar: ' '},
	})

	_, err := iso8583.Build(spec, buildMsg("0200", map[int]string{43: "MERCH4NT"}))
	if err == nil {
		t.Error("expected error for digit in alpha field")
	}
}
