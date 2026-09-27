package monitor

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

type ProcessQuery struct {
	Limit    int    `json:"limit"`
	AfterPID int    `json:"afterPid"`
	Search   string `json:"search"`
}
type ProcessPage struct {
	Items        []SystemProcess `json:"items"`
	NextAfterPID int             `json:"nextAfterPid,omitempty"`
}

func ListProcesses(limit int) ([]SystemProcess, error) {
	p, err := ListProcessPage(ProcessQuery{Limit: limit})
	return p.Items, err
}

// Each request observes a live list. A PID cursor avoids offset shifts when
// earlier processes exit; newly created lower PIDs appear after refreshing.
func ListProcessPage(q ProcessQuery) (ProcessPage, error) {
	if q.Limit == 0 {
		q.Limit = 50
	}
	if q.Limit < 1 || q.Limit > 100 || q.AfterPID < 0 || len(q.Search) > 256 {
		return ProcessPage{}, errors.New("invalid process page query")
	}
	out := ProcessPage{Items: make([]SystemProcess, 0, q.Limit+1)}
	search := strings.ToLower(q.Search)
	err := visitProcesses(func(p SystemProcess) {
		if p.PID <= q.AfterPID || (search != "" && !strings.Contains(strings.ToLower(p.Command), search) && !strings.Contains(strconv.Itoa(p.PID), search)) {
			return
		}
		i := sort.Search(len(out.Items), func(i int) bool { return out.Items[i].PID >= p.PID })
		if i > q.Limit {
			return
		}
		if len(p.Command) > 4096 {
			p.Command = strings.ToValidUTF8(p.Command[:4096], "�")
		}
		out.Items = append(out.Items, SystemProcess{})
		copy(out.Items[i+1:], out.Items[i:])
		out.Items[i] = p
		if len(out.Items) > q.Limit+1 {
			out.Items = out.Items[:q.Limit+1]
		}
	})
	if err != nil {
		return ProcessPage{}, err
	}
	if len(out.Items) > q.Limit {
		out.Items = out.Items[:q.Limit]
		out.NextAfterPID = out.Items[len(out.Items)-1].PID
	}
	return out, nil
}
