package services

import "testing"

func TestGenerateApiKey(t *testing.T) {
	plain1, hash1, err := GenerateApiKey()
	if err != nil {
		t.Fatalf("GenerateApiKey returned error: %v", err)
	}
	if plain1 == "" || hash1 == "" {
		t.Fatal("expected non-empty plaintext and hash")
	}
	if plain1 == hash1 {
		t.Fatal("plaintext must not equal its hash")
	}
	// base64url of 32 bytes (no padding) is 43 chars.
	if len(plain1) != 43 {
		t.Fatalf("unexpected plaintext length %d, want 43", len(plain1))
	}
	// sha256 hex is 64 chars.
	if len(hash1) != 64 {
		t.Fatalf("unexpected hash length %d, want 64", len(hash1))
	}

	plain2, _, err := GenerateApiKey()
	if err != nil {
		t.Fatalf("GenerateApiKey returned error: %v", err)
	}
	if plain1 == plain2 {
		t.Fatal("two generated keys must differ")
	}
}

func TestHashApiKeyDeterministic(t *testing.T) {
	const raw = "some-token"
	if HashApiKey(raw) != HashApiKey(raw) {
		t.Fatal("HashApiKey must be deterministic")
	}
	if HashApiKey(raw) == raw {
		t.Fatal("hash must differ from input")
	}
	if HashApiKey("a") == HashApiKey("b") {
		t.Fatal("different inputs must hash differently")
	}
}
