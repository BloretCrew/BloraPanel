// Package backup stores bounded, verifiable local snapshots and restore plans.
// Callers supply an already-authorized filesystem service; browser host paths
// never select the instance root or the private repository.
package backup

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"blora.dev/panel/internal/filesystem"
	"blora.dev/panel/internal/model"
)

var (
	ErrConflict    = errors.New("backup version or identity conflict")
	ErrLimit       = errors.New("backup budget exceeded")
	ErrInvalid     = errors.New("invalid backup or restore specification")
	ErrCorrupt     = errors.New("backup content or manifest verification failed")
	ErrInterrupted = errors.New("backup operation outcome requires reconciliation")
	ErrUnsupported = errors.New("backup capability not available")
)

type Options struct {
	StateDir              string
	MaxEntries            int
	MaxFileBytes          int64
	MaxTotalBytes         int64
	MaxArchiveBytes       int64
	MaxRepositoryBytes    int64
	MaxSnapshots          int
	MaxConcurrent         int
	Hooks                 HookRunner
	AuthorizeSnapshotRead func(context.Context, string, model.ResourceRef) error
	AuthorizeRestoreWrite func(context.Context, string, model.ResourceRef) error
}

type Consistency struct {
	Mode   string `json:"mode"`             // files, save, pause, stop, hooks
	Before string `json:"before,omitempty"` // configured capability/command reference
	After  string `json:"after,omitempty"`
}
type HookCall struct {
	ID       string            `json:"id"`
	BackupID string            `json:"backupId"`
	OwnerID  string            `json:"ownerId"`
	Resource model.ResourceRef `json:"resource"`
	Strategy Consistency       `json:"strategy"`
	Phase    string            `json:"phase"`
}
type HookReceipt struct {
	ID     string    `json:"id"`
	State  string    `json:"state"` // succeeded, failed, running, unknown
	Detail string    `json:"detail,omitempty"`
	At     time.Time `json:"at"`
}

// Run must durably accept this ID before its side effect. Reconcile observes
// the original operation and must not silently run it again.
type HookRunner interface {
	Run(context.Context, HookCall) (HookReceipt, error)
	Reconcile(context.Context, HookCall) (HookReceipt, error)
}

type CreateSpec struct {
	ID          string            `json:"id"`
	OwnerID     string            `json:"ownerId"`
	Source      model.ResourceRef `json:"source"`
	Path        string            `json:"path"`
	Version     string            `json:"version"`
	Compression string            `json:"compression"` // store or deflate
	Consistency Consistency       `json:"consistency"`
}
type Entry struct {
	Path     string    `json:"path"`
	Kind     string    `json:"kind"`
	Size     int64     `json:"size"`
	Hash     string    `json:"hash,omitempty"`
	Mode     uint32    `json:"mode"`
	Modified time.Time `json:"modified"`
}
type Manifest struct {
	Format          string     `json:"format"`
	ID              string     `json:"id"`
	Spec            CreateSpec `json:"spec"`
	CapturedVersion string     `json:"capturedVersion"`
	SourceRootID    string     `json:"sourceRootId"`
	SourceObjectID  string     `json:"sourceObjectId"`
	Created         time.Time  `json:"created"`
	Entries         []Entry    `json:"entries"`
	Total           int64      `json:"total"`
}
type Snapshot struct {
	Manifest        Manifest     `json:"manifest"`
	ArchiveHash     string       `json:"archiveHash"`
	ArchiveObjectID string       `json:"archiveObjectId"`
	ArchiveBytes    int64        `json:"archiveBytes"`
	State           string       `json:"state"`
	Before          *HookReceipt `json:"before,omitempty"`
	After           *HookReceipt `json:"after,omitempty"`
}
type SnapshotInfo struct {
	ID              string            `json:"id"`
	OwnerID         string            `json:"ownerId"`
	Source          model.ResourceRef `json:"source"`
	Path            string            `json:"path"`
	CapturedVersion string            `json:"capturedVersion"`
	Created         time.Time         `json:"created"`
	Entries         int               `json:"entries"`
	Total           int64             `json:"total"`
	ArchiveHash     string            `json:"archiveHash"`
	ArchiveBytes    int64             `json:"archiveBytes"`
	State           string            `json:"state"`
	Consistency     Consistency       `json:"consistency"`
	Before          *HookReceipt      `json:"before,omitempty"`
	After           *HookReceipt      `json:"after,omitempty"`
}
type Result struct {
	ID             string        `json:"id"`
	Stage          string        `json:"stage"`
	Completed      int           `json:"completed"`
	CurrentPath    string        `json:"currentPath,omitempty"`
	Bytes          int64         `json:"bytes"`
	Snapshot       *SnapshotInfo `json:"snapshot,omitempty"`
	Partial        bool          `json:"partial"`
	Unknown        bool          `json:"unknown"`
	CleanupPending bool          `json:"cleanupPending"`
	Error          string        `json:"error,omitempty"`
}

type RestoreRequest struct {
	ID       string            `json:"id"`
	OwnerID  string            `json:"ownerId"`
	BackupID string            `json:"backupId"`
	Target   model.ResourceRef `json:"target"`
	Path     string            `json:"path"`
	Version  string            `json:"version"`
}
type RestoreEntry struct {
	Entry
	TargetPath     string `json:"targetPath"`
	TargetVersion  string `json:"targetVersion"`
	TargetObjectID string `json:"targetObjectId,omitempty"`
}
type RestorePlan struct {
	Request         RestoreRequest `json:"request"`
	Hash            string         `json:"hash"`
	ArchiveHash     string         `json:"archiveHash"`
	ArchiveObjectID string         `json:"archiveObjectId"`
	TargetRootID    string         `json:"targetRootId"`
	TargetObjectID  string         `json:"targetObjectId,omitempty"`
	Entries         []RestoreEntry `json:"entries"`
	Total           int64          `json:"total"`
	Created         time.Time      `json:"created"`
}
type RestoreSpec struct {
	ID       string `json:"id"`
	OwnerID  string `json:"ownerId"`
	PlanID   string `json:"planId"`
	PlanHash string `json:"planHash"`
	// Every existing file must be explicitly confirmed with the planned hash.
	Overwrite map[string]string `json:"overwrite,omitempty"`
	// A bounded confirmation of every existing file in this exact immutable
	// plan. The API must obtain this hash from an explicit user confirmation;
	// it must never populate it automatically as an unscoped force option.
	OverwritePlanHash string `json:"overwritePlanHash,omitempty"`
}

type Retention struct {
	OwnerID  string            `json:"ownerId"`
	Source   model.ResourceRef `json:"source"`
	Path     string            `json:"path"`
	KeepLast int               `json:"keepLast"`
	MaxAge   time.Duration     `json:"maxAge"`
	MaxBytes int64             `json:"maxBytes"`
}

// TaskPayload carries the authoritative instance configuration frozen by
// Master task admission. Daemon supplies the actor/resource/task identity and
// obtains the corresponding root through its reference-counted filesFor hook.
type TaskPayload struct {
	Config    model.InstanceConfig `json:"config"`
	Create    *CreateSpec          `json:"create,omitempty"`
	Restore   *RestoreSpec         `json:"restore,omitempty"`
	Retention *Retention           `json:"retention,omitempty"`
}

type createCheckpoint struct {
	Spec            CreateSpec               `json:"spec"`
	Result          Result                   `json:"result"`
	Manifest        *Manifest                `json:"manifest,omitempty"`
	Before          *HookReceipt             `json:"before,omitempty"`
	After           *HookReceipt             `json:"after,omitempty"`
	ArchiveHash     string                   `json:"archiveHash,omitempty"`
	ArchiveObjectID string                   `json:"archiveObjectId,omitempty"`
	ArchiveBytes    int64                    `json:"archiveBytes,omitempty"`
	SourceRelation  filesystem.RelationFacts `json:"sourceRelation"`
}

type restoreCheckpoint struct {
	Spec           RestoreSpec            `json:"spec"`
	Result         Result                 `json:"result"`
	Cursor         int                    `json:"cursor"`
	MetadataCursor int                    `json:"metadataCursor"`
	Operation      string                 `json:"operation,omitempty"`
	Upload         *filesystem.UploadSpec `json:"upload,omitempty"`
	LastReceipt    json.RawMessage        `json:"lastReceipt,omitempty"`
	DirectoryIDs   map[string]string      `json:"directoryIds,omitempty"`
	FileIDs        map[string]string      `json:"fileIds,omitempty"`
}
