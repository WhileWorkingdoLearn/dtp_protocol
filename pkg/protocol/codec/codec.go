package codec

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
)

//The decoder reconstructs a Package from its byte representation using a single pass over the input, for performance reasons avoiding reflection and hash maps.
//It splits on | to obtain fields, splits each field once on the first :, resolves the key via a constant key→index table,
//and detects duplicates with a fixed-size boolean array, which eliminates per-decode allocations and hashing overhead.
//Integers are parsed with strconv.Atoi, the Pyl value is Base64-decoded back to the original []byte,
//and Rma is unescaped in the inverse order (%7C→|, %3A→:, %25→%) before being parsed via net.ResolveUDPAddr("udp", ...).
//The decoder expects the same schema as the encoder; unknown keys are rejected in the fast variant shown, and missing required keys are reported explicitly.

//Edge cases and limitations arise mostly from the flat text framing and the fixed schema. Empty values are legal for Pyl and Rma and decode to nil;
//empty values for integer fields are invalid and cause parsing errors. Invalid Base64 in Pyl or an ill-formed address string in Rma will surface as decoding errors;
//IPv6 addresses are supported because addr.String() yields the bracketed form and ResolveUDPAddr accepts it once unescaped.
//Only three delimiter characters are escaped, so other control characters remain as-is; if your downstream consumers treat newlines or tabs specially,
//consider additional sanitization. Integer range is constrained by the platform int size; extremely large numeric inputs can overflow on 32-bit systems,
//and you should enforce domain limits if negative values are not meaningful in your protocol.

// Feste Feldindizes für bool-Array
const (
	fieldSid = iota
	fieldUid
	fieldMsg
	fieldPId
	fieldBid
	fieldLid
	fieldTol
	fieldPyl
	fieldRma
	numFields
)

const (
	REQ State = iota
	OPN
	ALI
	CLD
	ACK
	RTY
	ERR
)

type State int

type Package struct {
	SessionID     int
	UserID        int
	MSgCode       State
	PackedID      int
	FrameBegin    int
	FrameEnd      int
	PayloadLength int
	Payload       []byte
	Rma           *net.UDPAddr
}

var fieldIndex = map[string]int{
	"Sid": fieldSid,
	"Uid": fieldUid,
	"Msg": fieldMsg,
	"PId": fieldPId,
	"Bid": fieldBid,
	"Lid": fieldLid,
	"Tol": fieldTol,
	"Pyl": fieldPyl,
	"Rma": fieldRma,
}

// Decode decodiert "Field:Value|Field:Value|..." in ein Package.
// Nutzt ein bool-Array statt Map für Duplikaterkennung → zero alloc, O(1).
func Decode(b []byte) (Package, error) {
	var out Package
	seen := [numFields]bool{}

	parts := strings.Split(string(b), "|")
	for i, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			return out, fmt.Errorf("invalid field at pos %d: %q (expected Key:Value)", i, part)
		}
		key, raw := kv[0], kv[1]

		idx, ok := fieldIndex[key]
		if !ok {
			return out, fmt.Errorf("unknown key: %q", key)
		}
		if seen[idx] {
			return out, fmt.Errorf("duplicate key: %q", key)
		}
		seen[idx] = true

		switch idx {
		case fieldSid:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Sid: %w", err)
			}
			out.SessionID = n
		case fieldUid:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Sid: %w", err)
			}
			out.UserID = n
		case fieldMsg:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Msg: %w", err)
			}
			out.MSgCode = State(n)
		case fieldPId:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("PId: %w", err)
			}
			out.PackedID = n
		case fieldBid:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Bid: %w", err)
			}
			out.FrameBegin = n
		case fieldLid:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Lid: %w", err)
			}
			out.FrameEnd = n
		case fieldTol:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Tol: %v", err)
			}
			out.PayloadLength = n
		case fieldPyl:
			if raw == "" {
				out.Payload = nil
				continue
			}
			data, err := base64.StdEncoding.DecodeString(raw)
			if err != nil {
				return out, fmt.Errorf("Pyl (Base64): %w", err)
			}
			out.Payload = data
		case fieldRma:
			if raw == "" {
				out.Rma = nil
				continue
			}
			addrStr := unescapeDelims(raw)
			udp, err := net.ResolveUDPAddr("udp", addrStr)
			if err != nil {
				return out, fmt.Errorf("Rma (UDPAddr): %w (value: %q)", err, addrStr)
			}
			out.Rma = udp
		}
	}

	// Pflichtfelder prüfen
	for name, idx := range fieldIndex {
		if !seen[idx] {
			return out, fmt.Errorf("Decoding: missing required key: %s", name)
		}
	}

	return out, nil
}

// Hilfsfunktion zum Ent-escapen von :, |, %
func unescapeDelims(s string) string {
	// Reihenfolge wichtig: zuerst Sonderzeichen zurück, dann %
	r := strings.NewReplacer(
		"%7C", "|",
		"%3A", ":",
		"%2D", "-",
		"%25", "%",
	)
	return r.Replace(s)
}

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

	sb.Grow(128 + len(p.Payload))

	// ints
	sb.WriteString("Sid:")
	sb.WriteString(strconv.Itoa(p.SessionID))
	sb.WriteByte('|')
	sb.WriteString("Uid:")
	sb.WriteString(strconv.Itoa(p.UserID))
	sb.WriteByte('|')

	sb.WriteString("Msg:")
	sb.WriteString(strconv.Itoa(int(p.MSgCode)))
	sb.WriteByte('|')

	sb.WriteString("PId:")
	sb.WriteString(strconv.Itoa(p.PackedID))
	sb.WriteByte('|')

	sb.WriteString("Bid:")
	sb.WriteString(strconv.Itoa(p.FrameBegin))
	sb.WriteByte('|')

	sb.WriteString("Lid:")
	sb.WriteString(strconv.Itoa(p.FrameEnd))
	sb.WriteByte('|')

	sb.WriteString("Tol:")
	sb.WriteString(strconv.Itoa(p.PayloadLength))
	sb.WriteByte('|')

	// payload → Base64
	sb.WriteString("Pyl:")
	if len(p.Payload) > 0 {
		sb.WriteString(base64.StdEncoding.EncodeToString(p.Payload))
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
	// Reihenfolge wichtig: erst %, dann andere Zeichen
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, "|", "%7C")
	s = strings.ReplaceAll(s, "-", "%2D")
	return s
}
