package master

import (
	"blora.dev/panel/internal/model"
	"context"
	"encoding/json"
	"strconv"
	"testing"
)

func TestTaskHistoryCursorFiltersCurrentUserAndValidatesInput(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	wanted := map[string]bool{}
	for _, actor := range []string{f.readerUser.ID, f.adminUser.ID, f.readerUser.ID} {
		task, _, err := f.store.Accept(ctx, model.Task{ActorID: actor, RequestID: model.ID(), Resource: instanceRef(f.instances[0]), Action: "fixture", Payload: []byte(`{}`)})
		if err != nil {
			t.Fatal(err)
		}
		if actor == f.readerUser.ID {
			wanted[task.ID] = true
		}
	}
	before := int64(0)
	seen := map[string]bool{}
	for {
		result := f.reader.request("GET", "/tasks?before="+strconv.FormatInt(before, 10)+"&limit=1", nil, "", 200)
		var items []model.Task
		json.Unmarshal(result["items"], &items)
		for _, task := range items {
			if !wanted[task.ID] || seen[task.ID] {
				t.Fatal("unauthorized or duplicate task")
			}
			seen[task.ID] = true
		}
		if err := json.Unmarshal(result["nextBefore"], &before); err != nil {
			t.Fatal(err)
		}
		if before < 0 {
			break
		}
	}
	if len(seen) != len(wanted) {
		t.Fatal("missing authorized history")
	}
	for _, query := range []string{"before=-1", "before=x", "before=0&offset=0", "before=0&limit=1001"} {
		f.reader.request("GET", "/tasks?"+query, nil, "", 400)
	}
}

func TestTaskSummaryAndStateFilterIncludeOldAuthorizedPending(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	tx, err := f.store.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	insert := func(actor string, state model.TaskState, payload string) string {
		task := model.Task{ID: model.ID(), ActorID: actor, RequestID: model.ID(), Resource: instanceRef(f.instances[0]), State: state, Revision: 1, Action: "fixture", Payload: json.RawMessage(payload)}
		data, err := json.Marshal(task)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO tasks(id,actor_id,request_id,resource_key,digest,state,revision,document) VALUES(?,?,?,?,?,?,?,?)", task.ID, actor, task.RequestID, task.Resource.Key(), "fixture", state, 1, data)
		if err != nil {
			t.Fatal(err)
		}
		return task.ID
	}
	old := insert(f.readerUser.ID, model.WaitingClient, `{}`)
	insert(f.adminUser.ID, model.WaitingClient, `{}`)
	insert(f.readerUser.ID, model.WaitingClient, `{"transferParentId":"internal"}`)
	for range 1001 {
		insert(f.readerUser.ID, model.Succeeded, `{}`)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	result := f.reader.request("GET", "/tasks/summary", nil, "", 200)
	if string(result["active"]) != "1" {
		t.Fatalf("incorrect authorized active total: %s", result["active"])
	}
	result = f.admin.request("GET", "/tasks/summary", nil, "", 200)
	if string(result["active"]) != "2" {
		t.Fatalf("admin active roots incomplete: active=%s states=%s", result["active"], result["states"])
	}
	result = f.reader.request("GET", "/tasks?before=0&state=WAITING_CLIENT", nil, "", 200)
	var tasks []model.Task
	if err := json.Unmarshal(result["items"], &tasks); err != nil || len(tasks) != 1 || tasks[0].ID != old {
		t.Fatal("state filtering lost old task or disclosed hidden task")
	}
	f.reader.request("GET", "/tasks?before=0&state=INVALID", nil, "", 400)
}
