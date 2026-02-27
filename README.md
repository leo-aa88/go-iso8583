# go-iso8583

A **spec-first, deterministic ISO8583 engine** for Go.

`go-iso8583` is built around explicit message specifications, strict validation, and predictable wire output. It avoids reflection and struct-tag magic in favor of full control over every field in the message.

ISO8583 remains a transport-layer concern. Your business structs stay separate.

---

## Philosophy

* **Spec as code** — every ISO8583 variant is defined explicitly in a `Spec` struct. Reviewable, versionable, auditable.
* **Deterministic behavior** — no hidden reflection or runtime tag parsing.
* **Strict validation before wire** — content type, encoding, length bounds, and padding enforced early.
* **Precise field-level errors** — failures are scoped to exact field numbers.
* **Minimal surface area** — zero external dependencies, focused on the core protocol.

This library is suited for:

* Custom ISO8583 variants
* Systems that require explicit message definitions
* Infrastructure components where predictability matters
* Environments that prefer compile-time visibility over tag-driven mapping

---

## Features

* Spec-driven field definitions (`Spec` + `FieldSpec`)
* Primary and secondary bitmap support (automatic bit-1 management)
* Fixed, LLVAR, and LLLVAR field types
* ASCII and BCD encodings
* Strict content-type enforcement (`Numeric`, `Alpha`, `AlphaNumeric`, `Binary`)
* Typed `*FieldError` for field-level failures
* Sentinel errors for structural issues
* Comprehensive table-driven test coverage
* Go 1.20+ only, no dependencies

---

## Installation

```sh
go get github.com/leo-aa88/go-iso8583
```

---

## Defining a Spec

A `Spec` maps ISO8583 field numbers to `FieldSpec` descriptors.
You define one `Spec` per ISO8583 variant.

```go
import lib "github.com/leo-aa88/go-iso8583/pkg/iso8583lib"

var mySpec = &lib.Spec{
    Fields: map[int]lib.FieldSpec{
        3: {
            LengthType:   lib.Fixed,
            MaxLength:    6,
            ContentType:  lib.Numeric,
            Encoding:     lib.ASCII,
            PadDirection: lib.PadLeft,
            PadChar:      '0',
        },
        2: {
            LengthType:  lib.LLVAR,
            MaxLength:   19,
            ContentType: lib.Numeric,
            Encoding:    lib.ASCII,
            PadChar:     '0',
        },
        48: {
            LengthType:  lib.LLLVAR,
            MaxLength:   999,
            ContentType: lib.AlphaNumeric,
            Encoding:    lib.ASCII,
            PadChar:     ' ',
        },
        4: {
            LengthType:   lib.Fixed,
            MaxLength:    12,
            ContentType:  lib.Numeric,
            Encoding:     lib.BCD,
            PadDirection: lib.PadLeft,
            PadChar:      '0',
        },
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

### Enums

| Type           | Values                                       |
| -------------- | -------------------------------------------- |
| `LengthType`   | `Fixed`, `LLVAR`, `LLLVAR`                   |
| `ContentType`  | `Numeric`, `Alpha`, `AlphaNumeric`, `Binary` |
| `EncodingType` | `ASCII`, `BCD`                               |
| `PadDirection` | `PadLeft`, `PadRight`                        |

---

## Building a Message

```go
import lib "github.com/leo-aa88/go-iso8583/pkg/iso8583lib"

spec := lib.NewISO87AsciiSpec()

msg := lib.NewMessage("0200")
msg.Set(2,  "4111111111111111")
msg.Set(3,  "000000")
msg.Set(4,  "000000010000")
msg.Set(11, "000001")
msg.Set(49, "840")

wire, err := lib.Build(spec, msg)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("%X\n", wire)
```

---

## Parsing a Message

```go
import lib "github.com/leo-aa88/go-iso8583/pkg/iso8583lib"

spec := lib.NewISO87AsciiSpec()

parsed, err := lib.Parse(spec, wire)
if err != nil {
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

## Bitmap Handling

* Primary bitmap: fields 1–64
* Secondary bitmap: fields 65–128

If any field > 64 is present, `Build` automatically sets bit 1 and emits the secondary bitmap. Manual bitmap management is not required.

```go
msg := lib.NewMessage("0800")
msg.Set(70, "301")

wire, _ := lib.Build(spec, msg)
```

---

## Variable-Length Fields

| Type     | Prefix         | Max |
| -------- | -------------- | --- |
| `LLVAR`  | 2 ASCII digits | 99  |
| `LLLVAR` | 3 ASCII digits | 999 |

`MaxLength` further constrains accepted values.
If declared wire length exceeds `MaxLength`, `Parse` returns `*FieldError`.

---

## BCD Encoding

BCD packs two decimal digits per byte.

```
"1234"  → [0x12, 0x34]
"12345" → [0x01, 0x23, 0x45]
```

Odd-length values are padded internally and normalized on decode.

Only `Numeric + BCD` combinations are valid.

---

## Error Model

Field-level errors:

```go
type FieldError struct {
    Field int
    Err   error
}
```

Structural sentinel errors:

| Error                | Condition           |
| -------------------- | ------------------- |
| `ErrTruncatedMTI`    | < 4 bytes available |
| `ErrTruncatedBitmap` | bitmap incomplete   |

Errors are explicit and type-checkable via `errors.As`.

---

## Testing

```sh
make test
make test-cover
```

Or:

```sh
go test ./...
go test ./... -coverprofile=coverage.out -covermode=atomic
```

---

## License

MIT — see [LICENSE](LICENSE).
