package helper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// SignatureHelper 签名助手
// 提供 HMAC-SHA256 签名生成、验证以及 App Secret 加密/解密功能
type SignatureHelper struct {
	// encryptionKey 用于 AES 加密 App Secret 的密钥
	// 必须是 16, 24, 或 32 字节长度（对应 AES-128, AES-192, AES-256）
	encryptionKey []byte
}

// 默认加密密钥（生产环境应通过配置文件设置）
// 必须是 32 字节长度（AES-256）
var defaultEncryptionKey = []byte("dwz-server-secret-key-32bytes!!!")

// signatureHelper 单例
var signatureHelper *SignatureHelper

// GetSignatureHelper 获取签名助手单例
func GetSignatureHelper() *SignatureHelper {
	if signatureHelper == nil {
		signatureHelper = &SignatureHelper{
			encryptionKey: defaultEncryptionKey,
		}
	}
	return signatureHelper
}

// SetEncryptionKey 设置加密密钥
// key 必须是 16, 24, 或 32 字节长度
func (s *SignatureHelper) SetEncryptionKey(key []byte) error {
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return errors.New("encryption key must be 16, 24, or 32 bytes")
	}
	s.encryptionKey = key
	return nil
}

// GenerateSignature 生成 HMAC-SHA256 签名
// 签名计算公式: HMAC-SHA256(secret, method + path + sorted_params_json + timestamp + nonce)
//
// 参数:
//   - secret: App Secret（明文）
//   - method: HTTP 方法（GET, POST, PUT, DELETE 等）
//   - path: 请求路径（如 /api/v1/shortlinks）
//   - params: 请求参数（会按 key 排序后转为 JSON）
//   - timestamp: Unix 时间戳（秒）
//   - nonce: 随机字符串
//
// 返回:
//   - 签名的十六进制字符串
func (s *SignatureHelper) GenerateSignature(secret, method, path string, params map[string]interface{}, timestamp int64, nonce string) string {
	// 构建待签名字符串
	stringToSign := s.buildStringToSign(method, path, params, timestamp, nonce)

	// 使用 HMAC-SHA256 计算签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// VerifySignature 验证签名
// 使用相同的参数重新计算签名，并与提供的签名进行比对
//
// 参数:
//   - secret: App Secret（明文）
//   - method: HTTP 方法
//   - path: 请求路径
//   - params: 请求参数
//   - timestamp: Unix 时间戳（秒）
//   - nonce: 随机字符串
//   - signature: 待验证的签名
//
// 返回:
//   - true 如果签名匹配，false 否则
func (s *SignatureHelper) VerifySignature(secret, method, path string, params map[string]interface{}, timestamp int64, nonce, signature string) bool {
	expectedSignature := s.GenerateSignature(secret, method, path, params, timestamp, nonce)
	// 使用 hmac.Equal 进行常量时间比较，防止时序攻击
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// buildStringToSign 构建待签名字符串
// 格式: method + path + sorted_params_json + timestamp + nonce
func (s *SignatureHelper) buildStringToSign(method, path string, params map[string]interface{}, timestamp int64, nonce string) string {
	var builder strings.Builder

	// 1. HTTP 方法（大写）
	builder.WriteString(strings.ToUpper(method))

	// 2. 请求路径
	builder.WriteString(path)

	// 3. 排序后的参数 JSON
	sortedParamsJSON := s.sortAndSerializeParams(params)
	builder.WriteString(sortedParamsJSON)

	// 4. 时间戳
	builder.WriteString(fmt.Sprintf("%d", timestamp))

	// 5. 随机数
	builder.WriteString(nonce)

	return builder.String()
}

// sortAndSerializeParams 将参数按 key 排序后序列化为 JSON 字符串
func (s *SignatureHelper) sortAndSerializeParams(params map[string]interface{}) string {
	if len(params) == 0 {
		return "{}"
	}

	// 获取所有 key 并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建有序的 map（使用 json.Marshal 保证顺序）
	orderedMap := make([]struct {
		Key   string
		Value interface{}
	}, len(keys))

	for i, k := range keys {
		orderedMap[i] = struct {
			Key   string
			Value interface{}
		}{k, params[k]}
	}

	// 手动构建 JSON 字符串以保证顺序（禁用 HTML 转义以兼容其他语言）
	var builder strings.Builder
	builder.WriteString("{")
	for i, item := range orderedMap {
		if i > 0 {
			builder.WriteString(",")
		}
		// 使用 Encoder 禁用 HTML 转义，避免 &<> 等字符被转义为 \uXXXX
		var keyBuf, valueBuf bytes.Buffer
		keyEncoder := json.NewEncoder(&keyBuf)
		keyEncoder.SetEscapeHTML(false)
		keyEncoder.Encode(item.Key)

		valueEncoder := json.NewEncoder(&valueBuf)
		valueEncoder.SetEscapeHTML(false)
		valueEncoder.Encode(item.Value)

		// Encoder.Encode 会添加换行符，需要 TrimSpace
		builder.WriteString(strings.TrimSpace(keyBuf.String()))
		builder.WriteString(":")
		builder.WriteString(strings.TrimSpace(valueBuf.String()))
	}
	builder.WriteString("}")

	return builder.String()
}

// GenerateAppID 生成唯一的 App ID
// 格式: app_ + 16字节随机hex字符串 (共36字符)
func (s *SignatureHelper) GenerateAppID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for AppID: %w", err)
	}
	return fmt.Sprintf("app_%s", hex.EncodeToString(bytes)), nil
}

// GenerateAppSecret 生成安全的 App Secret
// 生成32字节随机hex字符串 (64字符)
func (s *SignatureHelper) GenerateAppSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for AppSecret: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// EncryptAppSecret 使用 AES-GCM 加密 App Secret
// 返回 base64 编码的加密数据（包含 nonce）
//
// 参数:
//   - secret: 明文 App Secret
//
// 返回:
//   - 加密后的 base64 字符串
//   - 错误信息
func (s *SignatureHelper) EncryptAppSecret(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("secret cannot be empty")
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密数据（nonce 会被附加到密文前面）
	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)

	// 返回 base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAppSecret 解密 App Secret
// 输入 base64 编码的加密数据，返回明文
//
// 参数:
//   - encrypted: base64 编码的加密数据
//
// 返回:
//   - 解密后的明文 App Secret
//   - 错误信息
func (s *SignatureHelper) DecryptAppSecret(encrypted string) (string, error) {
	if encrypted == "" {
		return "", errors.New("encrypted data cannot be empty")
	}

	// Base64 解码
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 检查密文长度
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// 分离 nonce 和实际密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// ValidateTimestamp 验证时间戳是否在有效窗口内
// 默认窗口为 ±5 分钟（300秒）
//
// 参数:
//   - timestamp: 请求中的 Unix 时间戳（秒）
//   - currentTime: 当前服务器时间的 Unix 时间戳（秒）
//
// 返回:
//   - true 如果时间戳在有效窗口内，false 否则
func (s *SignatureHelper) ValidateTimestamp(timestamp, currentTime int64) bool {
	const maxTimeDiff = 300 // 5 分钟 = 300 秒

	diff := currentTime - timestamp
	if diff < 0 {
		diff = -diff
	}

	return diff <= maxTimeDiff
}

// ValidateNonce 验证 nonce 是否有效（非空）
//
// 参数:
//   - nonce: 随机字符串
//
// 返回:
//   - true 如果 nonce 有效，false 否则
func (s *SignatureHelper) ValidateNonce(nonce string) bool {
	return nonce != ""
}
