package jwt

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Parse extracts a token from a string
func Parse(tokenString string) (*Token, error) {
	var err error

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrMalformedToken
	}

	token := &Token{
		Content:   strings.Join(parts[0:2], "."),
		Signature: parts[2],
	}

	// parse header
	var header []byte
	if header, err = DecodeSegment(parts[0]); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(header, &token.Header); err != nil {
		return nil, err
	}

	// parse payload
	var payload []byte
	if payload, err = DecodeSegment(parts[1]); err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewBuffer(payload))
	dec.UseNumber()

	if err = dec.Decode(&token.Claims); err != nil {
		return nil, err
	}

	return token, nil
}
