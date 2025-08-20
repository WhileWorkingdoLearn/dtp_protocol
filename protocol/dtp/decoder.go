package dtp

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"

	protocol "github.com/WhilecodingDoLearn/dtp/protocol/types"
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
	fieldMsg
	fieldPId
	fieldBid
	fieldLid
	fieldPyl
	fieldRma
	numFields
)

var fieldIndex = map[string]int{
	"Sid": fieldSid,
	"Msg": fieldMsg,
	"PId": fieldPId,
	"Bid": fieldBid,
	"Lid": fieldLid,
	"Pyl": fieldPyl,
	"Rma": fieldRma,
}

// Decode decodiert "Field:Value|Field:Value|..." in ein Package.
// Nutzt ein bool-Array statt Map für Duplikaterkennung → zero alloc, O(1).
func Decode(b []byte) (protocol.Package, error) {
	var out protocol.Package
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
			out.Sid = n
		case fieldMsg:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Sid: %w", err)
			}
			out.Sid = n
		case fieldPId:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("PId: %w", err)
			}
			out.Pid = n
		case fieldBid:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Bid: %w", err)
			}
			out.Bid = n
		case fieldLid:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return out, fmt.Errorf("Lid: %w", err)
			}
			out.Lid = n
		case fieldPyl:
			if raw == "" {
				out.Pyl = nil
				continue
			}
			data, err := base64.StdEncoding.DecodeString(raw)
			if err != nil {
				return out, fmt.Errorf("Pyl (Base64): %w", err)
			}
			out.Pyl = data
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
			return out, fmt.Errorf("missing required key: %s", name)
		}
	}

	return out, nil
}

// Hilfsfunktion zum Ent-escapen von :, |, %
func unescapeDelims(s string) string {
	r := strings.NewReplacer(
		"%7C", "|",
		"%3A", ":",
		"%25", "%",
	)
	return r.Replace(s)
}
