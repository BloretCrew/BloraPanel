package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"blora.dev/panel/internal/systeminfo"
)

const firewallLeaseNamespace = "firewall_lease"

// firewallLeaseError carries a user-visible phase while preserving the
// underlying error for cancellation and diagnostics.
type firewallLeaseError struct {
	phase string
	err   error
}

func (e *firewallLeaseError) Error() string { return e.err.Error() }
func (e *firewallLeaseError) Unwrap() error { return e.err }

// firewallLeaseRecord is written before any host mutation. Keeping the prior
// and expected sets in the Daemon database lets a restart make a safe rollback
// decision without trusting an in-memory task or a browser connection.
type firewallLeaseRecord struct {
	Snapshot         *systeminfo.FirewallSnapshot `json:"snapshot,omitempty"`
	TaskID           string                       `json:"taskId"`
	ActorID          string                       `json:"actorId"`
	ConfirmRequestID string                       `json:"confirmRequestId,omitempty"`
	Previous         []string                     `json:"previous"`
	Desired          []string                     `json:"desired"`
	ConfirmUntil     time.Time                    `json:"confirmUntil"`
	State            string                       `json:"state"` // prepared, applied, confirmed, rollback_failed
	Error            string                       `json:"error,omitempty"`
}

var errFirewallLeaseConflict = errors.New("firewall lease was already confirmed by another request")

type firewallLease struct {
	record   firewallLeaseRecord
	revision int64
	done     chan struct{}
}

func stableFirewallPorts(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, rule := range in {
		if rule == "" || seen[rule] {
			continue
		}
		seen[rule] = true
		out = append(out, rule)
	}
	sort.Strings(out)
	return out
}

func sameFirewallLeaseSet(a, b []string) bool {
	left, right := stableFirewallPorts(a), stableFirewallPorts(b)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func (d *Daemon) currentFirewallPorts(ctx context.Context) ([]string, error) {
	if d.firewallCurrent != nil {
		return d.firewallCurrent(ctx)
	}
	return systeminfo.CurrentFirewallPorts(ctx)
}

func (d *Daemon) applyFirewall(ctx context.Context, desired []string) error {
	if d.firewallApply != nil {
		return d.firewallApply(ctx, desired)
	}
	return errors.New("firewall mutation requires a durable dual snapshot")
}

func (d *Daemon) rollbackFirewall(ctx context.Context, expected, previous []string) error {
	if d.firewallRollback != nil {
		return d.firewallRollback(ctx, expected, previous)
	}
	return errors.New("firewall rollback requires a durable dual snapshot")
}

func (d *Daemon) authorizeFirewall(ctx context.Context, actor string) error {
	if d.firewallAuthorize != nil {
		return d.firewallAuthorize(ctx, actor)
	}
	return d.authorizeTerminal(ctx, actor, model.ResourceRef{Kind: "node", ID: d.identity.NodeID, NodeID: d.identity.NodeID}, "host.manage")
}

func (d *Daemon) runFirewallApply(ctx context.Context, t model.Task) (any, error) {
	var p struct {
		PlanHash     string    `json:"planHash"`
		Desired      []string  `json:"desired"`
		ConfirmUntil time.Time `json:"confirmUntil"`
	}
	if err := json.Unmarshal(t.Payload, &p); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if p.ConfirmUntil.Before(now) || p.ConfirmUntil.After(now.Add(60*time.Second)) {
		return nil, errors.New("firewall confirmation expired")
	}
	if err := d.authorizeFirewall(ctx, t.ActorID); err != nil {
		return nil, err
	}
	previous, err := d.currentFirewallPorts(ctx)
	if err != nil {
		return nil, err
	}
	record := firewallLeaseRecord{
		TaskID:       t.ID,
		ActorID:      t.ActorID,
		Previous:     stableFirewallPorts(previous),
		Desired:      stableFirewallPorts(p.Desired),
		ConfirmUntil: p.ConfirmUntil.UTC(),
		State:        "prepared",
	}
	if d.firewallCurrent == nil && d.firewallApply == nil {
		snapshot, err := systeminfo.CaptureFirewall(ctx)
		if err != nil {
			return nil, err
		}
		if p.PlanHash == "" || p.PlanHash != systeminfo.FirewallPlanHash(snapshot, p.Desired) {
			return nil, errors.New("firewall preview changed; calculate and confirm a new preview")
		}
		record.Snapshot = &snapshot
		record.Previous = stableFirewallPorts(snapshot.Runtime)
	}
	revision, err := d.store.PutRecord(ctx, firewallLeaseNamespace, t.ID, 0, record)
	if err != nil {
		return nil, fmt.Errorf("persist firewall lease: %w", err)
	}
	lease := &firewallLease{record: record, revision: revision, done: make(chan struct{})}
	d.firewallMu.Lock()
	if _, exists := d.firewallLeases[t.ID]; exists {
		d.firewallMu.Unlock()
		_ = d.store.DeleteRecord(context.Background(), firewallLeaseNamespace, t.ID)
		return nil, errors.New("firewall task already has an active lease")
	}
	d.firewallLeases[t.ID] = lease
	d.firewallMu.Unlock()
	removeLease := func() {
		d.firewallMu.Lock()
		if d.firewallLeases[t.ID] == lease {
			delete(d.firewallLeases, t.ID)
		}
		d.firewallMu.Unlock()
	}
	defer removeLease()

	if record.Snapshot != nil {
		err = systeminfo.ApplyFirewallSnapshot(ctx, *record.Snapshot, record.Desired)
	} else {
		err = d.applyFirewall(ctx, record.Desired)
	}
	if err != nil {
		removeLease()
		// An adapter error can follow a partial mutation or failed compensation.
		// Keep the prepared record for guarded recovery and diagnosis.
		if record.Snapshot != nil {
			if rollbackErr := d.rollbackFirewallLease(lease); rollbackErr != nil {
				return nil, fmt.Errorf("apply firewall: %w; rollback: %v", err, rollbackErr)
			}
		}
		return nil, err
	}
	record.State = "applied"
	revision, err = d.store.PutRecord(ctx, firewallLeaseNamespace, t.ID, revision, record)
	if err != nil {
		// The mutation completed but its lease could not be recorded as applied;
		// use an independent bounded context before reporting the operation.
		rbErr := d.rollbackFirewallLease(lease)
		if rbErr != nil {
			return nil, fmt.Errorf("persist applied firewall lease: %w; rollback: %v", err, rbErr)
		}
		return nil, fmt.Errorf("persist applied firewall lease: %w", err)
	}
	d.firewallMu.Lock()
	lease.record = record
	lease.revision = revision
	d.firewallMu.Unlock()
	if _, err := d.transition(context.Background(), t.ID, model.Running, "awaiting_confirmation", nil, ""); err != nil {
		if rollbackErr := d.rollbackFirewallLease(lease); rollbackErr != nil {
			leaseErr := &firewallLeaseError{phase: "apply_record_unconfirmed_rollback_failed", err: fmt.Errorf("record awaiting confirmation: %w; rollback: %v", err, rollbackErr)}
			d.finishFirewallLeaseTask(t, leaseErr)
			return nil, leaseErr
		}
		leaseErr := &firewallLeaseError{phase: "apply_record_unconfirmed", err: err}
		d.finishFirewallLeaseTask(t, leaseErr)
		return nil, leaseErr
	}

	timer := time.NewTimer(time.Until(record.ConfirmUntil))
	defer timer.Stop()
	select {
	case <-lease.done:
		// confirmFirewallLease persisted the confirmed state before closing done.
		if !d.isFirewallLeaseConfirmed(lease) {
			return nil, &firewallLeaseError{phase: "confirmation_state_lost", err: errors.New("firewall confirmation state was lost")}
		}
		return d.finishConfirmedFirewall(d.confirmedFirewallRecord(lease, record))
	case <-timer.C:
		if d.isFirewallLeaseConfirmed(lease) {
			return d.finishConfirmedFirewall(d.confirmedFirewallRecord(lease, record))
		}
		if err := d.rollbackFirewallLease(lease); err != nil {
			leaseErr := &firewallLeaseError{phase: "confirmation_expired_rollback_failed", err: fmt.Errorf("firewall confirmation expired; rollback failed: %w", err)}
			d.finishFirewallLeaseTask(t, leaseErr)
			return nil, leaseErr
		}
		leaseErr := &firewallLeaseError{phase: "confirmation_expired_rolled_back", err: errors.New("firewall confirmation expired; rules rolled back")}
		d.finishFirewallLeaseTask(t, leaseErr)
		return nil, leaseErr
	case <-ctx.Done():
		if d.isFirewallLeaseConfirmed(lease) {
			return d.finishConfirmedFirewall(d.confirmedFirewallRecord(lease, record))
		}
		if err := d.rollbackFirewallLease(lease); err != nil {
			leaseErr := &firewallLeaseError{phase: "cancelled_rollback_failed", err: fmt.Errorf("firewall lease cancelled; rollback failed: %w", err)}
			d.finishFirewallLeaseTask(t, leaseErr)
			return nil, leaseErr
		}
		leaseErr := &firewallLeaseError{phase: "cancelled_rolled_back", err: ctx.Err()}
		d.finishFirewallLeaseTask(t, leaseErr)
		return nil, leaseErr
	}
}

// confirmedFirewallRecord reads the durable confirmation metadata that may
// have been written by a concurrent HTTP confirmation after runFirewallApply
// captured its initial record. Falling back keeps recovery compatible with
// older callers that construct an already-confirmed lease directly.
func (d *Daemon) confirmedFirewallRecord(lease *firewallLease, fallback firewallLeaseRecord) firewallLeaseRecord {
	d.firewallMu.Lock()
	defer d.firewallMu.Unlock()
	if lease != nil && lease.record.State == "confirmed" {
		return lease.record
	}
	return fallback
}

// runFirewallApply is also exercised directly by durable lease recovery tests.
// Persist the terminal failure here so that a caller which does not go through
// executeAdapter still leaves an auditable task phase. executeAdapter may make
// the same transition once more; the store rejects that terminal no-op.
func (d *Daemon) finishFirewallLeaseTask(t model.Task, leaseErr *firewallLeaseError) {
	if leaseErr == nil || leaseErr.phase == "" {
		return
	}
	state := model.Failed
	if strings.HasPrefix(leaseErr.phase, "cancelled_") {
		state = model.Cancelled
		// A running task must pass through CANCEL_REQUESTED before the store
		// accepts CANCELLED. The normal cancellation path already performs this
		// transition, while direct lease execution/recovery tests may not.
		_, _ = d.transition(context.Background(), t.ID, model.CancelRequested, "cancel_requested", nil, "")
	}
	if _, err := d.transition(context.Background(), t.ID, state, leaseErr.phase, nil, leaseErr.Error()); err == nil {
		var record firewallLeaseRecord
		if _, err := d.store.Record(context.Background(), firewallLeaseNamespace, t.ID, &record); err == nil && record.State == "rolled_back" {
			_ = d.store.DeleteRecord(context.Background(), firewallLeaseNamespace, t.ID)
		}
	}
}

func (d *Daemon) isFirewallLeaseConfirmed(lease *firewallLease) bool {
	d.firewallMu.Lock()
	confirmed := lease.record.State == "confirmed"
	d.firewallMu.Unlock()
	return confirmed
}

func (d *Daemon) rollbackFirewallLease(lease *firewallLease) error {
	d.firewallMu.Lock()
	if lease.record.State == "confirmed" {
		d.firewallMu.Unlock()
		return nil
	}
	record := lease.record
	revision := lease.revision
	record.State = "rollback_in_progress"
	lease.record = record
	d.firewallMu.Unlock()
	// Record the in-progress phase before touching the host. If the Daemon is
	// interrupted while the rollback command is running, startup recovery can
	// distinguish this lease from an unapplied or already confirmed one and
	// retry the guarded rollback.
	nextRevision, persistErr := d.store.PutRecord(context.Background(), firewallLeaseNamespace, record.TaskID, revision, record)
	if persistErr != nil {
		return fmt.Errorf("persist firewall rollback phase: %w", persistErr)
	}
	d.firewallMu.Lock()
	lease.revision = nextRevision
	d.firewallMu.Unlock()
	revision = nextRevision

	rbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var err error
	if record.Snapshot != nil {
		err = systeminfo.RestoreFirewallSnapshot(rbCtx, *record.Snapshot, record.Desired)
	} else if d.firewallRollback == nil && d.firewallCurrent == nil {
		err = errors.New("legacy firewall lease lacks zone and permanent snapshot; manual reconciliation required")
	} else {
		current, currentErr := d.currentFirewallPorts(rbCtx)
		if currentErr != nil || !sameFirewallLeaseSet(current, record.Previous) {
			err = d.rollbackFirewall(rbCtx, record.Desired, record.Previous)
		}
	}
	cancel()
	if err != nil {
		record.State = "rollback_failed"
		record.Error = err.Error()
		if nextRevision, putErr := d.store.PutRecord(context.Background(), firewallLeaseNamespace, record.TaskID, revision, record); putErr == nil {
			d.firewallMu.Lock()
			lease.record = record
			lease.revision = nextRevision
			d.firewallMu.Unlock()
		}
		return err
	}
	record.State = "rolled_back"
	record.Error = ""
	nextRevision, err = d.store.PutRecord(context.Background(), firewallLeaseNamespace, record.TaskID, revision, record)
	if err != nil {
		return fmt.Errorf("persist completed firewall rollback: %w", err)
	}
	d.firewallMu.Lock()
	lease.record, lease.revision = record, nextRevision
	d.firewallMu.Unlock()
	return nil
}

func (d *Daemon) confirmFirewallLease(ctx context.Context, taskID, taskActorID string, requestIDs ...string) (any, error) {
	if taskID == "" || taskActorID == "" {
		return nil, errors.New("firewall task identity required")
	}
	requestID := ""
	if len(requestIDs) > 0 {
		requestID = requestIDs[0]
	}
	d.firewallMu.Lock()
	lease := d.firewallLeases[taskID]
	if lease != nil {
		if lease.record.ActorID != taskActorID {
			d.firewallMu.Unlock()
			return nil, errors.New("firewall lease is not confirmable")
		}
		if lease.record.State == "confirmed" {
			if lease.record.ConfirmRequestID == requestID {
				d.firewallMu.Unlock()
				return map[string]any{"taskId": taskID, "confirmed": true}, nil
			}
			d.firewallMu.Unlock()
			return nil, errFirewallLeaseConflict
		}
		if lease.record.State != "applied" {
			d.firewallMu.Unlock()
			return nil, errors.New("firewall lease is not confirmable")
		}
		if time.Now().UTC().After(lease.record.ConfirmUntil) {
			d.firewallMu.Unlock()
			return nil, errors.New("firewall confirmation expired")
		}
		record := lease.record
		record.State = "confirmed"
		record.ConfirmRequestID = requestID
		revision, err := d.store.PutRecord(ctx, firewallLeaseNamespace, taskID, lease.revision, record)
		if err != nil {
			d.firewallMu.Unlock()
			return nil, err
		}
		lease.record = record
		lease.revision = revision
		close(lease.done)
		d.firewallMu.Unlock()
		return map[string]any{"taskId": taskID, "confirmed": true}, nil
	}
	d.firewallMu.Unlock()

	// A restart may have happened after the host mutation and before the
	// waiting goroutine was recreated. Confirming a durable applied lease is
	// safe; the task itself will still be reconciled by the normal startup path.
	var record firewallLeaseRecord
	revision, err := d.store.Record(ctx, firewallLeaseNamespace, taskID, &record)
	if err != nil {
		return nil, err
	}
	if record.ActorID != taskActorID {
		return nil, errors.New("firewall lease is not confirmable")
	}
	if record.State == "confirmed" {
		if record.ConfirmRequestID == requestID {
			return map[string]any{"taskId": taskID, "confirmed": true}, nil
		}
		return nil, errFirewallLeaseConflict
	}
	if record.State != "applied" {
		return nil, errors.New("firewall lease is not confirmable")
	}
	if time.Now().UTC().After(record.ConfirmUntil) {
		return nil, errors.New("firewall confirmation expired")
	}
	record.State = "confirmed"
	record.ConfirmRequestID = requestID
	if _, err := d.store.PutRecord(ctx, firewallLeaseNamespace, taskID, revision, record); err != nil {
		if errors.Is(err, storage.ErrConflict) {
			var latest firewallLeaseRecord
			if _, readErr := d.store.Record(ctx, firewallLeaseNamespace, taskID, &latest); readErr == nil && latest.State == "confirmed" {
				if latest.ConfirmRequestID == requestID {
					return map[string]any{"taskId": taskID, "confirmed": true}, nil
				}
				return nil, errFirewallLeaseConflict
			}
		}
		return nil, err
	}
	return map[string]any{"taskId": taskID, "confirmed": true}, nil
}

func (d *Daemon) recoverFirewallLeases(ctx context.Context) error {
	ids, err := d.store.RecordIDs(ctx, firewallLeaseNamespace)
	if err != nil {
		return err
	}
	for _, id := range ids {
		var record firewallLeaseRecord
		revision, err := d.store.Record(ctx, firewallLeaseNamespace, id, &record)
		if err != nil {
			continue
		}
		if record.TaskID == "" {
			record.TaskID = id
		}
		switch record.State {
		case "prepared":
			// A crash can occur after the host mutation and before the durable
			// "applied" update. Treat prepared as potentially mutated: discard it
			// only when the host still equals the recorded previous set.
			current, currentErr := d.currentFirewallPorts(ctx)
			if record.Snapshot == nil && d.firewallCurrent != nil && currentErr == nil && sameFirewallLeaseSet(current, record.Previous) {
				if err := d.finishRecoveredFirewallRollback(record); err != nil {
					return err
				}
				continue
			}
			lease := &firewallLease{record: record, revision: revision, done: make(chan struct{})}
			if err := d.rollbackFirewallLease(lease); err != nil {
				continue
			}
			if err := d.finishRecoveredFirewallRollback(record); err != nil {
				return err
			}
		case "confirmed":
			// Keep recovery evidence until the task terminal state is durable.
			if _, err := d.finishConfirmedFirewall(record); err != nil {
				return fmt.Errorf("recover confirmed firewall task: %w", err)
			}
		case "rolled_back":
			if err := d.finishRecoveredFirewallRollback(record); err != nil {
				return err
			}
		case "applied", "rollback_failed", "rollback_in_progress":
			lease := &firewallLease{record: record, revision: revision, done: make(chan struct{})}
			if err := d.rollbackFirewallLease(lease); err != nil {
				// Keep a bounded diagnostic record for the next restart and for
				// operators; no host mutation is retried without this check.
				continue
			}
			if err := d.finishRecoveredFirewallRollback(record); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Daemon) finishRecoveredFirewallRollback(record firewallLeaseRecord) error {
	task, err := d.store.Task(context.Background(), record.TaskID)
	if err != nil {
		return err
	}
	if !task.State.Terminal() {
		state := model.Failed
		if task.State == model.CancelRequested {
			state = model.Cancelled
		}
		if _, err := d.transition(context.Background(), task.ID, state, "rolled_back_after_restart", nil, "防火墙未确认变更已恢复原规则"); err != nil {
			return err
		}
	} else if task.State == model.Succeeded {
		return errors.New("rolled back firewall task conflicts with success")
	}
	return d.store.DeleteRecord(context.Background(), firewallLeaseNamespace, record.TaskID)
}

// Persist success before removing the confirmation recovery record.
func (d *Daemon) finishConfirmedFirewall(record firewallLeaseRecord) (any, error) {
	result := map[string]any{"rules": len(record.Desired), "confirmed": true, "confirmRequestId": record.ConfirmRequestID}
	task, err := d.store.Task(context.Background(), record.TaskID)
	if err != nil {
		return nil, &firewallLeaseError{phase: "confirmation_commit_pending", err: err}
	}
	if !task.State.Terminal() {
		body, _ := json.Marshal(result)
		if _, err = d.transition(context.Background(), record.TaskID, model.Succeeded, "confirmed", body, ""); err != nil {
			return nil, &firewallLeaseError{phase: "confirmation_commit_pending", err: err}
		}
	} else if task.State != model.Succeeded {
		return nil, errors.New("confirmed firewall task has a conflicting terminal state")
	}
	// Failure to prune is safe: startup retries without touching host rules.
	_ = d.store.DeleteRecord(context.Background(), firewallLeaseNamespace, record.TaskID)
	return result, nil
}
