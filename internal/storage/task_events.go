package storage

import (
	"context"
	"errors"
)

type TaskStageEvent struct {
	Revision   int64  `json:"revision"`
	RecordedAt int64  `json:"recordedAt"`
	State      string `json:"state"`
	Phase      string `json:"phase"`
	Error      string `json:"error,omitempty"`
	Truncated  bool   `json:"truncated"`
}

// Project only stage diagnostics; execution payloads and result blobs never
// enter this bounded history response. Revisions are durable per-task cursors.
func (s *Store) TaskStagePage(ctx context.Context, id string, before int64, limit int) ([]TaskStageEvent, bool, error) {
	if id == "" || before < 0 || limit < 1 || limit > 100 {
		return nil, false, errors.New("invalid task stage page")
	}
	query := `SELECT revision,recorded_at,json_extract(document,'$.state'),substr(COALESCE(json_extract(document,'$.phase'),''),1,512),substr(COALESCE(json_extract(document,'$.error'),''),1,4096),length(COALESCE(json_extract(document,'$.phase'),''))>512 OR length(COALESCE(json_extract(document,'$.error'),''))>4096 FROM task_events WHERE task_id=?`
	args := []any{id}
	if before > 0 {
		query += " AND revision<?"
		args = append(args, before)
	}
	args = append(args, limit+1)
	rows, err := s.DB.QueryContext(ctx, query+" ORDER BY revision DESC LIMIT ?", args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	items := []TaskStageEvent{}
	for rows.Next() {
		var item TaskStageEvent
		if err := rows.Scan(&item.Revision, &item.RecordedAt, &item.State, &item.Phase, &item.Error, &item.Truncated); err != nil {
			return nil, false, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	return items, more, nil
}
