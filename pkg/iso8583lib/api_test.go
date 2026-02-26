package iso8583lib_test

import (
	"bytes"
	"errors"
	"testing"

	lib "github.com/leo-aa88/go-iso8583/pkg/iso8583lib"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func mustBuild(t *testing.T, spec *lib.Spec, msg *lib.Message) []byte {
	t.Helper()
	wire, err := lib.Build(spec, msg)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return wire
}

func mustParse(t *testing.T, spec *lib.Spec, wire []byte) *lib.Message {
	t.Helper()
	msg, err := lib.Parse(spec, wire)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return msg
}

func newMsg(mti string, fields map[int]string) *lib.Message {
	m := lib.NewMessage(mti)
	for f, v := range fields {
		m.Set(f, v)
	}
	return m
}

// ── NewISO87AsciiSpec ─────────────────────────────────────────────────────────

func TestNewISO87AsciiSpec_NotNil(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	if spec == nil {
		t.Fatal("NewISO87AsciiSpec returned nil")
	}
	if len(spec.Fields) == 0 {
		t.Fatal("NewISO87AsciiSpec returned spec with no fields")
	}
}

// ── NewMessage ────────────────────────────────────────────────────────────────

func TestNewMessage_MTI(t *testing.T) {
	m := lib.NewMessage("0200")
	if m.MTI != "0200" {
		t.Errorf("want MTI %q got %q", "0200", m.MTI)
	}
}

func TestNewMessage_FieldsNotNil(t *testing.T) {
	m := lib.NewMessage("0200")
	if m.Fields == nil {
		t.Error("Fields map should be initialised")
	}
}

// ── Build ─────────────────────────────────────────────────────────────────────

func TestBuild_ReturnsBytes(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	msg := newMsg("0200", map[int]string{
		3: "000000",
		4: "000000010000",
	})
	wire := mustBuild(t, spec, msg)
	if len(wire) == 0 {
		t.Error("Build returned empty byte slice")
	}
}

func TestBuild_MTIAtStart(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	wire := mustBuild(t, spec, newMsg("0200", map[int]string{3: "000000"}))
	if got := string(wire[:4]); got != "0200" {
		t.Errorf("MTI at wire[0:4]: want %q got %q", "0200", got)
	}
}

func TestBuild_NilSpec_ReturnsError(t *testing.T) {
	_, err := lib.Build(nil, lib.NewMessage("0200"))
	if err == nil {
		t.Error("expected error for nil spec")
	}
}

func TestBuild_InvalidMTI_ReturnsError(t *testing.T) {
	_, err := lib.Build(lib.NewISO87AsciiSpec(), lib.NewMessage("020"))
	if err == nil {
		t.Error("expected error for 3-character MTI")
	}
}

func TestBuild_UnknownField_ReturnsFieldError(t *testing.T) {
	spec := &lib.Spec{Fields: map[int]lib.FieldSpec{
		2: {LengthType: lib.Fixed, MaxLength: 2, ContentType: lib.Numeric, Encoding: lib.ASCII, PadDirection: lib.PadLeft, PadChar: '0'},
	}}
	msg := newMsg("0200", map[int]string{2: "42", 9: "01"}) // field 9 not in spec
	_, err := lib.Build(spec, msg)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
	var fe *lib.FieldError
	if !errors.As(err, &fe) || fe.Field != 9 {
		t.Errorf("expected FieldError for field 9, got %v", err)
	}
}

func TestBuild_InvalidContent_ReturnsFieldError(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	msg := newMsg("0200", map[int]string{4: "ABCDEFGHIJKL"}) // field 4 is Numeric
	_, err := lib.Build(spec, msg)
	if err == nil {
		t.Fatal("expected error for alpha value in numeric field")
	}
	var fe *lib.FieldError
	if !errors.As(err, &fe) || fe.Field != 4 {
		t.Errorf("expected FieldError for field 4, got %v", err)
	}
}

// ── Parse ─────────────────────────────────────────────────────────────────────

func TestParse_NilSpec_ReturnsError(t *testing.T) {
	_, err := lib.Parse(nil, []byte("0200000000000000"))
	if err == nil {
		t.Error("expected error for nil spec")
	}
}

func TestParse_TruncatedMTI_ReturnsError(t *testing.T) {
	_, err := lib.Parse(lib.NewISO87AsciiSpec(), []byte("020"))
	if err == nil {
		t.Error("expected error for truncated MTI")
	}
}

func TestParse_TruncatedBitmap_ReturnsError(t *testing.T) {
	data := append([]byte("0200"), make([]byte, 4)...) // only 4 bitmap bytes, need 8
	_, err := lib.Parse(lib.NewISO87AsciiSpec(), data)
	if err == nil {
		t.Error("expected error for truncated bitmap")
	}
}

func TestParse_FieldError_CarriesFieldNumber(t *testing.T) {
	spec := &lib.Spec{Fields: map[int]lib.FieldSpec{
		3: {LengthType: lib.Fixed, MaxLength: 6, ContentType: lib.Numeric, Encoding: lib.ASCII, PadDirection: lib.PadLeft, PadChar: '0'},
	}}
	// Build a valid wire then truncate the field value.
	wire := mustBuild(t, spec, newMsg("0200", map[int]string{3: "123456"}))
	_, err := lib.Parse(spec, wire[:len(wire)-3])
	if err == nil {
		t.Fatal("expected parse error for truncated field")
	}
	var fe *lib.FieldError
	if !errors.As(err, &fe) || fe.Field != 3 {
		t.Errorf("expected FieldError for field 3, got %v", err)
	}
}

// ── Round-trip ────────────────────────────────────────────────────────────────

func TestRoundTrip_AuthRequest(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	original := newMsg("0100", map[int]string{
		2:  "4111111111111111",
		3:  "000000",
		4:  "000000050000",
		7:  "0201153000",
		11: "000042",
		12: "153000",
		13: "0201",
		22: "051",
		25: "00",
		49: "978",
	})

	wire := mustBuild(t, spec, original)
	parsed := mustParse(t, spec, wire)

	if parsed.MTI != "0100" {
		t.Errorf("MTI: want %q got %q", "0100", parsed.MTI)
	}
	for _, f := range []int{2, 3, 4, 7, 11, 12, 13, 22, 25, 49} {
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

func TestRoundTrip_AuthResponse(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	original := newMsg("0110", map[int]string{
		3:  "000000",
		4:  "000000050000",
		11: "000042",
		38: "123456",
		39: "00",
	})

	wire := mustBuild(t, spec, original)
	parsed := mustParse(t, spec, wire)

	if parsed.MTI != "0110" {
		t.Errorf("MTI: want %q got %q", "0110", parsed.MTI)
	}
	for _, f := range []int{3, 4, 11, 38, 39} {
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

func TestRoundTrip_NetworkManagement_SecondaryBitmap(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	original := newMsg("0800", map[int]string{
		11: "000001",
		70: "301",
	})

	wire := mustBuild(t, spec, original)

	// Bit 1 of primary bitmap must be set when secondary fields are present.
	if wire[4]&0x80 == 0 {
		t.Error("bit 1 not set; secondary bitmap missing")
	}

	parsed := mustParse(t, spec, wire)

	for _, f := range []int{11, 70} {
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

func TestRoundTrip_LLLVARField(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	original := newMsg("0200", map[int]string{
		11: "000001",
		48: "ADDITIONAL DATA FIELD WITH SPACES AND NUMBERS 1234",
	})

	wire := mustBuild(t, spec, original)
	parsed := mustParse(t, spec, wire)

	got, ok := parsed.Get(48)
	if !ok {
		t.Fatal("field 48: missing after parse")
	}
	want, _ := original.Get(48)
	if got != want {
		t.Errorf("field 48: want %q got %q", want, got)
	}
}

func TestRoundTrip_ParseThenBuild_ByteIdentical(t *testing.T) {
	spec := lib.NewISO87AsciiSpec()
	original := newMsg("0200", map[int]string{
		2:  "4111111111111111",
		3:  "000000",
		4:  "000000010000",
		11: "000001",
		39: "00",
	})

	wire1 := mustBuild(t, spec, original)
	parsed := mustParse(t, spec, wire1)
	wire2 := mustBuild(t, spec, parsed)

	if !bytes.Equal(wire1, wire2) {
		t.Errorf("parse→build not byte-identical\nwire1: %X\nwire2: %X", wire1, wire2)
	}
}

func TestRoundTrip_CustomSpec(t *testing.T) {
	spec := &lib.Spec{
		Fields: map[int]lib.FieldSpec{
			2:  {LengthType: lib.LLVAR, MaxLength: 19, ContentType: lib.Numeric, Encoding: lib.ASCII, PadChar: '0'},
			3:  {LengthType: lib.Fixed, MaxLength: 6, ContentType: lib.Numeric, Encoding: lib.ASCII, PadDirection: lib.PadLeft, PadChar: '0'},
			39: {LengthType: lib.Fixed, MaxLength: 2, ContentType: lib.AlphaNumeric, Encoding: lib.ASCII, PadDirection: lib.PadRight, PadChar: ' '},
		},
	}

	original := newMsg("0210", map[int]string{
		2:  "4111111111111111",
		3:  "000000",
		39: "00",
	})

	wire := mustBuild(t, spec, original)
	parsed := mustParse(t, spec, wire)

	for _, f := range []int{2, 3, 39} {
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
