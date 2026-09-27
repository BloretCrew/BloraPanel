package containers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (x *execution) mutate(ctx context.Context, phase, method, path string, body, out any) error {
	return x.mutateChecked(ctx, phase, method, path, body, out, nil)
}
func (x *execution) mutateChecked(ctx context.Context, phase, method, path string, body, out any, check func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := x.emit(phase, "Apply operation to the fixed Engine resource"); err != nil {
		return err
	}
	if check != nil {
		if err := check(); err != nil {
			return err
		}
	}
	x.entry.Result.Unknown = true
	if err := x.save(); err != nil {
		return err
	}
	err := x.manager.engine.api(ctx, method, path, body, out)
	if err != nil {
		var response *engineError
		if errors.As(err, &response) && response.Status >= 400 && response.Status < 500 {
			x.entry.Result.Unknown = false
		}
		return err
	}
	x.entry.Result.Changed = true
	return x.save()
}
func (x *execution) confirmed(objects ...Object) error {
	for _, o := range objects {
		x.trackTarget(o.Target)
		found := false
		for i, prior := range x.entry.Result.Facts {
			if resourceKey(prior.Target) == resourceKey(o.Target) {
				x.entry.Result.Facts[i] = o
				found = true
				break
			}
		}
		if !found {
			x.entry.Result.Facts = append(x.entry.Result.Facts, o)
		}
	}
	x.entry.Result.Unknown = false
	return x.save()
}
func resourceKey(t Target) string {
	if t.Kind == "volume" {
		return t.Kind + ":" + t.Name
	}
	return t.Kind + ":" + t.ID
}
func (x *execution) trackTarget(t Target) {
	for i, prior := range x.entry.Result.Targets {
		if resourceKey(prior) == resourceKey(t) {
			x.entry.Result.Targets[i] = t
			return
		}
	}
	x.entry.Result.Targets = append(x.entry.Result.Targets, t)
}
func notFound(err error) bool { var e *engineError; return errors.As(err, &e) && e.Status == 404 }
func createDocument(raw json.RawMessage, labels map[string]string) (map[string]any, error) {
	var doc map[string]any
	if len(raw) == 0 {
		doc = map[string]any{}
	} else if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.New("create document must be an object")
	}
	existing, _ := doc["Labels"].(map[string]any)
	if existing == nil {
		existing = map[string]any{}
	}
	for k, v := range labels {
		existing[k] = v
	}
	doc["Labels"] = existing
	return doc, nil
}
func (x *execution) engineOperation(ctx context.Context) error {
	op := x.entry.Operation
	e := x.manager.engine
	if op.StopSeconds <= 0 {
		op.StopSeconds = 30
	}
	if op.StopSeconds > 300 {
		return errors.New("stop timeout exceeds 300 seconds")
	}
	switch op.Action {
	case "container.create":
		doc, err := createDocument(op.Config, x.manager.labels(op.TaskID))
		if err != nil {
			return err
		}
		if _, ok := doc["Image"].(string); !ok {
			return errors.New("container Image is required")
		}
		host, _ := doc["HostConfig"].(map[string]any)
		if host == nil {
			host = map[string]any{}
		}
		if _, ok := host["LogConfig"]; !ok {
			host["LogConfig"] = map[string]any{"Type": "local", "Config": map[string]string{"max-size": "10m", "max-file": "3"}}
		}
		doc["HostConfig"] = host
		path := "/containers/create"
		if op.Name != "" {
			if !volumeName.MatchString(op.Name) {
				return ErrIdentity
			}
			path += "?name=" + url.QueryEscape(op.Name)
		}
		var response struct {
			ID string `json:"Id"`
		}
		if err = x.mutate(ctx, "create", "POST", path, doc, &response); err != nil {
			return err
		}
		o, err := e.inspect(ctx, Target{Kind: "container", ID: response.ID}, false)
		if err != nil {
			return err
		}
		if o.Labels[labelPrefix+"operation"] != op.TaskID || o.Labels[labelPrefix+"owner"] != x.manager.token {
			return ErrIdentity
		}
		return x.confirmed(o)
	case "container.start", "container.stop", "container.restart", "container.delete":
		if err := validateTarget(op.Target, "container"); err != nil {
			return err
		}
		o, err := e.inspect(ctx, op.Target, true)
		if err != nil {
			return err
		}
		var before containerInspect
		if err = json.Unmarshal(o.Details, &before); err != nil {
			return err
		}
		base := "/containers/" + op.Target.ID
		switch op.Action {
		case "container.start":
			if !before.State.Running {
				if err = x.mutate(ctx, "start", "POST", base+"/start", nil, nil); err != nil {
					return err
				}
			}
			o, err = e.inspect(ctx, op.Target, false)
			if err != nil {
				return err
			}
			if o.State != "running" {
				return fmt.Errorf("container did not remain running: %w", ErrUnknown)
			}
			return x.confirmed(o)
		case "container.stop":
			if before.State.Running {
				if err = x.mutate(ctx, "stop", "POST", base+"/stop?t="+strconv.Itoa(op.StopSeconds), nil, nil); err != nil {
					return err
				}
			}
			o, err = e.inspect(ctx, op.Target, false)
			if err != nil {
				return err
			}
			if o.State == "running" || o.State == "restarting" || o.State == "paused" {
				return ErrUnknown
			}
			return x.confirmed(o)
		case "container.restart":
			if before.State.Running {
				if err = x.mutate(ctx, "stop", "POST", base+"/stop?t="+strconv.Itoa(op.StopSeconds), nil, nil); err != nil {
					return err
				}
				stopped, err := e.inspect(ctx, op.Target, false)
				if err != nil {
					return err
				}
				if stopped.State == "running" || stopped.State == "restarting" {
					return ErrUnknown
				}
				if err = x.confirmed(stopped); err != nil {
					return err
				}
			}
			if err = ctx.Err(); err != nil {
				return err
			}
			if err = x.mutate(ctx, "start", "POST", base+"/start", nil, nil); err != nil {
				return err
			}
			o, err = e.inspect(ctx, op.Target, false)
			if err != nil {
				return err
			}
			if o.State != "running" {
				return ErrUnknown
			}
			return x.confirmed(o)
		case "container.delete":
			if err = x.mutate(ctx, "delete", "DELETE", base+"?v=false&force="+strconv.FormatBool(op.Force), nil, nil); err != nil {
				return err
			}
			_, err = e.inspect(ctx, op.Target, false)
			if !notFound(err) {
				if err != nil {
					return err
				}
				return ErrUnknown
			}
			x.entry.Result.Targets = append(x.entry.Result.Targets, op.Target)
			if err = x.confirmed(); err != nil {
				return err
			}
			for _, target := range op.DeleteVolumes {
				if err = x.deleteVolume(ctx, target); err != nil {
					return err
				}
			}
			return nil
		}
	case "image.pull":
		return x.pullImage(ctx, op.Image)
	case "image.delete":
		if err := validateTarget(op.Target, "image"); err != nil {
			return err
		}
		if _, err := e.inspect(ctx, op.Target, false); err != nil {
			return err
		}
		var deleted []map[string]string
		if err := x.mutate(ctx, "delete", "DELETE", "/images/"+op.Target.ID+"?force="+strconv.FormatBool(op.Force)+"&noprune=true", nil, &deleted); err != nil {
			return err
		}
		if _, err := e.inspect(ctx, op.Target, false); !notFound(err) {
			if err != nil {
				return err
			}
			return ErrUnknown
		}
		x.entry.Result.Targets = append(x.entry.Result.Targets, op.Target)
		return x.confirmed()
	case "volume.create":
		if !volumeName.MatchString(op.Name) {
			return ErrIdentity
		}
		if _, err := e.inspect(ctx, Target{Kind: "volume", Name: op.Name}, false); !notFound(err) {
			if err != nil {
				return err
			}
			return ErrConflict
		}
		doc, err := createDocument(op.Config, x.manager.labels(op.TaskID))
		if err != nil {
			return err
		}
		doc["Name"] = op.Name
		var response volumeInspect
		if err = x.mutate(ctx, "create", "POST", "/volumes/create", doc, &response); err != nil {
			return err
		}
		o, err := e.inspect(ctx, Target{Kind: "volume", Name: op.Name}, false)
		if err != nil {
			return err
		}
		if o.Labels[labelPrefix+"operation"] != op.TaskID {
			return ErrIdentity
		}
		return x.confirmed(o)
	case "volume.delete":
		return x.deleteVolume(ctx, op.Target)
	case "network.create":
		if !volumeName.MatchString(op.Name) {
			return ErrIdentity
		}
		doc, err := createDocument(op.Config, x.manager.labels(op.TaskID))
		if err != nil {
			return err
		}
		doc["Name"] = op.Name
		doc["CheckDuplicate"] = true
		var response struct {
			ID string `json:"Id"`
		}
		if err = x.mutate(ctx, "create", "POST", "/networks/create", doc, &response); err != nil {
			return err
		}
		o, err := e.inspect(ctx, Target{Kind: "network", ID: response.ID}, false)
		if err != nil {
			return err
		}
		if o.Labels[labelPrefix+"operation"] != op.TaskID {
			return ErrIdentity
		}
		return x.confirmed(o)
	case "network.delete":
		if err := validateTarget(op.Target, "network"); err != nil {
			return err
		}
		if _, err := e.inspect(ctx, op.Target, false); err != nil {
			return err
		}
		if err := x.mutate(ctx, "delete", "DELETE", "/networks/"+op.Target.ID, nil, nil); err != nil {
			return err
		}
		if _, err := e.inspect(ctx, op.Target, false); !notFound(err) {
			if err != nil {
				return err
			}
			return ErrUnknown
		}
		x.entry.Result.Targets = append(x.entry.Result.Targets, op.Target)
		return x.confirmed()
	default:
		return errors.New("unsupported container operation")
	}
	return errors.New("unsupported container operation")
}
func (x *execution) deleteVolume(ctx context.Context, t Target) error {
	if err := validateTarget(t, "volume"); err != nil {
		return err
	}
	if t.CreatedAt == "" || !fullID.MatchString(t.Fingerprint) {
		return ErrIdentity
	}
	if _, err := x.manager.engine.inspect(ctx, t, false); err != nil {
		return err
	}
	// Docker provides no atomic If-Match for name-addressed volumes. Recheck
	// after the persistent pre-effect callback; never use force or global prune.
	if err := x.mutateChecked(ctx, "delete-volume", "DELETE", "/volumes/"+url.PathEscape(t.Name)+"?force=false", nil, nil, func() error {
		_, err := x.manager.engine.inspect(ctx, t, false)
		return err
	}); err != nil {
		return err
	}
	if _, err := x.manager.engine.inspect(ctx, t, false); !notFound(err) {
		if err != nil {
			return err
		}
		return ErrUnknown
	}
	x.entry.Result.Targets = append(x.entry.Result.Targets, t)
	return x.confirmed()
}
func (x *execution) pullImage(ctx context.Context, reference string) error {
	if reference == "" || len(reference) > 512 || strings.ContainsAny(reference, " \t\r\n\x00") || strings.Contains(reference, "://") {
		return errors.New("invalid image reference")
	}
	if err := x.manager.engine.check(ctx); err != nil {
		return err
	}
	if err := x.emit("pull", "Pull requested image; layer progress is not an invented total percentage"); err != nil {
		return err
	}
	x.entry.Result.Unknown = true
	if err := x.save(); err != nil {
		return err
	}
	res, err := x.manager.engine.response(ctx, "POST", "/v"+apiVersion+"/images/create?fromImage="+url.QueryEscape(reference), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	limited := &io.LimitedReader{R: res.Body, N: (32 << 20) + 1}
	decoder := json.NewDecoder(limited)
	last := time.Time{}
	var pending *Progress
	for {
		var event struct {
			Status         string
			ID             string `json:"id"`
			Error          string
			ErrorDetail    struct{ Message string }
			ProgressDetail struct{ Current, Total int64 }
		}
		err = decoder.Decode(&event)
		if limited.N == 0 {
			return errors.New("image pull progress exceeds 32 MiB; remote result remains unknown")
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if event.Error != "" || event.ErrorDetail.Message != "" {
			x.entry.Result.Unknown = false
			return fmt.Errorf("image pull failed: %s %s", event.Error, event.ErrorDetail.Message)
		}
		x.entry.Result.Changed = true
		pending = &Progress{Phase: "pull", Message: event.Status, Layer: event.ID, Current: event.ProgressDetail.Current, Total: event.ProgressDetail.Total}
		if time.Since(last) > 200*time.Millisecond {
			if err = x.emitProgress(*pending); err != nil {
				return err
			}
			last = time.Now()
			pending = nil
		}
	}
	if pending != nil {
		if err := x.emitProgress(*pending); err != nil {
			return err
		}
	}
	// Inspect by reference only to resolve the final immutable image identity.
	var inspect struct {
		ID string `json:"Id"`
	}
	if err = x.manager.engine.api(ctx, "GET", "/images/"+url.PathEscape(reference)+"/json", nil, &inspect); err != nil {
		return err
	}
	o, err := x.manager.engine.inspect(ctx, Target{Kind: "image", ID: inspect.ID}, false)
	if err != nil {
		return err
	}
	return x.confirmed(o)
}
