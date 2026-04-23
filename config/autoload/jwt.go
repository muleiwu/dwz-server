package autoload

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"cnb.cool/mliev/open/go-web/pkg/helper"
)

type Jwt struct{}

func (Jwt) InitConfig() map[string]any {
	env := helper.GetEnv()
	secret := env.GetString("jwt.secret", "")
	expireHours := env.GetInt("jwt.expire_hours", 24)

	if secret == "" {
		secret = generateRandomSecret(32)
		appendJWTConfig(secret, expireHours)
		fmt.Printf("JWT Secret 自动生成: %s，并写入文件\n", secret)
	}

	return map[string]any{
		"jwt.secret":       secret,
		"jwt.expire_hours": expireHours,
	}
}

func generateRandomSecret(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func appendJWTConfig(secret string, expireHours int) {
	for _, path := range []string{"./config/config.yaml", "./config.yaml"} {
		if _, err := os.Stat(path); err == nil {
			content := fmt.Sprintf("\n# JWT配置（自动生成）\njwt:\n  secret: %s\n  expire_hours: %d\n", secret, expireHours)
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("[JWT]写入配置文件失败: %s\n", err.Error())
				return
			}
			_, _ = f.WriteString(content)
			_ = f.Close()
			return
		}
	}
}
