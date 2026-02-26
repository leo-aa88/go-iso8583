package iso8583

// Message is the in-memory representation of a parsed or to-be-built ISO8583 message.
// Field values are stored as raw []byte slices exactly as they were provided by the
// caller or decoded from the wire.  No padding or encoding is applied to values
// stored here; that is handled by Build/Parse.
type Message struct {
	// MTI is the 4-character Message Type Indicator (e.g., "0200", "0210").
	MTI string
	// Fields maps ISO8583 field numbers (2-128) to their raw string values.
	// Field 1 (secondary bitmap) is managed automatically and must not be set
	// by callers.
	Fields map[int][]byte
}

// NewMessage returns an initialised Message.
func NewMessage(mti string) *Message {
	return &Message{MTI: mti, Fields: make(map[int][]byte)}
}

// Set stores a string value for the given field number.
func (m *Message) Set(field int, value string) {
	m.Fields[field] = []byte(value)
}

// SetBytes stores a raw byte value for the given field number.
func (m *Message) SetBytes(field int, value []byte) {
	m.Fields[field] = value
}

// Get returns the string value for a field and whether it was present.
func (m *Message) Get(field int) (string, bool) {
	v, ok := m.Fields[field]
	return string(v), ok
}

// GetBytes returns the raw bytes for a field and whether it was present.
func (m *Message) GetBytes(field int) ([]byte, bool) {
	v, ok := m.Fields[field]
	return v, ok
}
