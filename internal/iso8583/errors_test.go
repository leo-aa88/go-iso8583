package iso8583_test

import (
	"errors"
	"testing"

	"github.com/leo-aa88/go-iso8583/internal/iso8583"
)

func TestFieldError_ErrorString(t *testing.T) {
	fe := &iso8583.FieldError{Field: 42, Err: errors.New("inner error")}
	want := "field 42: inner error"
	if got := fe.Error(); got != want {
		t.Errorf("FieldError.Error: want %q got %q", want, got)
	}
}

func TestFieldError_Unwrap_ErrorsIs(t *testing.T) {
	sentinel := errors.New("sentinel")
	fe := &iso8583.FieldError{Field: 2, Err: sentinel}
	if !errors.Is(fe, sentinel) {
		t.Error("errors.Is should find the wrapped sentinel through FieldError.Unwrap")
	}
}

func TestFieldError_Unwrap_ErrorsAs(t *testing.T) {
	inner := &iso8583.FieldError{Field: 7, Err: errors.New("inner")}
	outer := &iso8583.FieldError{Field: 99, Err: inner}

	var target *iso8583.FieldError
	if !errors.As(outer, &target) {
		t.Fatal("errors.As should find a *FieldError in the chain")
	}
	// errors.As finds the outermost match first.
	if target.Field != 99 {
		t.Errorf("expected outermost field 99, got %d", target.Field)
	}
}

func TestErrTruncatedMTI_Sentinel(t *testing.T) {
	_, err := iso8583.Parse(stdSpec(), []byte("020"))
	if !errors.Is(err, iso8583.ErrTruncatedMTI) {
		t.Errorf("expected ErrTruncatedMTI sentinel, got %v", err)
	}
}

func TestErrTruncatedBitmap_Sentinel(t *testing.T) {
	// 4-byte MTI + 4-byte partial bitmap (need 8).
	_, err := iso8583.Parse(stdSpec(), append([]byte("0200"), make([]byte, 4)...))
	if !errors.Is(err, iso8583.ErrTruncatedBitmap) {
		t.Errorf("expected ErrTruncatedBitmap sentinel, got %v", err)
	}
}

func TestErrUnknownField_IsFieldError(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		2: numericFixed(2),
	})

	msg := buildMsg("0200", map[int]string{2: "42", 9: "01"}) // field 9 not in spec
	_, err := iso8583.Build(spec, msg)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
	var fe *iso8583.FieldError
	if !errors.As(err, &fe) {
		t.Fatalf("expected *FieldError, got %T: %v", err, err)
	}
	if fe.Field != 9 {
		t.Errorf("expected FieldError.Field=9, got %d", fe.Field)
	}
}

func TestFieldError_ParseCarriesFieldNumber(t *testing.T) {
	spec := minSpec(map[int]iso8583.FieldSpec{
		3: numericFixed(6),
	})

	wire := mustBuild(t, spec, buildMsg("0200", map[int]string{3: "123456"}))
	// Truncate so field 3 cannot be fully read.
	_, err := iso8583.Parse(spec, wire[:len(wire)-3])
	if err == nil {
		t.Fatal("expected parse error for truncated field")
	}
	var fe *iso8583.FieldError
	if !errors.As(err, &fe) || fe.Field != 3 {
		t.Errorf("expected FieldError for field 3, got %v", err)
	}
}
