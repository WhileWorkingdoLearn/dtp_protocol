package protocol

import (
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"
)

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
			out.PId = n
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
