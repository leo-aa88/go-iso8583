package iso8583

func NewISO87AsciiSpec() *Spec {
	return &Spec{
		Fields: map[int]FieldSpec{
			2:  {LengthType: LLVAR, MaxLength: 19, ContentType: Numeric, Encoding: ASCII, PadChar: '0'},
			3:  {LengthType: Fixed, MaxLength: 6, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			4:  {LengthType: Fixed, MaxLength: 12, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			7:  {LengthType: Fixed, MaxLength: 10, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			11: {LengthType: Fixed, MaxLength: 6, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			12: {LengthType: Fixed, MaxLength: 6, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			13: {LengthType: Fixed, MaxLength: 4, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			22: {LengthType: Fixed, MaxLength: 3, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			25: {LengthType: Fixed, MaxLength: 2, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			35: {LengthType: LLVAR, MaxLength: 37, ContentType: AlphaNumeric, Encoding: ASCII, PadChar: ' '},
			37: {LengthType: Fixed, MaxLength: 12, ContentType: AlphaNumeric, Encoding: ASCII, PadDirection: PadRight, PadChar: ' '},
			38: {LengthType: Fixed, MaxLength: 6, ContentType: AlphaNumeric, Encoding: ASCII, PadDirection: PadRight, PadChar: ' '},
			39: {LengthType: Fixed, MaxLength: 2, ContentType: AlphaNumeric, Encoding: ASCII, PadDirection: PadRight, PadChar: ' '},
			41: {LengthType: Fixed, MaxLength: 8, ContentType: AlphaNumeric, Encoding: ASCII, PadDirection: PadRight, PadChar: ' '},
			42: {LengthType: Fixed, MaxLength: 15, ContentType: AlphaNumeric, Encoding: ASCII, PadDirection: PadRight, PadChar: ' '},
			43: {LengthType: Fixed, MaxLength: 40, ContentType: AlphaNumeric, Encoding: ASCII, PadDirection: PadRight, PadChar: ' '},
			48: {LengthType: LLLVAR, MaxLength: 999, ContentType: AlphaNumeric, Encoding: ASCII, PadChar: ' '},
			49: {LengthType: Fixed, MaxLength: 3, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			55: {LengthType: LLLVAR, MaxLength: 999, ContentType: Binary, Encoding: ASCII, PadChar: 0x00},
			70: {LengthType: Fixed, MaxLength: 3, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
			90: {LengthType: Fixed, MaxLength: 42, ContentType: Numeric, Encoding: ASCII, PadDirection: PadLeft, PadChar: '0'},
		},
	}
}
