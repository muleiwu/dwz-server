package autoload

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	envInterface "cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
)

type Jwt struct {
}

func (receiver Jwt) InitConfig(helper envInterface.HelperInterface) map[string]any {
	secret := helper.GetEnv().GetString("jwt.secret", "")
	expireHours := helper.GetEnv().GetInt("jwt.expire_hours", 24)

	if secret == "" {
		secret = generateRandomSecret(32)
		appendJWTConfig(secret, expireHours)
	}

	return map[string]any{
		"jwt.secret":       secret,
		"jwt.expire_hours": expireHours,
	}
}

// generateRandomSecret 生成随机密钥
func generateRandomSecret(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// appendJWTConfig 将JWT配置追加到配置文件末尾
func appendJWTConfig(secret string, expireHours int) {
	for _, path := range []string{"./config/config.yaml", "./config.yaml"} {
		if _, err := os.Stat(path); err == nil {
			content := fmt.Sprintf("\n# JWT配置（自动生成）\njwt:\n  secret: %s\n  expire_hours: %d\n", secret, expireHours)
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				f.WriteString(content)
				f.Close()
			}
			return
		}
	}
}
