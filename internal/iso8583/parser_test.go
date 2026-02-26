package iso8583_test

import (
	"errors"
	"testing"

	"github.com/leo-aa88/go-iso8583/internal/iso8583"
)

func TestParser_NilSpec(t *testing.T) {
	_, err := iso8583.Parse(nil, []byte("02000000000000000000"))
	if err == nil {
		t.Error("expected error for nil spec")
	}
}

func TestParser_TruncatedMTI(t *testing.T) {
	_, err := iso8583.Parse(stdSpec(), []byte("020"))
	if !errors.Is(err, iso8583.ErrTruncatedMTI) {
		t.Errorf("expected ErrTruncatedMTI, got %v", err)
	}
}

func TestParser_TruncatedBitmap(t *testing.T) {
	_, err := iso8583.Parse(stdSpec(), append([]byte("0200"), make([]byte, 4)...))
	if !errors.Is(err, iso8583.ErrTruncatedBitmap) {
		t.Errorf("expected ErrTruncatedBitmap, got %v", err)
	}
}

func TestParser_UnknownFieldInBitmap(t *testing.T) {
	// Build with a full spec that knows fields 2 and 3.
	fullSpec := minSpec(map[int]iso8583.FieldSpec{
		2: numericFixed(2),
		3: numericFixed(2),
	})
	wire := mustBuild(t, fullSpec, buildMsg("0200", map[int]string{2: "42", 3: "07"}))

	// Parse with a limited spec that only knows field 2.
	limitedSpec := minSpec(map[int]iso8583.FieldSpec{
		2: numericFixed(2),
	})
	_, err := iso8583.Parse(limitedSpec, wire)
	if err == nil {
		t.Error("expected error when bitmap references field not in spec")
	}
	var fe *iso8583.FieldError
	if !errors.As(err, &fe) || fe.Field != 3 {
		t.Errorf("expected FieldError for field 3, got %v", err)
	}
}

func TestParser_FixedField_InsufficientData(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		3: numericFixed(6),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{3: "123456"}))
	// Chop the last 2 bytes of the field value.
	_, err := iso8583.Parse(spec, wire[:len(wire)-2])
	if err == nil {
		t.Error("expected error for truncated fixed field")
	}
	var fe *iso8583.FieldError
	if !errors.As(err, &fe) || fe.Field != 3 {
		t.Errorf("expected FieldError for field 3, got %v", err)
	}
}

func TestParser_LLVARField_InvalidPrefix(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		2: numericLLVAR(19),
	})

	data := []byte("0200")
	data = append(data, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00) // bit 2 set
	data = append(data, []byte("XX4111111111111111")...)                // non-numeric prefix
	_, err := iso8583.Parse(spec, data)
	if err == nil {
		t.Error("expected error for non-numeric LLVAR prefix")
	}
}

func TestParser_LLVARField_InsufficientData(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		2: numericLLVAR(19),
	})

	data := []byte("0200")
	data = append(data, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
	data = append(data, []byte("161234")...) // prefix says 16, only 4 digits follow
	_, err := iso8583.Parse(spec, data)
	if err == nil {
		t.Error("expected error for insufficient LLVAR data")
	}
	var fe *iso8583.FieldError
	if !errors.As(err, &fe) || fe.Field != 2 {
		t.Errorf("expected FieldError for field 2, got %v", err)
	}
}

func TestParser_LLVARField_LengthExceedsMax(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		2: numericLLVAR(5),
	})

	data := []byte("0200")
	data = append(data, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
	data = append(data, []byte("101234567890")...) // prefix 10 > MaxLength 5
	_, err := iso8583.Parse(spec, data)
	if err == nil {
		t.Error("expected error when LLVAR declared length exceeds MaxLength")
	}
}

func TestParser_LLLVARField_InvalidPrefix(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		48: alphaNumericLLLVAR(999),
	})

	data := []byte("0200")
	// Bit 48: byte (48-1)/8=5, bit (48-1)%8=7 → 0x01 in byte 5.
	data = append(data, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00)
	data = append(data, []byte("ABCHello")...) // non-numeric 3-char prefix
	_, err := iso8583.Parse(spec, data)
	if err == nil {
		t.Error("expected error for non-numeric LLLVAR prefix")
	}
}

func TestParser_LLLVARField_InsufficientData(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		48: alphaNumericLLLVAR(999),
	})

	data := []byte("0200")
	data = append(data, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00)
	data = append(data, []byte("100ABC")...) // prefix says 100, only 3 bytes follow
	_, err := iso8583.Parse(spec, data)
	if err == nil {
		t.Error("expected error for insufficient LLLVAR data")
	}
}

func TestParser_LLLVARField_LengthExceedsMax(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		48: alphaNumericLLLVAR(10),
	})

	data := []byte("0200")
	data = append(data, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00)
	data = append(data, []byte("050HELLOWORLD")...) // prefix 50 > MaxLength 10
	_, err := iso8583.Parse(spec, data)
	if err == nil {
		t.Error("expected error when LLLVAR declared length exceeds MaxLength")
	}
}

func TestParser_ValidFixedField(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		3: numericFixed(6),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{3: "123456"}))
	parsed := mustParse(t, spec, wire)

	v, ok := parsed.Get(3)
	if !ok || v != "123456" {
		t.Errorf("field 3: want %q ok=true, got %q ok=%v", "123456", v, ok)
	}
}

func TestParser_MTIParsedCorrectly(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		39: alphaNumericFixed(2),
	})

	for _, mti := range []string{"0200", "0210", "0800", "0110"} {
		t.Run(mti, func(t *testing.T) {
			wire := mustBuild(t, spec, buildMsg(mti, map[int]string{39: "00"}))
			parsed := mustParse(t, spec, wire)
			if parsed.MTI != mti {
				t.Errorf("MTI: want %q got %q", mti, parsed.MTI)
			}
		})
	}
}
