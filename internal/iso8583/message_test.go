package iso8583_test

import (
	"bytes"
	"testing"

	"github.com/leo-aa88/go-iso8583/internal/iso8583"
)

func TestMessage_NewMessage_StoresMTI(t *testing.T) {
	m := iso8583.NewMessage("0200")
	if m.MTI != "0200" {
		t.Errorf("NewMessage: want MTI %q got %q", "0200", m.MTI)
	}
}

func TestMessage_NewMessage_FieldsInitialised(t *testing.T) {
	m := iso8583.NewMessage("0200")
	if m.Fields == nil {
		t.Error("NewMessage: Fields map should be initialised, got nil")
	}
}

func TestMessage_Set_Get_StringRoundTrip(t *testing.T) {
	m := iso8583.NewMessage("0200")
	m.Set(2, "4111111111111111")

	v, ok := m.Get(2)
	if !ok {
		t.Fatal("Get: field 2 should be present after Set")
	}
	if v != "4111111111111111" {
		t.Errorf("Get: want %q got %q", "4111111111111111", v)
	}
}

func TestMessage_SetBytes_GetBytes_RoundTrip(t *testing.T) {
	m := iso8583.NewMessage("0200")
	original := []byte{0x01, 0x23, 0xFF}
	m.SetBytes(2, original)

	b, ok := m.GetBytes(2)
	if !ok {
		t.Fatal("GetBytes: field 2 should be present after SetBytes")
	}
	if !bytes.Equal(b, original) {
		t.Errorf("GetBytes: want %v got %v", original, b)
	}
}

func TestMessage_Get_MissingField_ReturnsFalse(t *testing.T) {
	m := iso8583.NewMessage("0200")
	_, ok := m.Get(99)
	if ok {
		t.Error("Get: expected ok=false for field that was never set")
	}
}

func TestMessage_GetBytes_MissingField_ReturnsFalse(t *testing.T) {
	m := iso8583.NewMessage("0200")
	_, ok := m.GetBytes(99)
	if ok {
		t.Error("GetBytes: expected ok=false for field that was never set")
	}
}

func TestMessage_Set_OverwritesPreviousValue(t *testing.T) {
	m := iso8583.NewMessage("0200")
	m.Set(2, "1111")
	m.Set(2, "9999")

	v, _ := m.Get(2)
	if v != "9999" {
		t.Errorf("Set overwrite: want %q got %q", "9999", v)
	}
}

func TestMessage_Set_StoresAsBytes(t *testing.T) {
	m := iso8583.NewMessage("0200")
	m.Set(2, "hello")

	// Get as bytes should return the UTF-8 encoding of "hello".
	b, ok := m.GetBytes(2)
	if !ok {
		t.Fatal("GetBytes should find field set via Set")
	}
	if !bytes.Equal(b, []byte("hello")) {
		t.Errorf("expected %v got %v", []byte("hello"), b)
	}
}

func TestMessage_SetBytes_VisibleViaGet(t *testing.T) {
	m := iso8583.NewMessage("0200")
	m.SetBytes(3, []byte("000000"))

	v, ok := m.Get(3)
	if !ok {
		t.Fatal("Get should find field set via SetBytes")
	}
	if v != "000000" {
		t.Errorf("expected %q got %q", "000000", v)
	}
}

func TestMessage_MultipleFields(t *testing.T) {
	m := iso8583.NewMessage("0100")
	m.Set(2, "4111111111111111")
	m.Set(3, "000000")
	m.Set(4, "000000010000")

	fields := map[int]string{2: "4111111111111111", 3: "000000", 4: "000000010000"}
	for field, want := range fields {
		got, ok := m.Get(field)
		if !ok {
			t.Errorf("field %d: not found", field)
			continue
		}
		if got != want {
			t.Errorf("field %d: want %q got %q", field, want, got)
		}
	}
}
