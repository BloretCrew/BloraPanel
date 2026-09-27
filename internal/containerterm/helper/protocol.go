// Package helper is the narrowly scoped, unprivileged in-container exec helper.
// It never exposes a shell on the host or accepts a host PID from Docker inspect.
package helper

const Path = "/usr/local/lib/blora/exec-helper"
const Version = 1
const Prefix = "\x1eBLORA-EXEC/1 "
const TokenEnvironment = "BLORA_EXEC_TOKEN"

type Identity struct {
	Version    int    `json:"version"`
	PID        int    `json:"pid"`
	StartTicks uint64 `json:"startTicks"`
	SessionID  int    `json:"sessionId"`
	Token      string `json:"token"`
}
type Probe struct {
	Version     int  `json:"version"`
	PIDFD       bool `json:"pidfd"`
	ProcessTree bool `json:"processTree"`
}
