package iso8583_test

import (
	"testing"

	"github.com/leo-aa88/go-iso8583/internal/iso8583"
)

// minSpec creates a Spec containing only the specified fields.
func minSpec(fields map[int]iso8583.FieldSpec) *iso8583.Spec {
	return &iso8583.Spec{Fields: fields}
}

// stdSpec returns the built-in ISO87 ASCII spec.
func stdSpec() *iso8583.Spec { return iso8583.NewISO87AsciiSpec() }

// buildMsg constructs a Message from a map of field→string values.
func buildMsg(mti string, fields map[int]string) *iso8583.Message {
	m := iso8583.NewMessage(mti)
	for k, v := range fields {
		m.Set(k, v)
	}
	return m
}

// mustBuild calls Build and fatally fails the test on error.
func mustBuild(t *testing.T, spec *iso8583.Spec, msg *iso8583.Message) []byte {
	t.Helper()
	b, err := iso8583.Build(spec, msg)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	return b
}

// mustParse calls Parse and fatally fails the test on error.
func mustParse(t *testing.T, spec *iso8583.Spec, data []byte) *iso8583.Message {
	t.Helper()
	m, err := iso8583.Parse(spec, data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	return m
}

// numericFixed is a convenience FieldSpec for a left-zero-padded ASCII numeric field.
func numericFixed(maxLen int) iso8583.FieldSpec {
	return iso8583.FieldSpec{
		LengthType:   iso8583.Fixed,
		MaxLength:    maxLen,
		ContentType:  iso8583.Numeric,
		Encoding:     iso8583.ASCII,
		PadDirection: iso8583.PadLeft,
		PadChar:      '0',
	}
}

// alphaNumericFixed is a convenience FieldSpec for a right-space-padded ASCII alphanumeric field.
func alphaNumericFixed(maxLen int) iso8583.FieldSpec {
	return iso8583.FieldSpec{
		LengthType:   iso8583.Fixed,
		MaxLength:    maxLen,
		ContentType:  iso8583.AlphaNumeric,
		Encoding:     iso8583.ASCII,
		PadDirection: iso8583.PadRight,
		PadChar:      ' ',
	}
}

// numericLLVAR is a convenience FieldSpec for an LLVAR ASCII numeric field.
func numericLLVAR(maxLen int) iso8583.FieldSpec {
	return iso8583.FieldSpec{
		LengthType:  iso8583.LLVAR,
		MaxLength:   maxLen,
		ContentType: iso8583.Numeric,
		Encoding:    iso8583.ASCII,
		PadChar:     '0',
	}
}

// alphaNumericLLLVAR is a convenience FieldSpec for an LLLVAR ASCII alphanumeric field.
func alphaNumericLLLVAR(maxLen int) iso8583.FieldSpec {
	return iso8583.FieldSpec{
		LengthType:  iso8583.LLLVAR,
		MaxLength:   maxLen,
		ContentType: iso8583.AlphaNumeric,
		Encoding:    iso8583.ASCII,
		PadChar:     ' ',
	}
}
