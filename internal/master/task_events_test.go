package master

import (
	"blora.dev/panel/internal/model"
	"blora.dev/panel/internal/storage"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func TestTaskStagesPaginationProjectionAndCurrentAuthorization(t *testing.T) {
	f := newFileFixture(t)
	ctx := context.Background()
	task, _, err := f.store.Accept(ctx, model.Task{ActorID: f.readerUser.ID, RequestID: model.ID(), Resource: instanceRef(f.instances[0]), Action: "fixture", Payload: json.RawMessage(`{"secret":"not-stage-content"}`)})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 55; i++ {
		task, err = f.store.UpdateTask(ctx, task.ID, task.Revision, model.WaitingClient, "waiting-"+strconv.Itoa(i), nil, "")
		if err != nil {
			t.Fatal(err)
		}
	}
	task, err = f.store.UpdateTask(ctx, task.ID, task.Revision, model.Failed, strings.Repeat("p", 513), nil, strings.Repeat("x", 4097))
	if err != nil {
		t.Fatal(err)
	}
	path := "/tasks/" + task.ID + "/events"
	before := int64(0)
	seen := map[int64]bool{}
	for {
		result := f.reader.request("GET", path+"?before="+strconv.FormatInt(before, 10)+"&limit=20", nil, "", 200)
		if strings.Contains(string(result["items"]), "not-stage-content") {
			t.Fatal("payload leaked into stage diagnostics")
		}
		var items []storage.TaskStageEvent
		if err := json.Unmarshal(result["items"], &items); err != nil {
			t.Fatal(err)
		}
		if len(items) > 20 {
			t.Fatal("unbounded stage page")
		}
		for _, event := range items {
			if seen[event.Revision] {
				t.Fatal("duplicate stage event")
			}
			seen[event.Revision] = true
			if event.Revision == task.Revision && (!event.Truncated || len(event.Phase) != 512 || len(event.Error) != 4096) {
				t.Fatal("diagnostic bounds not marked")
			}
		}
		if err := json.Unmarshal(result["nextBefore"], &before); err != nil {
			t.Fatal(err)
		}
		if before < 0 {
			break
		}
	}
	if int64(len(seen)) != task.Revision {
		t.Fatal("stage history incomplete")
	}
	f.reader.request("GET", path+"?limit=101", nil, "", 400)
	f.reader.request("GET", path+"?before=-1", nil, "", 400)
	f.admin.request("DELETE", "/grants", model.Grant{UserID: f.readerUser.ID, Resource: instanceRef(f.instances[0]), Action: "instance.read"}, model.ID(), 200)
	f.reader.request("GET", path, nil, "", 404)
}
