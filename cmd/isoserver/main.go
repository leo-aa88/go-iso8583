package main

import (
	"fmt"
	"log"

	lib "github.com/leo-aa88/go-iso8583/pkg/iso8583lib"
)

func main() {
	// Use the built-in ISO8583-1987 ASCII spec as a starting point.
	spec := lib.NewISO87AsciiSpec()

	// ------------------------------------------------------------------
	// Build a financial transaction request (0200).
	// ------------------------------------------------------------------
	req := lib.NewMessage("0200")
	req.Set(2, "4111111111111111") // PAN (LLVAR, 16 digits)
	req.Set(3, "000000")           // Processing code (Fixed, 6 digits)
	req.Set(4, "000000010000")     // Transaction amount (Fixed, 12 digits)
	req.Set(7, "0101120000")       // Transmission date/time
	req.Set(11, "000001")          // STAN
	req.Set(12, "120000")          // Local time
	req.Set(13, "0101")            // Local date
	req.Set(49, "840")             // Currency code – USD

	wire, err := lib.Build(spec, req)
	if err != nil {
		log.Fatalf("Build failed: %v", err)
	}
	fmt.Printf("Wire (hex): %X\n", wire)
	fmt.Printf("Wire (str): %s\n\n", wire)

	// ------------------------------------------------------------------
	// Parse the message back.
	// ------------------------------------------------------------------
	parsed, err := lib.Parse(spec, wire)
	if err != nil {
		log.Fatalf("Parse failed: %v", err)
	}

	fmt.Printf("MTI: %s\n", parsed.MTI)
	for _, f := range []int{2, 3, 4, 7, 11, 12, 13, 49} {
		if v, ok := parsed.Get(f); ok {
			fmt.Printf("  Field %3d: %s\n", f, v)
		}
	}

	// ------------------------------------------------------------------
	// Build a response (0210) using a secondary-bitmap field (field 70).
	// ------------------------------------------------------------------
	resp := lib.NewMessage("0210")
	resp.Set(39, "00")  // Approval code
	resp.Set(70, "301") // Network management information code (secondary bitmap)

	wire2, err := lib.Build(spec, resp)
	if err != nil {
		log.Fatalf("Build response failed: %v", err)
	}
	fmt.Printf("\nResponse wire (hex): %X\n", wire2)

	parsed2, err := lib.Parse(spec, wire2)
	if err != nil {
		log.Fatalf("Parse response failed: %v", err)
	}
	fmt.Printf("Response MTI: %s\n", parsed2.MTI)
	if v, ok := parsed2.Get(39); ok {
		fmt.Printf("  Field  39: %s\n", v)
	}
	if v, ok := parsed2.Get(70); ok {
		fmt.Printf("  Field  70: %s\n", v)
	}
}
