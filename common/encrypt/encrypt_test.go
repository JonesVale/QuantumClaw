package encrypt

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := DeriveKey("test-secret-key-1234567890abcd")
	original := []byte("sk-1234567890abcdef1234567890abcdef")

	encrypted, err := Encrypt(original, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if encrypted == "" {
		t.Fatal("Encrypt returned empty string")
	}
	if encrypted == string(original) {
		t.Fatal("Encrypt returned plaintext!")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if !bytes.Equal(decrypted, original) {
		t.Fatalf("Decrypt returned %s, want %s", decrypted, original)
	}
}

func TestEncryptDifferentKeys(t *testing.T) {
	key1 := DeriveKey("secret-1")
	key2 := DeriveKey("secret-2")
	original := []byte("sk-test-key-12345")

	enc1, _ := Encrypt(original, key1)
	dec1, err := Decrypt(enc1, key1)
	if err != nil {
		t.Fatalf("Decrypt with same key failed: %v", err)
	}
	if !bytes.Equal(dec1, original) {
		t.Fatal("Round-trip with same key failed")
	}

	_, err = Decrypt(enc1, key2)
	if err == nil {
		t.Fatal("Decrypt with different key should fail!")
	}
}

func TestDeriveKeyLength(t *testing.T) {
	key := DeriveKey("short")
	if len(key) != 32 {
		t.Errorf("DeriveKey returned key length %d, want 32", len(key))
	}

	key2 := DeriveKey("this-is-a-longer-secret-key-that-should-be-truncated-to-32-bytes")
	if len(key2) != 32 {
		t.Errorf("DeriveKey with long input returned key length %d, want 32", len(key2))
	}
}

func TestEmptyPlaintext(t *testing.T) {
	key := DeriveKey("test-key")
	encrypted, err := Encrypt([]byte(""), key)
	if err != nil {
		t.Fatalf("Encrypt empty failed: %v", err)
	}
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt empty failed: %v", err)
	}
	if len(decrypted) != 0 {
		t.Fatal("Decrypt empty should return empty")
	}
}
