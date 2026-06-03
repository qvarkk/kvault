package web

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// deviceLabel derives a human-readable default label ("Chrome on Windows") from
// the request's User-Agent so a freshly issued key is recognizable in the key
// list. Users can rename it later. Falls back to "Unknown device".
func deviceLabel(ctx *gin.Context) string {
	ua := ctx.GetHeader("User-Agent")
	browser := uaBrowser(ua)
	os := uaOS(ua)

	switch {
	case browser != "" && os != "":
		return browser + " on " + os
	case browser != "":
		return browser
	case os != "":
		return os
	default:
		return "Unknown device"
	}
}

func uaBrowser(ua string) string {
	switch {
	// Order matters: Edge/Chrome UAs also contain "Safari"; Edge contains "Chrome".
	case strings.Contains(ua, "Edg"):
		return "Edge"
	case strings.Contains(ua, "OPR") || strings.Contains(ua, "Opera"):
		return "Opera"
	case strings.Contains(ua, "Firefox"):
		return "Firefox"
	case strings.Contains(ua, "Chrome"):
		return "Chrome"
	case strings.Contains(ua, "Safari"):
		return "Safari"
	default:
		return ""
	}
}

func uaOS(ua string) string {
	switch {
	case strings.Contains(ua, "iPhone"):
		return "iPhone"
	case strings.Contains(ua, "iPad"):
		return "iPad"
	case strings.Contains(ua, "Android"):
		return "Android"
	case strings.Contains(ua, "Windows"):
		return "Windows"
	case strings.Contains(ua, "Mac OS X") || strings.Contains(ua, "Macintosh"):
		return "macOS"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	default:
		return ""
	}
}
