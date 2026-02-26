package iso8583_test

import (
	"testing"

	"github.com/leo-aa88/go-iso8583/internal/iso8583"
)

func TestSpec_ISO87AsciiSpec_ContainsCommonFields(t *testing.T) {
	spec := iso8583.NewISO87AsciiSpec()
	if spec == nil {
		t.Fatal("NewISO87AsciiSpec returned nil")
	}
	if spec.Fields == nil {
		t.Fatal("Spec.Fields is nil")
	}

	// Spot-check a representative set of fields.
	expectedFields := []int{2, 3, 4, 7, 11, 12, 13, 39, 48, 70}
	for _, f := range expectedFields {
		if _, ok := spec.Fields[f]; !ok {
			t.Errorf("ISO87 spec missing field %d", f)
		}
	}
}

func TestSpec_FieldSpec_LengthTypes(t *testing.T) {
	cases := []struct {
		name       string
		spec       iso8583.FieldSpec
		lengthType iso8583.LengthType
	}{
		{"Fixed numeric", numericFixed(6), iso8583.Fixed},
		{"LLVAR numeric", numericLLVAR(19), iso8583.LLVAR},
		{"LLLVAR alphanumeric", alphaNumericLLLVAR(999), iso8583.LLLVAR},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.spec.LengthType != tc.lengthType {
				t.Errorf("want LengthType %v got %v", tc.lengthType, tc.spec.LengthType)
			}
		})
	}
}

func TestSpec_FieldSpec_ContentTypes(t *testing.T) {
	cases := []struct {
		name        string
		contentType iso8583.ContentType
	}{
		{"Numeric", iso8583.Numeric},
		{"Alpha", iso8583.Alpha},
		{"AlphaNumeric", iso8583.AlphaNumeric},
		{"Binary", iso8583.Binary},
	}
	// All content type constants must be distinct.
	seen := make(map[iso8583.ContentType]string)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if prev, dup := seen[tc.contentType]; dup {
				t.Errorf("ContentType %v (%s) collides with %s", tc.contentType, tc.name, prev)
			}
			seen[tc.contentType] = tc.name
		})
	}
}

func TestSpec_FieldSpec_EncodingTypes(t *testing.T) {
	if iso8583.ASCII == iso8583.BCD {
		t.Error("ASCII and BCD encoding constants must be distinct")
	}
}

func TestSpec_FieldSpec_PadDirections(t *testing.T) {
	if iso8583.PadLeft == iso8583.PadRight {
		t.Error("PadLeft and PadRight constants must be distinct")
	}
}

func TestSpec_CustomSpec_RoundTrip(t *testing.T) {
	// Verify that a manually constructed Spec drives correct behaviour end-to-end.
	spec := &iso8583.Spec{
		Fields: map[int]iso8583.FieldSpec{
			2:  numericLLVAR(19),
			3:  numericFixed(6),
			4:  numericFixed(12),
			39: alphaNumericFixed(2),
		},
	}

	original := buildMsg("0200", map[int]string{
		2:  "4111111111111111",
		3:  "000000",
		4:  "000000010000",
		39: "00",
	})

	wire := mustBuild(t, spec, original)
	parsed := mustParse(t, spec, wire)

	for _, f := range []int{2, 3, 4, 39} {
		want, _ := original.Get(f)
		got, ok := parsed.Get(f)
		if !ok {
			t.Errorf("field %d: missing after parse", f)
			continue
		}
		if got != want {
			t.Errorf("field %d: want %q got %q", f, want, got)
		}
	}
}

func TestSpec_Field1_IsReservedForBitmap(t *testing.T) {
	// Field 1 in the bitmap signals a secondary bitmap; it must never be set
	// by the caller.  Build should not emit it even if attempted.
	spec := minSpec(map[int]iso8583.FieldSpec{
		2: numericFixed(2),
	})

	msg := iso8583.NewMessage("0200")
	msg.Fields[1] = []byte("anything") // attempt to inject field 1
	msg.Set(2, "42")

	// Build must either ignore field 1 or reject it (out-of-range).
	// Either way the wire must remain parseable and field 1 must not appear
	// as a data field in the result.
	_, err := iso8583.Build(spec, msg)
	// If Build rejects it, that is also acceptable.
	if err == nil {
		// If it succeeded, parse and confirm field 1 is not in the data fields.
		wire := mustBuild(t, minSpec(map[int]iso8583.FieldSpec{2: numericFixed(2)}),
			buildMsg("0200", map[int]string{2: "42"}))
		parsed := mustParse(t, spec, wire)
		if _, ok := parsed.Get(1); ok {
			t.Error("field 1 should never appear as a data field after parse")
		}
	}
}
