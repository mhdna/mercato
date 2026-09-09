package token

import "time"

type Maker interface {
	// create a new token for a specific user
	CreateToken(userID int64, username string, duration time.Duration) (string, *Payload, error)

	// check if token is valid or not
	VerifyToken(token string) (*Payload, error)
}
