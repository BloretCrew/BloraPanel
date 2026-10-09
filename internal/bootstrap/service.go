package bootstrap

// The stable launcher owns only the management child. Instance processes, log
// helpers and Windows Job keepers retain their independent recorded lifetimes.
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const ManagedFlag = "--managed-service"

type ServiceChild struct {
	InstallRoot, UpdateRoot, Token string
}

// ParseServiceChild consumes the private launcher prefix, retaining the user's
// original arguments and the installation root used for relative configuration.
func ParseServiceChild(args []string) (*ServiceChild, []string, error) {
	if len(args) == 0 || args[0] != ManagedFlag {
		return nil, args, nil
	}
	if len(args) < 4 || !filepath.IsAbs(args[1]) || !filepath.IsAbs(args[2]) || len(args[3]) != 32 {
		return nil, nil, errors.New("invalid managed service arguments")
	}
	if _, err := hex.DecodeString(args[3]); err != nil {
		return nil, nil, err
	}
	return &ServiceChild{args[1], args[2], args[3]}, args[4:], nil
}

type Activation struct {
	Revision  string `json:"revision"`
	Binary    string `json:"binary"`
	Directory string `json:"directory"`
}

func serviceWrite(root, name string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(root, ".activation-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return replaceServiceFile(f.Name(), filepath.Join(root, name))
}

func readActivation(root, name string) (Activation, error) {
	var a Activation
	p := filepath.Join(root, name)
	st, err := os.Lstat(p)
	if err != nil {
		return a, err
	}
	if !st.Mode().IsRegular() || st.Size() > 8192 {
		return a, errors.New("invalid activation record")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return a, err
	}
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.DisallowUnknownFields()
	if err = d.Decode(&a); err != nil {
		return a, err
	}
	if d.Decode(new(any)) != io.EOF {
		return a, errors.New("invalid activation record")
	}
	if len(a.Revision) != 40 && len(a.Revision) != 64 {
		return a, errors.New("invalid activation revision")
	}
	if _, err = hex.DecodeString(a.Revision); err != nil {
		return a, err
	}
	// Only a verified component within its immutable revision directory can
	// be activated; never execute an arbitrary persisted filesystem path.
	slot := filepath.Base(a.Directory)
	validSlot := slot == a.Revision
	if suffix, ok := strings.CutPrefix(slot, a.Revision+"-stage-"); ok && suffix != "" && len(suffix) <= 32 {
		validSlot = true
		for _, c := range suffix {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
				validSlot = false
			}
		}
	}
	expected := filepath.Join(root, "revisions", slot)
	if !validSlot || filepath.Clean(a.Directory) != expected || filepath.Dir(a.Binary) != expected || !filepath.IsAbs(a.Binary) {
		return a, errors.New("activation escapes revision directory")
	}
	resolved, err := filepath.EvalSymlinks(expected)
	if err != nil || resolved != expected {
		return a, errors.New("activation directory must not contain symlinks")
	}
	st, err = os.Lstat(a.Binary)
	if err != nil || !st.Mode().IsRegular() {
		return a, errors.New("activation binary is missing or not regular")
	}
	return a, nil
}

func RequestActivation(root string, a Activation) error {
	if err := serviceWrite(root, "pending.json", a); err != nil {
		return err
	}
	_, err := readActivation(root, "pending.json")
	if err != nil {
		_ = os.Remove(filepath.Join(root, "pending.json"))
	}
	return err
}

func (c *ServiceChild) Ready() error {
	return serviceWrite(c.UpdateRoot, "ready-"+c.Token+".json", true)
}

// ReadyAndWait prevents a candidate from reporting a successful update before
// its launcher durably commits the active version. Initial startup has no
// pending activation and does not wait for an update acknowledgement.
func (c *ServiceChild) ReadyAndWait(ctx context.Context) error {
	pending, err := readActivation(c.UpdateRoot, "pending.json")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := c.Ready(); err != nil {
		return err
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return err
	}
	if executable != pending.Binary {
		return nil
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		active, e := readActivation(c.UpdateRoot, "active.json")
		if e == nil && active == pending {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return errors.New("launcher did not commit the active version")
		case <-ticker.C:
		}
	}
}

// WatchStop uses private, per-launch control files to support graceful Windows
// shutdown as well as Unix signals, without invoking a shell or killing runs.
func (c *ServiceChild) WatchStop(ctx context.Context, stop context.CancelFunc) {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if _, err := os.Stat(filepath.Join(c.UpdateRoot, "stop-"+c.Token+".json")); err == nil {
				stop()
				return
			}
		}
	}
}

type ServiceOptions struct {
	Executable, InstallRoot, UpdateRoot string
	Args                                []string
	ReadyTimeout                        time.Duration
}

type serviceProcess struct {
	command *exec.Cmd
	done    chan error
	token   string
}

func startService(o ServiceOptions, binary string) (*serviceProcess, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, err
	}
	p := &serviceProcess{token: hex.EncodeToString(token[:]), done: make(chan error, 1)}
	args := append([]string{ManagedFlag, o.InstallRoot, o.UpdateRoot, p.token}, o.Args...)
	p.command = exec.Command(binary, args...)
	p.command.Dir = o.InstallRoot
	p.command.Stdin, p.command.Stdout, p.command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := p.command.Start(); err != nil {
		return nil, err
	}
	go func() { p.done <- p.command.Wait() }()
	return p, nil
}

func cleanControl(o ServiceOptions, p *serviceProcess) {
	_ = os.Remove(filepath.Join(o.UpdateRoot, "ready-"+p.token+".json"))
	_ = os.Remove(filepath.Join(o.UpdateRoot, "stop-"+p.token+".json"))
}

func stopService(o ServiceOptions, p *serviceProcess) {
	_ = serviceWrite(o.UpdateRoot, "stop-"+p.token+".json", true)
	select {
	case <-p.done:
	case <-time.After(30 * time.Second):
		_ = p.command.Process.Kill() // Exact child only, never a process group.
		<-p.done
	}
	cleanControl(o, p)
}

// Supervise changes the management executable without overwriting its running
// image (including Windows). The installation directory and configuration stay
// stable; there is never more than one management child writing the state.
func Supervise(ctx context.Context, o ServiceOptions) error {
	if !filepath.IsAbs(o.Executable) || !filepath.IsAbs(o.InstallRoot) || !filepath.IsAbs(o.UpdateRoot) {
		return errors.New("service paths must be absolute")
	}
	if err := os.MkdirAll(o.UpdateRoot, 0700); err != nil {
		return err
	}
	unlock, err := lockService(o.UpdateRoot)
	if err != nil {
		return err
	}
	defer unlock()
	if o.ReadyTimeout == 0 {
		o.ReadyTimeout = 90 * time.Second
	}
	active, err := readActivation(o.UpdateRoot, "active.json")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	binary := o.Executable
	if err == nil {
		binary = active.Binary
	}
	var candidate *Activation
	if next, e := readActivation(o.UpdateRoot, "pending.json"); e == nil {
		candidate = &next
	} else if !errors.Is(e, os.ErrNotExist) {
		return e
	}
	for {
		if ctx.Err() != nil {
			return nil
		}
		selected := binary
		if candidate != nil {
			selected = candidate.Binary
		}
		p, err := startService(o, selected)
		if err != nil {
			if candidate == nil {
				return err
			}
			_ = serviceWrite(o.UpdateRoot, "activation-error.json", map[string]string{"revision": candidate.Revision, "error": "candidate could not start; previous version retained"})
			_ = os.Remove(filepath.Join(o.UpdateRoot, "pending.json"))
			candidate = nil
			continue
		}
		deadline := time.NewTimer(o.ReadyTimeout)
		var deadlineC <-chan time.Time = deadline.C
		if candidate == nil {
			deadline.Stop()
			deadlineC = nil
		} // An offline initial node keeps reconnecting.
		ticker := time.NewTicker(50 * time.Millisecond)
		ready := false
		ended := false
		var endErr error
		for !ready && !ended {
			select {
			case <-ctx.Done():
				ticker.Stop()
				deadline.Stop()
				stopService(o, p)
				return nil
			case endErr = <-p.done:
				ended = true
			case <-deadlineC:
				stopService(o, p)
				ended = true
				endErr = errors.New("management readiness timed out")
			case <-ticker.C:
				_, e := os.Stat(filepath.Join(o.UpdateRoot, "ready-"+p.token+".json"))
				ready = e == nil
			}
		}
		ticker.Stop()
		deadline.Stop()
		if !ready {
			cleanControl(o, p)
			if candidate == nil {
				return fmt.Errorf("management service did not become ready: %w", endErr)
			}
			_ = serviceWrite(o.UpdateRoot, "activation-error.json", map[string]string{"revision": candidate.Revision, "error": "candidate readiness failed; previous version retained"})
			_ = os.Remove(filepath.Join(o.UpdateRoot, "pending.json"))
			candidate = nil
			continue
		}
		if candidate != nil {
			if err := serviceWrite(o.UpdateRoot, "active.json", *candidate); err != nil {
				stopService(o, p)
				return err
			}
			binary = candidate.Binary
			active = *candidate
			candidate = nil
			_ = os.Remove(filepath.Join(o.UpdateRoot, "pending.json"))
			_ = os.Remove(filepath.Join(o.UpdateRoot, "activation-error.json"))
		}
		select {
		case <-ctx.Done():
			stopService(o, p)
			return nil
		case endErr = <-p.done:
			cleanControl(o, p)
		}
		next, e := readActivation(o.UpdateRoot, "pending.json")
		if e == nil && endErr == nil {
			candidate = &next
			continue
		}
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return e
		}
		return endErr
	}
}
