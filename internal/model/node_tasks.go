package model

// IsNodeTaskAction identifies tasks whose resource is the host node itself.
// Authorization remains a separate check at acceptance and execution time.
func IsNodeTaskAction(action string) bool {
	if IsContainerAction(action) {
		return true
	}
	switch action {
	case "terminal.create", "terminal.close", "extension.task", "process.terminate",
		"system.service.start", "system.service.stop", "system.service.restart",
		"system.task.enable", "system.task.disable", "system.firewall.apply":
		return true
	}
	return false
}
