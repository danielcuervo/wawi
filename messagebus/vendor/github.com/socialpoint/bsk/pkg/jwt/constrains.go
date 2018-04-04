package jwt

import "fmt"

// Constraint represents a claim that must be satisfied for a token to be valid
type Constraint func(*Token) error

// ConstraintError raises when a constraint is not satisfied
type ConstraintError string

func (e ConstraintError) Error() string {
	return fmt.Sprintf("claim %s is either missing or invalid", string(e))
}

// HD returns a constraint for the Hosted Domain claim
func HD(hd string) Constraint {
	return func(token *Token) error {
		if token.Claims.Hd != hd {
			return ConstraintError("hd")
		}

		return nil
	}
}

// Iss returns a constraint for the Issuer claim
func Iss(issuers ...string) Constraint {
	return func(token *Token) error {
		for _, issuer := range issuers {
			if token.Claims.Iss == issuer {
				return nil
			}
		}

		return ConstraintError("iss")
	}
}

// Aud returns a constraint for the Audience claim
func Aud(aud string) Constraint {
	return func(token *Token) error {
		if token.Claims.Aud != aud {
			return ConstraintError("aud")
		}

		return nil
	}
}
