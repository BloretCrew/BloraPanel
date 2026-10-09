// Package coreupdate prepares immutable release installations without changing the
// running installation. The service owns the final resource-preserving handoff.
package coreupdate

import (
	"context"
	"errors"
	"regexp"
	"runtime/debug"
	"time"
)

// These variables are set by release builds using -X.
var Version = "development"
var Revision = ""

const DefaultRepository = "https://github.com/BloretCrew/BloraPanel"
const DefaultChannel = "beta"

var (
	ErrBusy            = errors.New("an update operation is already running")
	ErrPreviewRequired = errors.New("check the exact target revision before applying it")
	ErrIncompatible    = errors.New("the selected revision requires a maintenance update")
	revisionPattern    = regexp.MustCompile(`^[a-f0-9]{40}$`)
)

type Config struct {
	Repository string `json:"repository"`
	Channel    string `json:"channel"`
	APIURL     string `json:"apiUrl,omitempty"`
}

type Manifest struct {
	FormatVersion     int    `json:"formatVersion"`
	Version           string `json:"version"`
	ProtocolVersion   uint32 `json:"protocolVersion"`
	SchemaVersion     int    `json:"schemaVersion"`
	PeerProtocolMin   uint32 `json:"peerProtocolMin"`
	PeerProtocolMax   uint32 `json:"peerProtocolMax"`
	PreserveInstances bool   `json:"preserveInstances"`
	SchemaFingerprint string `json:"schemaFingerprint"`
	Revision          string `json:"revision"`
}

func (m Manifest) SupportsPeer(protocol uint32) bool {
	return m.PeerProtocolMin > 0 && protocol >= m.PeerProtocolMin && protocol <= m.PeerProtocolMax
}

type Preview struct {
	Revision        string    `json:"revision"`
	CurrentRevision string    `json:"currentRevision"`
	Version         string    `json:"version"`
	Compatible      bool      `json:"compatible"`
	Reason          string    `json:"reason,omitempty"`
	UpToDate        bool      `json:"upToDate"`
	CheckedAt       time.Time `json:"checkedAt"`
	Manifest        Manifest  `json:"manifest"`
	ReleaseURL      string    `json:"releaseUrl,omitempty"`
	Tag             string    `json:"tag,omitempty"`
	DownloadBytes   int64     `json:"downloadBytes,omitempty"`
}

type Prepared struct {
	Revision     string   `json:"revision"`
	Binary       string   `json:"binary"`
	Directory    string   `json:"directory"`
	WebDirectory string   `json:"webDirectory,omitempty"`
	Manifest     Manifest `json:"manifest"`
}

type Job struct {
	ID              string    `json:"id"`
	Revision        string    `json:"revision"`
	State           string    `json:"state"`
	Phase           string    `json:"phase"`
	Detail          string    `json:"detail,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Prepared        *Prepared `json:"prepared,omitempty"`
	DownloadedBytes int64     `json:"downloadedBytes,omitempty"`
	TotalBytes      int64     `json:"totalBytes,omitempty"`
}

type Status struct {
	Component  string   `json:"component"`
	Version    string   `json:"version"`
	Revision   string   `json:"revision"`
	Source     Config   `json:"source"`
	Platform   string   `json:"platform"`
	Job        *Job     `json:"job,omitempty"`
	Preview    *Preview `json:"preview,omitempty"`
	Checking   bool     `json:"checking"`
	CheckError string   `json:"checkError,omitempty"`
}

type Options struct {
	Root            string
	Component       string
	InstallRoot     string
	CurrentRevision string
	Config          Config
	CurrentVersion  string
	// Ready is called only after the prebuilt archive and runtime are verified.
	// It must leave active resources intact or return an error without switching.
	Ready func(context.Context, Prepared) error
}

func BuildRevision() string {
	if revisionPattern.MatchString(Revision) {
		return Revision
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && revisionPattern.MatchString(setting.Value) {
				return setting.Value
			}
		}
	}
	return ""
}
