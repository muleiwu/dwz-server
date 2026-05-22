package service

import (
	"strings"

	"github.com/mileusna/useragent"
)

type trafficMetadata struct {
	DeviceType string
	Browser    string
	OS         string
	IsBot      bool
	BotName    string
}

func parseTrafficMetadata(userAgentString string) trafficMetadata {
	if strings.TrimSpace(userAgentString) == "" {
		return trafficMetadata{
			DeviceType: "unknown",
			Browser:    "unknown",
			OS:         "unknown",
		}
	}

	ua := useragent.Parse(userAgentString)
	deviceType := "unknown"
	switch {
	case ua.Bot:
		deviceType = "bot"
	case ua.Mobile:
		deviceType = "mobile"
	case ua.Tablet:
		deviceType = "tablet"
	case ua.Desktop:
		deviceType = "desktop"
	}

	browser := ua.Name
	if browser == "" {
		browser = "unknown"
	}
	osName := ua.OS
	if osName == "" {
		osName = "unknown"
	}
	botName := ""
	if ua.Bot {
		botName = browser
		if botName == "unknown" {
			botName = "bot"
		}
	}

	return trafficMetadata{
		DeviceType: deviceType,
		Browser:    browser,
		OS:         osName,
		IsBot:      ua.Bot,
		BotName:    botName,
	}
}
