package logic_account

import (
	"context"
	"errors"
	"fmt"

	"github.com/Fiagram/standalone/internal/configs"
	"golang.org/x/crypto/bcrypt"
)

type Hash interface {
	Hash(ctx context.Context, input string) (hashed string, err error)
	IsHashEqual(ctx context.Context, input string, hashed string) (bool, error)
}

type hash struct {
	hashConfig configs.Hash
}

func NewHash(hashConfig configs.Hash) Hash {
	return &hash{
		hashConfig,
	}
}

func (h hash) Hash(_ context.Context, input string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(input), h.hashConfig.Cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash data: %w", err)
	}
	return string(hashed), nil
}

func (h hash) IsHashEqual(_ context.Context, input string, hashed string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("failed to compare input: %w", err)
	}
	return true, nil
}
