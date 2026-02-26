package iso8583_test

import (
	"errors"
	"testing"

	"github.com/leo-aa88/go-iso8583/internal/iso8583"
)

func TestBitmap_PrimaryOnly(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		2: numericFixed(2),
		3: numericFixed(2),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{2: "42", 3: "07"}))

	bitmapSection := wire[4:12]

	// Bit 1 (secondary indicator) must NOT be set — no field > 64.
	if bitmapSection[0]&0x80 != 0 {
		t.Error("bit 1 (secondary bitmap indicator) should not be set when no secondary fields present")
	}
	// Bit 2 → mask 0b01000000
	if bitmapSection[0]&0x40 == 0 {
		t.Error("bit 2 should be set in primary bitmap")
	}
	// Bit 3 → mask 0b00100000
	if bitmapSection[0]&0x20 == 0 {
		t.Error("bit 3 should be set in primary bitmap")
	}
	// Total: MTI(4) + bitmap(8) + field2(2) + field3(2) = 16
	if len(wire) != 16 {
		t.Errorf("expected wire length 16, got %d", len(wire))
	}
}

func TestBitmap_PrimaryAndSecondary(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		2:  numericFixed(2),
		70: numericFixed(3),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{2: "42", 70: "301"}))

	// Both bitmaps occupy wire[4:20].
	bitmapSection := wire[4:20]

	// Bit 1 set → secondary bitmap present.
	if bitmapSection[0]&0x80 == 0 {
		t.Error("bit 1 must be set to indicate secondary bitmap")
	}
	// Bit 2 set → field 2 present.
	if bitmapSection[0]&0x40 == 0 {
		t.Error("bit 2 must be set")
	}
	// Field 70: 0-indexed position 69; secondary position 69-64=5.
	// Secondary byte 0, bit from MSB: mask = 1<<(7-5) = 0x04.
	if bitmapSection[8]&0x04 == 0 {
		t.Error("bit 70 must be set in secondary bitmap byte 0")
	}
}

func TestBitmap_AutoSetBit1ForSecondaryField(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		70: numericFixed(3),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{70: "001"}))

	if wire[4]&0x80 == 0 {
		t.Error("Build must automatically set bit 1 when a secondary-bitmap field is present")
	}
}

func TestBitmap_TruncatedPrimary(t *testing.T) {
	// Only 6 bytes for the bitmap instead of the required 8.
	data := append([]byte("0200"), make([]byte, 6)...)
	_, err := iso8583.Parse(stdSpec(), data)
	if !errors.Is(err, iso8583.ErrTruncatedBitmap) {
		t.Errorf("expected ErrTruncatedBitmap, got %v", err)
	}
}

func TestBitmap_TruncatedSecondary(t *testing.T) {
	// Primary bitmap with bit 1 set but secondary bytes absent.
	data := []byte("0200")
	primary := make([]byte, 8)
	primary[0] = 0x80 // Bit 1 set → secondary required.
	data = append(data, primary...)

	spec := minSpec(map[int]iso8583.FieldSpec{
		70: numericFixed(3),
	})
	_, err := iso8583.Parse(spec, data)
	if !errors.Is(err, iso8583.ErrTruncatedBitmap) {
		t.Errorf("expected ErrTruncatedBitmap, got %v", err)
	}
}

func TestBitmap_RoundTrip(t *testing.T) {
	spec := stdSpec()
	original := buildMsg("0200", map[int]string{
		2:  "4111111111111111",
		3:  "000000",
		4:  "000000010000",
		70: "301",
	})

	wire := mustBuild(t, spec, original)
	parsed := mustParse(t, spec, wire)

	for _, f := range []int{2, 3, 4, 70} {
		ov, _ := original.Get(f)
		pv, _ := parsed.Get(f)
		if ov != pv {
			t.Errorf("field %d: original=%q parsed=%q", f, ov, pv)
		}
	}
}
