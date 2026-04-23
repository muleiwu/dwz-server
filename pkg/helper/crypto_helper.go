package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

// 密文格式:"aesgcm:" + base64(nonce || ciphertext)。
// key 来自 JWT secret 的 SHA-256,避免单独管理一套加密密钥。
const cryptoPrefix = "aesgcm:"

func cryptoKey() ([]byte, error) {
	cfg := GetHelper().GetConfig()
	secret := cfg.GetString("jwt.secret", "")
	if secret == "" {
		return nil, errors.New("jwt.secret 未配置,无法派生加密密钥")
	}
	sum := sha256.Sum256([]byte(secret))
	return sum[:], nil
}

// EncryptSecret 用对称加密保护需要在数据库里长期保存的 secret。
// 空串直接返回空串,方便调用方无感地处理"未配置"情形。
func EncryptSecret(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	key, err := cryptoKey()
	if err != nil {
		return "", err
	}
	return encryptWithKey(key, plaintext)
}

func encryptWithKey(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return cryptoPrefix + base64.StdEncoding.EncodeToString(ct), nil
}

// DecryptSecret 与 EncryptSecret 互为逆操作。对非预期前缀的串视作历史明文直接返回,
// 便于从无加密的旧数据平滑迁移。
func DecryptSecret(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if !strings.HasPrefix(ciphertext, cryptoPrefix) {
		return ciphertext, nil
	}
	key, err := cryptoKey()
	if err != nil {
		return "", err
	}
	return decryptWithKey(key, ciphertext)
}

func decryptWithKey(key []byte, ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, cryptoPrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("密文长度异常")
	}
	nonce, ct := raw[:nonceSize], raw[nonceSize:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}
