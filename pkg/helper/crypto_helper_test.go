package helper

import (
	"crypto/sha256"
	"strings"
	"testing"
)

func testKey() []byte {
	sum := sha256.Sum256([]byte("test-secret-for-crypto"))
	return sum[:]
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := testKey()
	plaintext := "super-secret-client-secret"

	ct, err := encryptWithKey(key, plaintext)
	if err != nil {
		t.Fatalf("encryptWithKey failed: %v", err)
	}
	if !strings.HasPrefix(ct, cryptoPrefix) {
		t.Fatalf("ciphertext missing prefix, got %q", ct)
	}
	if strings.Contains(ct, plaintext) {
		t.Fatalf("ciphertext should not contain plaintext")
	}

	got, err := decryptWithKey(key, ct)
	if err != nil {
		t.Fatalf("decryptWithKey failed: %v", err)
	}
	if got != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, got)
	}
}

func TestEncryptYieldsDifferentCiphertextEachCall(t *testing.T) {
	key := testKey()
	a, err := encryptWithKey(key, "abc")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	b, err := encryptWithKey(key, "abc")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if a == b {
		t.Fatal("expected fresh nonce to produce different ciphertexts for the same plaintext")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	key := testKey()
	ct, err := encryptWithKey(key, "hello")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	otherSum := sha256.Sum256([]byte("another-secret"))
	if _, err := decryptWithKey(otherSum[:], ct); err == nil {
		t.Fatal("decrypt with wrong key should fail")
	}
}

func TestDecryptSecretPlaintextPassthrough(t *testing.T) {
	// 历史明文(无前缀)应原样返回,便于无缝迁移。
	out, err := DecryptSecret("legacy-plain-secret")
	if err != nil {
		t.Fatalf("DecryptSecret returned err: %v", err)
	}
	if out != "legacy-plain-secret" {
		t.Fatalf("expected passthrough, got %q", out)
	}
}

func TestDecryptSecretEmpty(t *testing.T) {
	out, err := DecryptSecret("")
	if err != nil {
		t.Fatalf("DecryptSecret returned err: %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty, got %q", out)
	}
}
