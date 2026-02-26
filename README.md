# go-iso8583

A production-grade, spec-driven ISO8583 message parsing and building library written in Go.

## Features

- **Spec-driven architecture** — every field's length type, content type, encoding, and padding is defined in a `Spec` struct; no hardcoded logic
- **Full bitmap support** — primary bitmap (fields 1–64) and secondary bitmap (fields 65–128) with automatic bit-1 management
- **All common length types** — Fixed, LLVAR (2-digit ASCII prefix), and LLLVAR (3-digit ASCII prefix)
- **Multiple encodings** — ASCII and BCD (packed binary-coded decimal)
- **Strict validation** — content type, length bounds, and encoding errors are all caught before touching the wire
- **Typed errors** — every error identifies the field number and failure reason
- **Zero external dependencies** — pure Go 1.20+
- **Comprehensive test suite** — table-driven unit tests covering parsing, building, encoding, bitmaps, and error paths

---

## Project Structure

```
go-iso8583/
├── cmd/
│   └── isoserver/
│       └── main.go              # Example application (build → parse → print)
├── internal/
│   └── iso8583/
│       ├── bitmap.go            # Primary/secondary bitmap encode & decode
│       ├── bitmap_test.go
│       ├── builder.go           # Build()
│       ├── builder_test.go
│       ├── encoding.go          # BCD, ASCII, padding, content validation
│       ├── encoding_test.go
│       ├── errors.go            # FieldError, sentinel errors
│       ├── errors_test.go
│       ├── message.go           # Message type with Set/Get helpers
│       ├── message_test.go
│       ├── parser.go            # Parse()
│       ├── parser_test.go
│       ├── spec.go              # Spec, FieldSpec, and all enum types
│       ├── spec_iso87.go        # Built-in ISO8583-1987 ASCII spec
│       ├── spec_test.go
│       └── testhelpers_test.go  # Shared test helpers
├── pkg/
│   └── iso8583lib/
│       ├── api.go               # Public re-export of internal types + functions
│       └── api_test.go          # Integration tests via the public API
└── Makefile
```

---

## Installation

```sh
go get github.com/leo-aa88/go-iso8583
```

---

## Spec Definition

A `Spec` maps field numbers to `FieldSpec` descriptors. You define one `Spec` per ISO8583 variant you need to support.

```go
import lib "github.com/leo-aa88/go-iso8583/pkg/iso8583lib"

var mySpec = &lib.Spec{
    Fields: map[int]lib.FieldSpec{
        // Fixed 6-digit numeric field, left-padded with '0'
        3: {
            LengthType:   lib.Fixed,
            MaxLength:    6,
            ContentType:  lib.Numeric,
            Encoding:     lib.ASCII,
            PadDirection: lib.PadLeft,
            PadChar:      '0',
        },
        // LLVAR field (2-digit ASCII length prefix), up to 19 digits
        2: {
            LengthType:  lib.LLVAR,
            MaxLength:   19,
            ContentType: lib.Numeric,
            Encoding:    lib.ASCII,
            PadChar:     '0',
        },
        // LLLVAR alphanumeric field (3-digit ASCII length prefix), up to 999 chars
        48: {
            LengthType:  lib.LLLVAR,
            MaxLength:   999,
            ContentType: lib.AlphaNumeric,
            Encoding:    lib.ASCII,
            PadChar:     ' ',
        },
        // Fixed numeric field with BCD encoding
        4: {
            LengthType:   lib.Fixed,
            MaxLength:    12,
            ContentType:  lib.Numeric,
            Encoding:     lib.BCD,    // "000000050000" → 6 bytes on wire
            PadDirection: lib.PadLeft,
            PadChar:      '0',
        },
        // Secondary bitmap field (field > 64)
        70: {
            LengthType:   lib.Fixed,
            MaxLength:    3,
            ContentType:  lib.Numeric,
            Encoding:     lib.ASCII,
            PadDirection: lib.PadLeft,
            PadChar:      '0',
        },
    },
}
```

### Enum reference

| Type | Constants |
|------|-----------|
| `LengthType` | `Fixed`, `LLVAR`, `LLLVAR` |
| `ContentType` | `Numeric`, `Alpha`, `AlphaNumeric`, `Binary` |
| `EncodingType` | `ASCII`, `BCD` |
| `PadDirection` | `PadLeft`, `PadRight` |

---

## Build Example

```go
import lib "github.com/leo-aa88/go-iso8583/pkg/iso8583lib"

spec := lib.NewISO87AsciiSpec() // or your own Spec

msg := lib.NewMessage("0200")
msg.Set(2,  "4111111111111111") // LLVAR PAN
msg.Set(3,  "000000")           // Fixed processing code
msg.Set(4,  "000000010000")     // Fixed amount
msg.Set(11, "000001")           // Fixed STAN
msg.Set(49, "840")              // Fixed currency code

wire, err := lib.Build(spec, msg)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%X\n", wire)
```

---

## Parse Example

```go
import lib "github.com/leo-aa88/go-iso8583/pkg/iso8583lib"

spec := lib.NewISO87AsciiSpec()

parsed, err := lib.Parse(spec, wire)
if err != nil {
    // Check if it is a field-level error
    var fe *lib.FieldError
    if errors.As(err, &fe) {
        log.Fatalf("field %d failed: %v", fe.Field, fe.Err)
    }
    log.Fatal(err)
}

fmt.Println("MTI:", parsed.MTI)

if pan, ok := parsed.Get(2); ok {
    fmt.Println("PAN:", pan)
}
```

---

## Secondary Bitmap

ISO8583 supports up to 128 fields. The primary bitmap covers fields 1–64 and the secondary bitmap covers fields 65–128.

**Bit 1** of the primary bitmap is the secondary-bitmap indicator. When set, the parser reads an additional 8 bytes immediately after the primary bitmap as the secondary bitmap.

`Build` handles this automatically: if any field number > 64 is present in your `Message`, bit 1 is set and the secondary bitmap is emitted. You never need to manage bit 1 yourself.

```go
msg := lib.NewMessage("0800")
msg.Set(70, "301") // Field 70 is in the secondary bitmap

wire, _ := lib.Build(spec, msg)
// wire[4] & 0x80 != 0  ← bit 1 auto-set
// wire[4:20] contains both bitmaps (16 bytes total)
```

---

## LLVAR and LLLVAR

Variable-length fields are prefixed on the wire with their content length as ASCII decimal digits.

| Type | Prefix | Max representable length |
|------|--------|--------------------------|
| `LLVAR` | 2 digits | 99 |
| `LLLVAR` | 3 digits | 999 |

The `MaxLength` in `FieldSpec` further caps what values are accepted. If a parsed message declares a length exceeding `MaxLength`, `Parse` returns a `*FieldError`.

```
Wire for LLVAR field 2 value "4111111111111111" (16 digits):
  "16" + "4111111111111111"

Wire for LLLVAR field 48 value "HELLO" (5 chars):
  "005" + "HELLO"
```

---

## BCD Encoding

BCD (Binary-Coded Decimal) packs two decimal digits per byte. Only `Numeric` + `BCD` combinations are valid.

```
"1234"  → [0x12, 0x34]   (2 bytes)
"12345" → [0x01, 0x23, 0x45]  (odd → leading zero nibble added)
```

On decode, any leading zero nibble inserted for odd-length values is removed automatically.

---

## Error Handling

All field-level errors are returned as `*FieldError`:

```go
type FieldError struct {
    Field int    // ISO8583 field number
    Err   error  // specific reason
}
```

Sentinel errors for non-field failures:

| Error | When |
|-------|------|
| `ErrTruncatedMTI` | fewer than 4 bytes available |
| `ErrTruncatedBitmap` | bitmap bytes are cut off |

```go
var fe *lib.FieldError
if errors.As(err, &fe) {
    switch {
    case fe.Field == 2:
        // PAN-specific handling
    default:
        log.Printf("field %d: %v", fe.Field, fe.Err)
    }
}
```

---

## Running Tests
```sh
make test                  # run all tests
make test-verbose          # run with -v
make test-cover            # print coverage summary
make cover-html            # open HTML coverage report in browser
```

Or directly with go:
```sh
go test ./...
go test ./internal/iso8583/ -v
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out
```

---

## Related Tools

The following are external tools and projects within the ecosystem that you might find useful when working with ISO 8583 messages. Please note that these are maintained independently.

* **[IsoFluent](https://isofluent.com)** — A free, client-side visual debugging suite for payment systems engineers. It provides a web-based interface for parsing and inspecting ISO 8583 messages, alongside utilities for handling EMV and TLV data.

---

## License

MIT — see [LICENSE](LICENSE).
