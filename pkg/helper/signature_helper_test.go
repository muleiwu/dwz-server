package helper

import (
	"testing"
	"time"
)

// TestGenerateAndVerifySignature tests the basic signature generation and verification
func TestGenerateAndVerifySignature(t *testing.T) {
	helper := GetSignatureHelper()

	secret := "test-secret-key"
	method := "POST"
	path := "/api/v1/shortlinks"
	params := map[string]interface{}{
		"url":  "https://example.com",
		"name": "test",
	}
	timestamp := time.Now().Unix()
	nonce := "random-nonce-123"

	// Generate signature
	signature := helper.GenerateSignature(secret, method, path, params, timestamp, nonce)

	if signature == "" {
		t.Error("Generated signature should not be empty")
	}

	// Verify signature
	if !helper.VerifySignature(secret, method, path, params, timestamp, nonce, signature) {
		t.Error("Signature verification should succeed with same parameters")
	}

	// Verify with wrong secret should fail
	if helper.VerifySignature("wrong-secret", method, path, params, timestamp, nonce, signature) {
		t.Error("Signature verification should fail with wrong secret")
	}
}

// TestGenerateAppID tests App ID generation
func TestGenerateAppID(t *testing.T) {
	helper := GetSignatureHelper()

	appID, err := helper.GenerateAppID()
	if err != nil {
		t.Errorf("GenerateAppID should not return error: %v", err)
	}

	if appID == "" {
		t.Error("Generated AppID should not be empty")
	}

	// Check format: should start with "app_"
	if len(appID) < 4 || appID[:4] != "app_" {
		t.Errorf("AppID should start with 'app_', got: %s", appID)
	}

	// Generate another and ensure uniqueness
	appID2, _ := helper.GenerateAppID()
	if appID == appID2 {
		t.Error("Generated AppIDs should be unique")
	}
}

// TestGenerateAppSecret tests App Secret generation
func TestGenerateAppSecret(t *testing.T) {
	helper := GetSignatureHelper()

	secret, err := helper.GenerateAppSecret()
	if err != nil {
		t.Errorf("GenerateAppSecret should not return error: %v", err)
	}

	if secret == "" {
		t.Error("Generated AppSecret should not be empty")
	}

	// Should be 64 characters (32 bytes hex encoded)
	if len(secret) != 64 {
		t.Errorf("AppSecret should be 64 characters, got: %d", len(secret))
	}

	// Generate another and ensure uniqueness
	secret2, _ := helper.GenerateAppSecret()
	if secret == secret2 {
		t.Error("Generated AppSecrets should be unique")
	}
}

// TestEncryptDecryptAppSecret tests encryption and decryption round-trip
func TestEncryptDecryptAppSecret(t *testing.T) {
	helper := GetSignatureHelper()

	originalSecret := "my-super-secret-app-secret-12345"

	// Encrypt
	encrypted, err := helper.EncryptAppSecret(originalSecret)
	if err != nil {
		t.Errorf("EncryptAppSecret should not return error: %v", err)
	}

	if encrypted == "" {
		t.Error("Encrypted secret should not be empty")
	}

	if encrypted == originalSecret {
		t.Error("Encrypted secret should be different from original")
	}

	// Decrypt
	decrypted, err := helper.DecryptAppSecret(encrypted)
	if err != nil {
		t.Errorf("DecryptAppSecret should not return error: %v", err)
	}

	if decrypted != originalSecret {
		t.Errorf("Decrypted secret should match original. Got: %s, Expected: %s", decrypted, originalSecret)
	}
}

// TestValidateTimestamp tests timestamp validation
func TestValidateTimestamp(t *testing.T) {
	helper := GetSignatureHelper()
	currentTime := time.Now().Unix()

	// Valid timestamp (within 5 minutes)
	if !helper.ValidateTimestamp(currentTime, currentTime) {
		t.Error("Current timestamp should be valid")
	}

	// Valid timestamp (4 minutes ago)
	if !helper.ValidateTimestamp(currentTime-240, currentTime) {
		t.Error("Timestamp 4 minutes ago should be valid")
	}

	// Valid timestamp (4 minutes in future)
	if !helper.ValidateTimestamp(currentTime+240, currentTime) {
		t.Error("Timestamp 4 minutes in future should be valid")
	}

	// Invalid timestamp (6 minutes ago)
	if helper.ValidateTimestamp(currentTime-360, currentTime) {
		t.Error("Timestamp 6 minutes ago should be invalid")
	}

	// Invalid timestamp (6 minutes in future)
	if helper.ValidateTimestamp(currentTime+360, currentTime) {
		t.Error("Timestamp 6 minutes in future should be invalid")
	}
}

// TestValidateNonce tests nonce validation
func TestValidateNonce(t *testing.T) {
	helper := GetSignatureHelper()

	// Valid nonce
	if !helper.ValidateNonce("valid-nonce") {
		t.Error("Non-empty nonce should be valid")
	}

	// Invalid nonce (empty)
	if helper.ValidateNonce("") {
		t.Error("Empty nonce should be invalid")
	}
}

// TestSignatureSensitivity tests that signature changes when any input changes
func TestSignatureSensitivity(t *testing.T) {
	helper := GetSignatureHelper()

	secret := "test-secret"
	method := "POST"
	path := "/api/v1/test"
	params := map[string]interface{}{"key": "value"}
	timestamp := int64(1703232000)
	nonce := "nonce123"

	baseSignature := helper.GenerateSignature(secret, method, path, params, timestamp, nonce)

	// Change method
	sig1 := helper.GenerateSignature(secret, "GET", path, params, timestamp, nonce)
	if sig1 == baseSignature {
		t.Error("Signature should change when method changes")
	}

	// Change path
	sig2 := helper.GenerateSignature(secret, method, "/api/v2/test", params, timestamp, nonce)
	if sig2 == baseSignature {
		t.Error("Signature should change when path changes")
	}

	// Change params
	sig3 := helper.GenerateSignature(secret, method, path, map[string]interface{}{"key": "different"}, timestamp, nonce)
	if sig3 == baseSignature {
		t.Error("Signature should change when params change")
	}

	// Change timestamp
	sig4 := helper.GenerateSignature(secret, method, path, params, timestamp+1, nonce)
	if sig4 == baseSignature {
		t.Error("Signature should change when timestamp changes")
	}

	// Change nonce
	sig5 := helper.GenerateSignature(secret, method, path, params, timestamp, "different-nonce")
	if sig5 == baseSignature {
		t.Error("Signature should change when nonce changes")
	}
}
