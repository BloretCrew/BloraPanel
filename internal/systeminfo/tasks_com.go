package systeminfo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// All executable script text is fixed. Task names are passed as data through
// the child environment, never interpolated into PowerShell source.
const taskCOMPrelude = `$ErrorActionPreference='Stop'
[Console]::OutputEncoding=New-Object System.Text.UTF8Encoding($false)
$service=New-Object -ComObject 'Schedule.Service'
$service.Connect()
`
const taskCOMList = taskCOMPrelude + `
$rows=New-Object 'System.Collections.Generic.SortedList[string,object]' ([StringComparer]::Ordinal)
$folders=New-Object 'System.Collections.Generic.Queue[object]'
$folders.Enqueue($service.GetFolder('\'))
$limit=[int]$env:BLORA_TASK_LIMIT
$visited=0
while($folders.Count -gt 0){
 $folder=$folders.Dequeue()
 $visited++
 if($visited -gt 10000){throw 'Task folder enumeration limit exceeded'}
 foreach($task in $folder.GetTasks(1)){
  if([StringComparer]::Ordinal.Compare([string]$task.Path,$env:BLORA_TASK_AFTER) -le 0){continue}
  $states=@('unknown','disabled','queued','ready','running')
  $state='unknown'
  if([int]$task.State -ge 0 -and [int]$task.State -lt $states.Length){$state=$states[[int]$task.State]}
  $next=''
  if($task.NextRunTime.Year -gt 1900){$next=$task.NextRunTime.ToUniversalTime().ToString('o')}
  $rows[[string]$task.Path]=[pscustomobject]@{name=$task.Path;state=$state;schedule=$next}
  if($rows.Count -gt $limit){$rows.RemoveAt($rows.Count-1)}
 }
 foreach($child in $folder.GetFolders(0)){if($folders.Count -ge 10000){throw 'Task folder queue limit exceeded'};$folders.Enqueue($child)}
}
ConvertTo-Json -InputObject @($rows.Values) -Compress
`
const taskCOMAction = taskCOMPrelude + `
$path=$env:BLORA_TASK_PATH
$index=$path.LastIndexOf('\')
$folderPath=$path.Substring(0,$index)
if($folderPath -eq ''){$folderPath='\'}
$name=$path.Substring($index+1)
$folder=$service.GetFolder($folderPath)
$task=$folder.GetTask($name)
$enabled=$env:BLORA_TASK_ENABLED -eq 'true'
$task.Enabled=$enabled
if([bool]($folder.GetTask($name).Enabled) -ne $enabled){throw 'Task enablement not confirmed'}
`

type taskOutput struct{ bytes.Buffer }

func (b *taskOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 1<<20 {
		return 0, errors.New("task output exceeded limit")
	}
	return b.Buffer.Write(p)
}
func runTaskCOM(ctx context.Context, script string, env ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Env = append(os.Environ(), env...)
	var out, diagnostic taskOutput
	cmd.Stdout = &out
	cmd.Stderr = &diagnostic
	if err := cmd.Run(); err != nil {
		return nil, errors.New("Task Scheduler COM operation failed: " + err.Error() + " " + strings.TrimSpace(diagnostic.String()))
	}
	return out.Bytes(), nil
}
func listWindowsTasks(ctx context.Context, limit int, after string) ([]ScheduledTask, error) {
	b, err := runTaskCOM(ctx, taskCOMList, "BLORA_TASK_LIMIT="+strconv.Itoa(limit), "BLORA_TASK_AFTER="+after)
	if err != nil {
		return nil, err
	}
	var items []ScheduledTask
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, err
	}
	if len(items) > limit {
		return nil, errors.New("task list exceeded requested limit")
	}
	for _, item := range items {
		if !strings.HasPrefix(item.Name, `\`) || !ValidTaskName(item.Name) {
			return nil, errors.New("invalid task identity returned")
		}
	}
	return items, nil
}
func windowsTaskAction(ctx context.Context, name, action string) error {
	if !strings.HasPrefix(name, `\`) || !ValidTaskName(name) {
		return errors.New("full Windows task path required")
	}
	_, err := runTaskCOM(ctx, taskCOMAction, "BLORA_TASK_PATH="+name, "BLORA_TASK_ENABLED="+strconv.FormatBool(action == "enable"))
	return err
}
