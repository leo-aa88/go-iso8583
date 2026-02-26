package iso8583

// bitmapBytes is the wire size of one bitmap tier (primary or secondary).
const bitmapBytes = 8 // 64 bits

// parseBitmaps reads one or two 8-byte bitmaps from data starting at offset.
// It returns a 128-element boolean array and the number of bytes consumed.
// Bit 1 of the primary bitmap indicates the presence of a secondary bitmap.
func parseBitmaps(data []byte, offset int) ([128]bool, int, error) {
	var bitmap [128]bool

	if offset+bitmapBytes > len(data) {
		return bitmap, 0, ErrTruncatedBitmap
	}

	// Parse primary bitmap (bits 1-64).
	parseBitmapTier(data[offset:offset+bitmapBytes], bitmap[:64])
	consumed := bitmapBytes

	// Bit 1 set → secondary bitmap follows (bits 65-128).
	if bitmap[0] {
		if offset+consumed+bitmapBytes > len(data) {
			return bitmap, 0, ErrTruncatedBitmap
		}
		parseBitmapTier(data[offset+consumed:offset+consumed+bitmapBytes], bitmap[64:128])
		consumed += bitmapBytes
	}

	return bitmap, consumed, nil
}

// parseBitmapTier maps 8 raw bytes into 64 consecutive bool slots.
func parseBitmapTier(raw []byte, bits []bool) {
	for i, b := range raw {
		for j := 0; j < 8; j++ {
			bits[i*8+j] = b&(1<<(7-uint(j))) != 0
		}
	}
}

// buildBitmaps encodes the field presence map into one or two 8-byte bitmap tiers.
// It automatically sets bit 1 when any field > 64 is present.
func buildBitmaps(fields map[int][]byte) []byte {
	var bits [128]bool

	for fieldNum := range fields {
		if fieldNum < 2 || fieldNum > 128 {
			continue
		}
		bits[fieldNum-1] = true
	}

	// Determine if we need a secondary bitmap.
	needSecondary := false
	for i := 64; i < 128; i++ {
		if bits[i] {
			needSecondary = true
			break
		}
	}

	if needSecondary {
		bits[0] = true // Set bit 1 to signal secondary bitmap presence.
	}

	if needSecondary {
		out := make([]byte, bitmapBytes*2)
		encodeBitmapTier(bits[:64], out[:bitmapBytes])
		encodeBitmapTier(bits[64:128], out[bitmapBytes:])
		return out
	}

	out := make([]byte, bitmapBytes)
	encodeBitmapTier(bits[:64], out)
	return out
}

// encodeBitmapTier packs 64 bool slots into 8 bytes.
func encodeBitmapTier(bits []bool, out []byte) {
	for i := 0; i < 8; i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				b |= 1 << (7 - uint(j))
			}
		}
		out[i] = b
	}
}
