package provider

import (
	"testing"
	"time"
)

// TestTokenPersistence_SecretDerivedKey verifies that the on-disk token cache
// key is bound to the client secret: the same secret round-trips, and a
// different secret (same customer/client, so the same cache file) fails closed
// and purges the file instead of decrypting it.
func TestTokenPersistence_SecretDerivedKey(t *testing.T) {
	// os.UserHomeDir reads HOME on Unix and USERPROFILE on Windows; set both so
	// the cache stays under t.TempDir() on every platform.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	const (
		customerID = "cust-123"
		clientID   = "client-abc"
		secret     = "s3cr3t-value"
		token      = "bearer-token-xyz"
	)
	expiresAt := time.Now().Add(time.Hour)

	tp := NewTokenPersistence(customerID, clientID, secret)
	if err := tp.SaveToken(customerID, clientID, token, expiresAt); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	got, err := tp.LoadToken(customerID, clientID)
	if err != nil {
		t.Fatalf("LoadToken with correct secret failed: %v", err)
	}
	if got == nil || got.Token != token {
		t.Fatalf("expected cached token %q, got %+v", token, got)
	}

	// The successful load above leaves the file in place; a load with a different
	// secret must fail to decrypt it and purge it.
	wrong := NewTokenPersistence(customerID, clientID, "different-secret")
	if _, err := wrong.LoadToken(customerID, clientID); err == nil {
		t.Fatalf("expected decryption to fail with a different client secret")
	}

	// The wrong-secret load should have removed the undecryptable file.
	if got, err := tp.LoadToken(customerID, clientID); err != nil || got != nil {
		t.Fatalf("expected empty cache after failed-decrypt purge, got token=%+v err=%v", got, err)
	}
}
