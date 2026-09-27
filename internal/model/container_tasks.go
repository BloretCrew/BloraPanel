package model

func IsContainerAction(action string) bool {
	switch action {
	case "container.create", "container.start", "container.stop", "container.restart", "container.delete", "image.pull", "image.delete", "volume.create", "volume.delete", "network.create", "network.delete", "compose.apply", "compose.delete", "compose.save":
		return true
	}
	return false
}
