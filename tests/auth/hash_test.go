package auth_test

import (
	"testing"

	"github.com/Woun1zoN/go-identity-service/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "password"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == "" {
		t.Fatal("hash should not be empty")
	}

	err = auth.CheckPassword(hash, password)
	if err != nil {
		t.Errorf("expected password to match, got error: %v", err)
	}

	err = auth.CheckPassword(hash, "wrongPassword")
	if err == nil {
		t.Errorf("expected error for wrong password, got nil")
	}
}