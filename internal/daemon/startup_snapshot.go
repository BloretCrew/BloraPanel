package daemon

import (
	"context"
	"encoding/json"
	"sync"

	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/protocol"
)

func (d *Daemon) startupSnapshot(ctx context.Context, c *protocol.Conn) error {
	ids, err := d.store.RecordIDs(ctx, "instance")
	if err != nil {
		return err
	}
	for _, id := range ids {
		d.mu.Lock()
		lock := d.locks[id]
		if lock == nil {
			lock = &sync.Mutex{}
			d.locks[id] = lock
		}
		d.mu.Unlock()
		lock.Lock()
		var i model.Instance
		_, err := d.store.Record(ctx, "instance", id, &i)
		if err == nil && i.RunID != "" {
			_, observation, observeErr := d.runtime.Recover(ctx, i.RunID)
			state := observation.State
			if observeErr != nil {
				state = "UNKNOWN"
			} else if observation.Exited {
				state = "STOPPED"
				d.finishLog(i.RunID)
			}
			if state != i.State {
				i.State = state
				err = d.saveInstance(ctx, &i)
			}
		}
		lock.Unlock()
		if err != nil {
			return err
		}
		b, _ := json.Marshal(map[string]any{"instances": []model.Instance{i}})
		if err := c.Send(ctx, protocol.Envelope{Type: protocol.TypeSnapshot, Payload: b}); err != nil {
			return err
		}
	}
	b, _ := json.Marshal(map[string]any{"instances": []model.Instance{}, "startupReady": true, "startupId": d.startupID})
	return c.Send(ctx, protocol.Envelope{Type: protocol.TypeSnapshot, Payload: b})
}
