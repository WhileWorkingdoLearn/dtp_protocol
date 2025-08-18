package protocol

import (
	"encoding/base64"
	"strconv"
	"strings"
)

/*
The encoder produces the canonical byte representation from a Package instance without reflection.Integer fields are written as base-10 strings.
The binary payload Pyl is Base64-encoded so that it remains safe within the textual envelope.
The UDP address Rma is serialized using addr.String() and made delimiter-safe by escaping %, :, and | to %25, %3A, and %7C, respectively; all other characters are left unchanged.
Fields are emitted in a fixed order (Sid|PId|Bid|Lid|Pyl|Rma) to keep the output deterministic, and a strings.Builder is pre-grown to reduce reallocations.
The function returns a []byte that is directly consumable by the decoder and stable across platforms as long as the struct definition remains unchanged.

Edge cases and limitations arise mostly from the flat text framing and the fixed schema. Empty values are legal for Pyl and Rma and decode to nil;
empty values for integer fields are invalid and cause parsing errors. Invalid Base64 in Pyl or an ill-formed address string in Rma will surface as decoding errors;
IPv6 addresses are supported because addr.String() yields the bracketed form and ResolveUDPAddr accepts it once unescaped.
Only three delimiter characters are escaped, so other control characters remain as-is; if your downstream consumers treat newlines or tabs specially,
consider additional sanitization. Integer range is constrained by the platform int size; extremely large numeric inputs can overflow on 32-bit systems,
and you should enforce domain limits if negative values are not meaningful in your protocol.
*/

func Encode(p Package) []byte {
	var sb strings.Builder

	sb.Grow(128 + len(p.Pyl))

	// ints
	sb.WriteString("Sid:")
	sb.WriteString(strconv.Itoa(p.Sid))
	sb.WriteByte('|')

	sb.WriteString("Msg:")
	sb.WriteString(strconv.Itoa(p.Sid))
	sb.WriteByte('|')

	sb.WriteString("PId:")
	sb.WriteString(strconv.Itoa(p.Pid))
	sb.WriteByte('|')

	sb.WriteString("Bid:")
	sb.WriteString(strconv.Itoa(p.Bid))
	sb.WriteByte('|')

	sb.WriteString("Lid:")
	sb.WriteString(strconv.Itoa(p.Lid))
	sb.WriteByte('|')

	// payload → Base64
	sb.WriteString("Pyl:")
	if len(p.Pyl) > 0 {
		sb.WriteString(base64.StdEncoding.EncodeToString(p.Pyl))
	}
	sb.WriteByte('|')

	// UDPAddr → escaped
	sb.WriteString("Rma:")
	if p.Rma != nil {
		sb.WriteString(escapeDelims(p.Rma.String()))
	}

	return []byte(sb.String())
}

// Hilfsfunktion: Escape für Delimiter-Zeichen
func escapeDelims(s string) string {
	// Reihenfolge wichtig: erst %, dann : und |
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, "|", "%7C")
	return s
}
