package model

import (
	"encoding/json"
	"errors"
	"strings"
)

type InstanceSettings struct {
	Revision int64           `json:"revision"`
	Name     *string         `json:"name,omitempty"`
	Group    *string         `json:"group,omitempty"`
	Tags     *[]string       `json:"tags,omitempty"`
	Config   *InstanceConfig `json:"config,omitempty"`
}

// NormalizeInstanceConfig applies the same limits to creation and editing.
// A configured executable is never run while validating or saving a template.
func NormalizeInstanceConfig(c *InstanceConfig) error {
	if c.Mode != "native" && c.Mode != "container" {
		return errors.New("运行模式必须是 native 或 container")
	}
	if len(c.Command) == 0 || len(c.Command) > 128 || strings.TrimSpace(c.Command[0]) == "" {
		return errors.New("需要启动命令，最多128个参数")
	}
	if c.StopSeconds < 1 {
		c.StopSeconds = 30
	}
	if c.KillSeconds < 1 {
		c.KillSeconds = 10
	}
	if c.StopSeconds > 600 || c.KillSeconds > 120 || len(c.StopInput) > 8192 {
		return errors.New("停止期限或输入超出上限")
	}
	if c.MemoryBytes < 0 || c.CPUQuota < 0 || c.PidsLimit < 0 || c.PidsLimit > 1048576 {
		return errors.New("资源限制不能为负数，进程数上限为1048576")
	}
	if len(c.Environment) > 128 {
		return errors.New("环境变量最多128项")
	}
	for _, arg := range append(append([]string{}, c.Command...), c.Directory, c.Image, c.StopInput) {
		if strings.ContainsRune(arg, 0) {
			return errors.New("配置不能包含空字符")
		}
	}
	for key, value := range c.Environment {
		if key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(value, 0) {
			return errors.New("环境变量名称或值无效")
		}
	}
	if c.Mode == "container" {
		if c.Image == "" || c.UID == nil || *c.UID == 0 || c.GID == nil || *c.GID == 0 {
			return errors.New("隔离容器需要镜像及非root UID/GID")
		}
		if c.Directory != "" && c.Directory != "/workspace" {
			return errors.New("隔离容器的工作目录固定为 /workspace")
		}
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if len(b) > 24<<10 {
		return errors.New("实例配置总计不能超过24 KiB")
	}
	return nil
}

func LaunchConfiguration(c InstanceConfig) string {
	c.StopSeconds, c.KillSeconds, c.StopInput, c.Escalate, c.Autostart = 0, 0, "", false, false
	b, _ := json.Marshal(c)
	return string(b)
}

type InstanceTemplate struct {
	ID          string         `json:"templateId"`
	Name        string         `json:"name"`
	Platform    string         `json:"platform"`
	Description string         `json:"description"`
	Config      InstanceConfig `json:"config"`
}

func InstanceTemplates() []InstanceTemplate {
	return []InstanceTemplate{
		{ID: "native-linux", Name: "Linux 通用程序", Platform: "linux", Description: "填写已安装程序路径与参数；原生程序需主机管理权限。", Config: InstanceConfig{Mode: "native", Command: []string{"/path/to/program"}, StopSeconds: 30, KillSeconds: 10, Escalate: true}},
		{ID: "native-windows", Name: "Windows 通用程序", Platform: "windows", Description: "填写可执行文件完整路径与参数；由 Windows Job 管理归属进程。", Config: InstanceConfig{Mode: "native", Command: []string{`C:\path\program.exe`}, StopSeconds: 30, KillSeconds: 10, Escalate: true}},
		{ID: "java-game", Name: "Java 游戏服务", Platform: "any", Description: "预填 jar 启动与 stop 换行退出；需自行提供Java、服务端文件并调整参数。", Config: InstanceConfig{Mode: "native", Command: []string{"java", "-Xms1G", "-Xmx2G", "-jar", "server.jar", "nogui"}, StopSeconds: 120, KillSeconds: 10, StopInput: "stop\n", Escalate: true}},
		{ID: "isolated-container", Name: "隔离容器", Platform: "linux", Description: "选择包含固定执行辅助程序的镜像，指定非root UID/GID。", Config: InstanceConfig{Mode: "container", Command: []string{"/path/to/program"}, Directory: "/workspace", StopSeconds: 30, KillSeconds: 10, Escalate: true}},
	}
}
