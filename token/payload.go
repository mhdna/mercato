package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrExpiredtoken = errors.New("token has expired")

type Payload struct {
	ID uuid.UUID `json:"id"`
	// UserID identifies the acting user for audit stamping. Tokens issued
	// before this field existed decode with a zero UserID.
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(userID int64, username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	payload := &Payload{
		ID:        tokenID,
		UserID:    userID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

// validate tokens
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredtoken
	}
	return nil
}
