# ISO8583 Message Processing Library in Go

This Go library provides tools for parsing and generating ISO8583 financial transaction messages, a standard format used by the banking industry for the transmission of transaction data.

## Features

- Parse ISO8583 messages from a string representation into a structured format.
- Generate ISO8583 message strings from structured data.
- Support for primary bitmap parsing and generation to identify present data fields.
- Utility functions for common operations such as field padding, trimming, and validation.
- Customizable field descriptors to adapt to different versions of the ISO8583 standard.

## Prerequisites

- Go version 1.15 or higher.

## Installation

To install the ISO8583 library, use the following `go get` command:

```sh
go get github.com/leo-aa88/go-iso8583
```

Replace `github.com/leo-aa88/go-iso8583` with the actual path to your library on GitHub.

## Usage

### Parsing an ISO8583 Message

To parse an ISO8583 message string into a structured format:

```go
package main

import (
    "fmt"
    "log"

    "github.com/leo-aa88/go-iso8583"
)

func main() {
    messageStr := "0200C000000000000000413334455667788990600000100005000"
    message, err := iso8583.ParseMessage(messageStr)
    if err != nil {
        log.Fatalf("Failed to parse message: %v", err)
    }

    fmt.Printf("Parsed Message:\nMTI: %s\n", message.MTI)
    for i, present := range message.Bitmap {
        if present {
            fmt.Printf("Field %d: %s\n", i+1, message.Fields[i+1])
        }
    }
}
```

### Generating an ISO8583 Message

To generate an ISO8583 message string from structured data:

```go
package main

import (
    "fmt"

    "github.com/leo-aa88/go-iso8583"
)

func main() {
    responseMsg := iso8583.Message{
        MTI: "0210",
        Bitmap: [128]bool{true}, // Simplify: Only including necessary fields for example
        Fields: map[int]string{
            39: "00", // Approval response code
        },
    }

    messageStr := iso8583.GenerateMessage(&responseMsg)
    fmt.Println("Generated ISO8583 Message:", messageStr)
}
```

## Contributing

We welcome contributions! Please feel free to submit pull requests, report bugs, and suggest features.

## License

This project is licensed under the [GPL License](LICENSE).

## Acknowledgments

- This library is intended for educational and informational purposes. It may require modifications to be used in production systems.

## Project Structure

```
go-iso8583/
├── cmd/                        # Main applications for this project
│   └── isoserver/              # A sample application server handling ISO8583 messages
│       └── main.go             # The entry point for the server
├── internal/                   # Private application and library code
│   ├── parser/                 # Logic for parsing ISO8583 messages
│   │   └── parser.go           # Implementation of the parsing logic
│   ├── generator/              # Logic for generating ISO8583 messages
│   │   └── generator.go        # Implementation of the generation logic
│   └── iso8583/                # Core ISO8583 message structures and utilities
│       ├── message.go          # Defines the Message struct and basic operations
│       └── types.go            # Additional types and constants for ISO8583 processing
├── pkg/                        # Library code that's ok to use by external applications
│   └── iso8583lib/             # Publicly accessible functions and types for ISO8583 processing
│       └── api.go              # Public API interfaces for parsing and generating messages
├── test/                       # Additional external test apps and test data
│   └── iso8583_test.go         # Test cases for ISO8583 message processing
└── go.mod                      # Go module definition
└── README.md                   # Project overview and documentation
```

### Description of Components

- **`cmd/`**: This directory contains the main applications for your project. For an ISO8583 handling project, you might have a server that listens for ISO8583 messages, processes them, and sends responses.

- **`internal/`**: The `internal` directory is for private application code. This is where the bulk of your ISO8583 processing logic lives, split into logical subpackages like `parser` and `generator` for better organization.

    - **`parser/`**: Contains code for parsing ISO8583 messages from various formats into your internal message structures.
    - **`generator/`**: Contains code for generating ISO8583 messages based on internal structures or business logic.
    - **`iso8583/`**: Defines the core data structures and utility functions specific to ISO8583 message handling.

- **`pkg/`**: Contains library code that can be used by other applications. This is where you define the public API for your ISO8583 library, making it easy for external developers to integrate with your package.

- **`test/`**: Contains additional tests, including integration and end-to-end tests. This directory can also hold test data or scripts for testing your ISO8583 message handlers.

- **`go.mod`**: Defines your project's module and its dependencies. This file is crucial for managing dependencies in a straightforward manner.

- **`README.md`**: Provides an overview of your project, how to set it up, and how to use it. This is the first document users and contributors will look at, so it's important to keep it up-to-date.

### Key Points

- **Separation of Concerns**: By dividing your project into logical units (parsing, generating, core utilities), you make the codebase easier to navigate and maintain.

- **Scalability**: This structure allows you to easily add new functionalities, such as support for additional message versions or extensions, by adding new packages or files without disrupting existing code.

- **Encapsulation**: Keeping internal implementation details in `internal/` ensures that the library's public API in `pkg/` remains clean and focused, reducing the risk of unintended usage or dependencies on internal structures.

- **Testing**: With a clear structure, it's easier to write and maintain unit tests for each package, as well as integration tests that cover the interactions between packages.

This structure serves as a starting point. Depending on the scope and scale of your ISO8583 processing needs, you might need to adjust it, but adhering to Go's conventions and structuring your project thoughtfully from the start will pay off in the long run.

## Example

### Use Case Scenario

Imagine you are developing a payment gateway that receives transaction requests from Point of Sale (POS) terminals. These requests are formatted according to the ISO 8583 standard. Your task is to parse these requests to extract information such as the card number, transaction amount, and merchant identifier, then respond with an approval or denial message.

We'll focus on a simplified version of an ISO 8583 message that includes the following fields:

- **MTI (Message Type Indicator)**: Specifies the version of the ISO 8583 standard and the message class, function, and origin. For this example, we'll use `0200` for a financial transaction request.
- **Primary Account Number (PAN)**: Field 2, a variable-length field containing the card number.
- **Processing Code**: Field 3, a fixed-length field that specifies the transaction type, account affected, and special conditions.
- **Transaction Amount**: Field 4, a fixed-length field representing the amount for the transaction.

### Step 1: Define Field Descriptors

First, we update our `fieldDescriptors` map to include our relevant fields. This would be part of our setup in the ISO8583 library (assuming these descriptors are defined within the `internal/iso8583/types.go` or a similar configuration file):

```go
var fieldDescriptors = map[int]FieldDescriptor{
    2: {Length: 19, Variable: true, Type: "n"},  // Primary account number, variable length up to 19 digits
    3: {Length: 6, Variable: false, Type: "n"},  // Processing code, 6 digits
    4: {Length: 12, Variable: false, Type: "n"}, // Transaction amount, 12 digits
}
```

### Step 2: Parse an Incoming ISO8583 Message

Let's assume we receive a message string from a POS terminal. For simplicity, the message string will be simplified and not include a bitmap indicating the presence of fields. In a real scenario, the message would contain a bitmap, and parsing logic would use it to determine which fields are present.

```go
messageStr := "0200B23..."; // Simplified; B23... represents the rest of the message including PAN, processing code, and amount
```

We would use the `ParseMessage` function from our library to parse this message. This function would internally use the `parseBitmap` and other utilities we've defined:

```go
message, err := iso8583.ParseMessage(messageStr)
if err != nil {
    log.Fatalf("Failed to parse message: %v", err)
}
```

### Step 3: Generate an Acknowledgment Message

After parsing the message and processing the transaction (e.g., validating the card number, checking the transaction against business rules), we need to generate a response message. Let's assume the transaction is approved, and we generate an ISO 8583 message with an MTI of `0210` to indicate a financial transaction response.

```go
responseMsg := iso8583.Message{
    MTI: "0210",
    Fields: map[int]string{
        39: "00", // Response code field (39) set to "00" indicating approval
    },
}
bitmap := [128]bool{true} // Simplify: Only field 39 is present, actual implementation would calculate the bitmap based on present fields
responseMsgStr := iso8583.GenerateMessage(&responseMsg, bitmap)

// Send responseMsgStr back to the POS terminal
```

### Summary

This example demonstrates a high-level use of the ISO8583 library to parse an incoming message and generate a response. The actual implementation details would involve more complexity, including handling the bitmap correctly, dealing with variable-length fields, and ensuring compliance with the specific version of the ISO 8583 standard being used.

Note that the example simplifies many aspects of ISO 8583 message handling, such as the extraction and setting of fields based on a real bitmap, error handling, and network communication logistics, to focus on illustrating how the core functions of the library might be used.

To include a version with the bitmap in the processing of an ISO8583 message, we need to consider the structure of an ISO8583 message which includes the Message Type Indicator (MTI), the bitmap, and the data elements (fields). The bitmap is a binary or hexadecimal representation indicating which fields are present in the message.

Let's assume a simplified version of an ISO8583 message including a primary bitmap (not using a secondary bitmap for simplicity) and fields for a financial transaction request. Our example message will include the following fields:

- MTI: `0200` for a financial transaction request
- Primary Account Number (Field 2)
- Processing Code (Field 3)
- Transaction Amount (Field 4)

### Example Message Structure with Bitmap

The bitmap will indicate the presence of fields 2, 3, and 4. Assuming no other fields are present, and using a 64-bit (16 hexadecimal digits) primary bitmap, our bitmap for this example would look something like this in binary: `11000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000`, where the first bit (not used here) is for the secondary bitmap indicator and bits 2, 3, and 4 are set to indicate the presence of fields 2, 3, and 4, respectively. In hexadecimal, this bitmap translates to `C000000000000000`.

Given this setup, a message string including the MTI, bitmap, and fields might look like this (with field values simplified):

```plaintext
0200C000000000000000413334455667788990600000100005000
```

- `0200` is the MTI.
- `C000000000000000` is the bitmap indicating that fields 2, 3, and 4 are present.
- The rest of the message string includes the values for fields 2, 3, and 4. For simplicity, let's assume:
  - Field 2 (Primary Account Number) is "413334455667788", a 15-digit number.
  - Field 3 (Processing Code) is "006000", a 6-digit number.
  - Field 4 (Transaction Amount) is "000010000500", representing an amount with leading zeros for padding.

### Parsing the Message

When parsing this message, the `ParseMessage` function would use the bitmap to determine which fields to parse out of the message string. Here's a high-level overview of how that might be implemented, using the functions and structures we've defined:

```go
messageStr := "0200C000000000000000413334455667788990600000100005000"

// Parse the message
message, err := iso8583.ParseMessage(messageStr)
if err != nil {
    log.Fatalf("Failed to parse message: %v", err)
}

fmt.Printf("Parsed Message:\nMTI: %s\n", message.MTI)
for i, present := range message.Bitmap {
    if present {
        fmt.Printf("Field %d: %s\n", i+1, message.Fields[i+1])
    }
}
```

This example assumes the `ParseMessage` function is sophisticated enough to interpret the bitmap and extract the fields accordingly. The actual parsing would involve converting the hexadecimal bitmap to a binary form (or a boolean array), checking which fields are indicated as present, and then parsing those fields based on their definitions in the `fieldDescriptors` map or equivalent structure.

### Generating a Response Message

For generating a response message, you would construct a `Message` struct with the appropriate MTI, set the bitmap to indicate which fields are included in the response, and populate those fields with data. The `GenerateMessage` function would then create a message string from this struct, including the appropriate bitmap.

This simplified example demonstrates parsing and generating ISO8583 messages with a bitmap. The actual implementation details would be more complex and need to account for variable-length fields, the exact format of the fields, and the full 128-field capacity of ISO8583 messages (including the use of a secondary bitmap if necessary).
