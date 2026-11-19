package users

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUserValidationAndNormalization(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		phone   *string
		wantErr error
	}{
		{name: "empty name", input: "   ", wantErr: ErrNameRequired},
		{name: "name too long", input: strings.Repeat("a", 51), wantErr: ErrNameRequired},
		{name: "phone too long", input: "Alice", phone: stringPtr(strings.Repeat("1", 16)), wantErr: ErrPhoneInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUser(tt.input, tt.phone)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	phone := "  +123456789  "
	got, err := NewUser("  Alice  ", &phone)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Alice" || got.PhoneNumber == nil || *got.PhoneNumber != "+123456789" {
		t.Fatalf("unexpected normalized user: %#v", got)
	}
	if got.ID == [16]byte{} || got.CreatedAt.IsZero() {
		t.Fatalf("user identity/time not initialized: %#v", got)
	}
}

func stringPtr(s string) *string { return &s }
