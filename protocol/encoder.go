package protocol

import (
	"encoding/base64"
	"strconv"
	"strings"
)

// Encode serialisiert ein Package ins Format
// "Sid:<int>|PId:<int>|Bid:<int>|Lid:<int>|Pyl:<base64>|Rma:<escaped ip:port>"
func Encode(p Package) []byte {
	var sb strings.Builder
	// grob vorkonfigurieren, damit keine Re-Allokationen passieren
	sb.Grow(128 + len(p.Pyl))

	// ints
	sb.WriteString("Sid:")
	sb.WriteString(strconv.Itoa(p.Sid))
	sb.WriteByte('|')

	sb.WriteString("Msg:")
	sb.WriteString(strconv.Itoa(p.Sid))
	sb.WriteByte('|')

	sb.WriteString("PId:")
	sb.WriteString(strconv.Itoa(p.PId))
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
