// Package systeminfo reports available host-management backends without mutating the host.
package systeminfo

import (
	"os/exec"
	"runtime"
)

type Capabilities struct {
	Platform string `json:"platform"`
	Services string `json:"services"`
	Firewall string `json:"firewall"`
	Tasks    string `json:"scheduledTasks"`
}

func Detect() Capabilities {
	c := Capabilities{Platform: runtime.GOOS, Services: "unavailable", Firewall: "unavailable", Tasks: "unavailable"}
	switch runtime.GOOS {
	case "linux":
		if _, e := exec.LookPath("systemctl"); e == nil {
			c.Services = "systemd"
		}
		if _, e := exec.LookPath("firewall-cmd"); e == nil {
			c.Firewall = "firewalld"
		}
		if _, e := exec.LookPath("systemctl"); e == nil {
			c.Tasks = "systemd"
		}
	case "windows":
		c.Services = "scm"
		if _, e := exec.LookPath("netsh.exe"); e == nil {
			c.Firewall = "netsh"
		}
		if _, e := exec.LookPath("powershell.exe"); e == nil {
			c.Tasks = "task-scheduler"
		}
	case "darwin":
		if _, e := exec.LookPath("launchctl"); e == nil {
			c.Services = "launchd"
		}
		if _, e := exec.LookPath("pfctl"); e == nil {
			c.Firewall = "pf"
		}
		if _, e := exec.LookPath("launchctl"); e == nil {
			c.Tasks = "launchd"
		}
	}
	return c
}
