package token

import "time"

type Maker interface {
	// create a new token for a specific user
	CreateToken(username string, duration time.Duration) (string, *Payload, error)

	// check if token is valid or not
	VerifyToken(token string) (*Payload, error)
}
