package daemon

import "strings"

func managementWebSocketURL(value string) string {
	value = strings.TrimRight(value, "/")
	if strings.HasPrefix(value, "https://") {
		return "wss://" + strings.TrimPrefix(value, "https://")
	}
	return "ws://" + strings.TrimPrefix(value, "http://")
}
