#!/usr/bin/env python3
"""Supervise isolated Linux wall-clock monitoring/minute-backup endurance.

start --seconds 86400 starts a detached worker with immutable service binaries.
status/stop/resume take its printed private run directory. A resumed worker
starts a NEW continuous coverage interval; a gap never counts as passing time.
No system clock, global service, network rule, or production node is changed.
"""
import argparse
import collections
import datetime as dt
import fcntl
import hashlib
import http.cookiejar
import json
import logging
import logging.handlers
import os
from pathlib import Path
import secrets
import shutil
import signal
import socket
import sqlite3
import ssl
import subprocess
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.request
import uuid

REPO = Path(__file__).resolve().parents[1]
MIB = 1024 * 1024
TERMINAL = {'SUCCEEDED', 'FAILED', 'INTERRUPTED', 'CANCELLED'}


def utc():
    return dt.datetime.now(dt.timezone.utc).isoformat()


def atomic(path, value):
    temporary = path.with_suffix('.tmp')
    with temporary.open('w', encoding='utf-8') as stream:
        os.chmod(temporary, 0o600)
        json.dump(value, stream, ensure_ascii=False, indent=2)
        stream.flush()
        os.fsync(stream.fileno())
    os.replace(temporary, path)


def read(path):
    return json.loads(path.read_text(encoding='utf-8'))


def identity(pid):
    root = Path('/proc') / str(pid)
    fields = (root / 'stat').read_text().rsplit(') ', 1)[1].split()
    return {'pid': pid, 'startTicks': fields[19], 'exe': str((root / 'exe').resolve(strict=True)),
            'bootId': Path('/proc/sys/kernel/random/boot_id').read_text().strip()}


def alive(expected):
    try:
        return identity(expected['pid']) == expected
    except (FileNotFoundError, ProcessLookupError, ValueError):
        return False


def owned_signal(expected, sig):
    # pidfd binds a signal to this incarnation even if the PID is reused after
    # validation; the executable, boot and start ticks must also match.
    try:
        descriptor = os.pidfd_open(expected['pid'])
    except ProcessLookupError:
        return False
    try:
        if not alive(expected):
            return False
        signal.pidfd_send_signal(descriptor, sig)
        return True
    finally:
        os.close(descriptor)


def logger(root, name, size=2 * MIB, backups=4):
    result = logging.getLogger(str(root / name))
    if result.handlers:
        return result
    result.setLevel(logging.INFO)
    result.propagate = False
    handler = logging.handlers.RotatingFileHandler(root / name, maxBytes=size, backupCount=backups)
    handler.setFormatter(logging.Formatter('%(message)s'))
    result.addHandler(handler)
    return result


class Interrupted(Exception):
    pass


class Runner:
    def __init__(self, root):
        self.root = root
        self.config = read(root / 'run.json')
        self.state = read(root / 'checkpoint.json') if (root / 'checkpoint.json').exists() else {}
        self.children = {}
        self.log = logger(root, 'events.jsonl', MIB, 8)
        self.samples = logger(root, 'samples.jsonl', 4 * MIB, 8)
        self.stopping = False
        self.csrf = ''
        self.client = None
        self.histories = collections.deque(maxlen=120)
        self.started = time.monotonic()
        self.start_wall = time.time()
        self.sample_count = 0
        self.maxima = {}
        self.baseline = {}
        self.phases = []
        self.fault_windows = []
        self.scheduled = {}
        self.versions = self.state.get('versions', {})
        self.last_instance_at = 0
        self.last_sample_mono = None
        self.max_sample_gap = 0
        self.dates = set()
        signal.signal(signal.SIGTERM, self.interrupt)
        signal.signal(signal.SIGINT, self.interrupt)

    def interrupt(self, *_):
        self.stopping = True

    def event(self, stage, **values):
        self.log.info(json.dumps({'utc': utc(), 'stage': stage, **values}, ensure_ascii=False))

    def check(self):
        if self.stopping or (self.root / 'stop.request').exists():
            raise Interrupted('explicit stop requested')

    def wait(self, seconds):
        deadline = time.monotonic() + seconds
        while time.monotonic() < deadline:
            self.check()
            time.sleep(min(1, max(0, deadline - time.monotonic())))

    def until(self, predicate, timeout, description):
        deadline = time.monotonic() + timeout
        next_heartbeat = time.monotonic() + 5
        while time.monotonic() < deadline:
            self.check()
            if predicate():
                self.state.pop('activeWait', None)
                return
            if time.monotonic() >= next_heartbeat:
                self.save(activeWait=description)
                next_heartbeat = time.monotonic() + 5
            self.wait(.5)
        raise RuntimeError('deadline: ' + description)

    def save(self, status='RUNNING', **extra):
        self.state.update({'status': status, 'updatedAt': utc(), 'worker': identity(os.getpid()),
                           'origin': self.config['origin'], 'segmentStartedAt': self.segment_start,
                           'requiredSeconds': self.config['seconds'], 'elapsedSeconds': time.monotonic() - self.started,
                           'sampleCount': self.sample_count, 'successfulSlots': len(self.scheduled),
                           'datesUTC': sorted(self.dates), 'phases': self.phases, 'maxima': self.maxima,
                           'faultWindows': self.fault_windows, 'maxOrdinarySampleGapSeconds': self.max_sample_gap,
                           'baseline': self.baseline, 'versions': self.versions, **extra})
        atomic(self.root / 'checkpoint.json', self.state)

    def start_service(self, name, command):
        output = logger(self.root, name + '.log')
        child = subprocess.Popen(command, cwd=REPO, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                 start_new_session=True)
        entry = {'identity': identity(child.pid), 'command': command}
        self.state.setdefault('services', {})[name] = entry
        self.children[name] = child
        self.save()

        def drain():
            # Service output is diagnostic only. Read bounded chunks even if
            # an unexpected child writes a line without a newline.
            while chunk := child.stdout.read1(8192):
                output.info(chunk.decode('utf-8', errors='replace').rstrip())
            child.stdout.close()
        threading.Thread(target=drain, daemon=True).start()
        self.event('service_started', service=name, identity=entry['identity'])

    def stop_service(self, name):
        entry = self.state.get('services', {}).get(name)
        if not entry or not alive(entry['identity']):
            return
        owned_signal(entry['identity'], signal.SIGCONT)
        owned_signal(entry['identity'], signal.SIGINT)
        deadline = time.monotonic() + 15
        while alive(entry['identity']) and time.monotonic() < deadline:
            child = self.children.get(name)
            if child is not None and child.poll() is not None:
                break
            time.sleep(.2)
        if alive(entry['identity']):
            owned_signal(entry['identity'], signal.SIGKILL)
        child = self.children.get(name)
        if child is not None:
            child.wait(timeout=10)
        self.event('service_stopped', service=name, identity=entry['identity'])

    def master(self):
        self.start_service('master', [str(self.root / 'bin/blora-master'), '--state-dir', str(self.root / 'master'),
                           '--listen', self.config['origin'].removeprefix('https://'), '--origin', self.config['origin'],
                           '--static-dir', str(REPO / 'web/dist')])
        self.connect_client()
        def healthy():
            try:
                self.call('/healthz', api=False)
                return True
            except (OSError, urllib.error.URLError):
                return False
        self.until(healthy, 30, 'Master health')
        self.login()

    def connect_client(self):
        self.client = urllib.request.build_opener(
            urllib.request.HTTPSHandler(context=ssl.create_default_context(cafile=str(self.root / 'master/tls.crt'))),
            urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))

    def login(self):
        credentials = read(self.root / 'credentials.json')
        self.csrf = self.call('/login', credentials, method='POST', retry_login=False)['csrfToken']

    def call(self, path, value=None, method='GET', api=True, retry_login=True):
        request = urllib.request.Request(self.config['origin'] + ('/api/v1' if api else '') + path,
            data=None if value is None else json.dumps(value).encode(), method=method,
            headers={'Content-Type': 'application/json', 'Origin': self.config['origin'],
                     'X-CSRF-Token': self.csrf, 'Idempotency-Key': str(uuid.uuid4())})
        try:
            with self.client.open(request, timeout=10) as response:
                body = response.read(2 * MIB + 1)
                if len(body) > 2 * MIB:
                    raise RuntimeError('response exceeds evidence reader budget')
                return json.loads(body) if body else {}
        except urllib.error.HTTPError as error:
            if error.code == 401 and retry_login:
                error.close()
                self.login()
                # 401 rejected before admission, so this retry cannot replay
                # an accepted remote side effect.
                return self.call(path, value, method, api, False)
            error.close()
            raise RuntimeError(f'HTTP {error.code}: {method} {path}') from None

    def task(self, task_id):
        current = None
        def finished():
            nonlocal current
            current = self.call('/tasks/' + task_id)['task']
            return current['state'] in TERMINAL
        self.until(finished, 45, 'task ' + task_id)
        if current['state'] != 'SUCCEEDED':
            raise RuntimeError('task did not succeed: ' + task_id + ' ' + current['state'])
        return current

    def nodes_online(self):
        return any(node['nodeId'] == self.state['nodeId'] and node['state'] == 'ONLINE'
                   for node in self.call('/nodes')['items'])

    def setup(self):
        first = not (self.root / 'master/tls.crt').exists()
        if first:
            password = secrets.token_urlsafe(32)
            atomic(self.root / 'credentials.json', {'name': 'endurance-admin', 'password': password})
            password_file = self.root / 'initial-password'
            password_file.touch(mode=0o600)
            password_file.write_text(password)
            result = subprocess.run([str(self.root / 'bin/blora-master'), '--state-dir', str(self.root / 'master'),
                       '--init', '--admin-name', 'endurance-admin', '--password-file', str(password_file)],
                       stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=30)
            password_file.unlink()
            if result.returncode:
                raise RuntimeError('private Master initialization failed')
        else:
            # Resume is explicit. Stop only the recorded service incarnations,
            # then reconnect to the SAME private SQLite/Daemon identities.
            for name in ('daemon', 'master'):
                self.stop_service(name)
            self.event('resume_new_continuous_segment', previousUpdatedAt=self.state.get('updatedAt'))
        self.master()
        if not (self.root / 'daemon.json').exists():
            token = self.call('/nodes/enrollments', {'name': 'endurance-node'}, 'POST')['token']
            ticket = self.root / 'enrollment'
            ticket.touch(mode=0o600)
            ticket.write_text(token)
            atomic(self.root / 'daemon.json', {'stateDir': str(self.root / 'daemon'),
                'masterUrl': self.config['origin'], 'caFile': str(self.root / 'master/tls.crt'),
                'enrollmentFile': str(ticket), 'allowPGIDFallback': True})
        self.start_service('daemon', [str(self.root / 'bin/blora-daemon'), '--config', str(self.root / 'daemon.json')])
        def connected():
            nodes = self.call('/nodes')['items']
            if len(nodes) == 1 and nodes[0]['state'] == 'ONLINE':
                self.state['nodeId'] = nodes[0]['nodeId']
                return True
            return False
        self.until(connected, 35, 'private Daemon ONLINE')
        if 'instanceId' not in self.state:
            directory = self.root / 'resource'
            directory.mkdir(mode=0o700)
            command = "trap 'exit 0' TERM; payload='" + ('x' * 1024) + "'; while :; do printf 'ENDURANCE:%s\\n' \"$payload\"; sleep 1; done"
            instance = self.call('/instances', {'nodeId': self.state['nodeId'], 'name': 'endurance-owned',
                'config': {'mode': 'native', 'directory': str(directory), 'command': ['/bin/sh', '-c', command],
                           'stopSeconds': 3, 'killSeconds': 3, 'escalate': True}}, 'POST')['instance']
            self.state['instanceId'] = instance['instanceId']
            self.save()
        self.base = '/instances/' + self.state['instanceId']
        instance = self.call(self.base)['instance']
        if instance['state'] != 'RUNNING':
            self.task(self.call(self.base + '/actions', {'action': 'start'}, 'POST')['task']['taskId'])
            self.until(lambda: self.call(self.base)['instance']['state'] == 'RUNNING', 30, 'owned instance RUNNING')
        self.state['runId'] = self.call(self.base)['instance']['runId']
        self.source = 'endurance-source.txt'
        if 'scheduleId' in self.state:
            # If supervision died between a completed file write and its local
            # checkpoint, recover the server's CURRENT value by reading it;
            # never replay an uncertain old write.
            content = self.call(self.base + '/files/content?path=' + self.source)
            actual_version = content['version']
            self.versions[actual_version] = content['text']
            self.state['sourceVersion'] = actual_version
        if 'scheduleId' not in self.state:
            self.write_source(self.state.get('sourceVersion', 'missing'))
            schedule = self.call('/schedules', {'revision': 0,
                'resource': {'kind': 'instance', 'id': self.state['instanceId'], 'nodeId': self.state['nodeId']},
                'action': 'backup.create', 'args': {'path': self.source, 'compression': 'deflate',
                'consistency': {'mode': 'files'}}, 'cron': '* * * * *', 'timezone': 'Asia/Shanghai',
                'misfire': 'run_once', 'overlap': 'wait_one', 'enabled': True}, 'POST')['schedule']
            self.state['scheduleId'] = schedule['spec']['id']
            self.state['scheduleRemoved'] = False
        self.save()
        self.event('fixture_ready', nodeId=self.state['nodeId'], instanceId=self.state['instanceId'],
                   runId=self.state['runId'], scheduleId=self.state['scheduleId'])

    def write_source(self, version):
        text = '真实墙钟备份 ' + utc() + '\n' + str(uuid.uuid4()) + '\n'
        task = self.call(self.base + '/files/content', {'path': self.source, 'text': text, 'version': version}, 'PUT')['task']
        self.task(task['taskId'])
        version = self.call(self.base + '/files/stat?path=' + self.source)['version']
        self.versions[version] = text
        self.state['sourceVersion'] = version

    def all_schedule_tasks(self):
        before = 0
        results = []
        while True:
            page = self.call('/tasks?before=' + str(before) + '&limit=100')
            results.extend(task for task in page['items'] if task.get('payload', {}).get('scheduleId') == self.state['scheduleId'])
            before = page['nextBefore']
            if before < 0:
                return results

    def schedule_check(self):
        tasks = self.call('/tasks?before=0&limit=50')['items']
        for task in reversed(tasks):
            payload = task.get('payload', {})
            if payload.get('scheduleId') != self.state['scheduleId']:
                continue
            slot = payload['scheduleSlot']
            if slot in self.scheduled:
                if self.scheduled[slot]['taskId'] != task['taskId']:
                    raise RuntimeError('duplicate schedule slot accepted')
                continue
            task = self.task(task['taskId'])
            snapshot = task['result']['snapshot']
            version = payload['create']['version']
            if task['action'] != 'backup.create' or not task['requestId'].startswith('schedule:'):
                raise RuntimeError('scheduled task identity mismatch')
            if snapshot['state'] != 'ready' or version not in self.versions:
                raise RuntimeError('snapshot/source version mismatch')
            if dt.datetime.fromisoformat(slot.replace('Z', '+00:00')).timestamp() % 60:
                raise RuntimeError('scheduler slot is not a real minute boundary')
            self.scheduled[slot] = {'taskId': task['taskId'], 'snapshotId': snapshot['id'], 'version': version}
            self.event('minute_slot_succeeded', slot=slot, **self.scheduled[slot])
            self.write_source(self.state['sourceVersion'])
            self.task(self.call(self.base + '/backups/prune', {'path': self.source, 'keepLast': 3}, 'POST')['task']['taskId'])
        snapshots = self.call(self.base + '/backups')['items']
        owned = [s for s in snapshots if s['path'] == self.source and s['state'] == 'ready']
        if len(owned) > 3:
            raise RuntimeError('backup retention exceeds 3 ready snapshots')

    def resources(self):
        result = {}
        for name, entry in self.state['services'].items():
            expected = entry['identity']
            if not alive(expected):
                raise RuntimeError('service identity unexpectedly disappeared: ' + name)
            root = Path('/proc') / str(expected['pid'])
            status = (root / 'status').read_text().splitlines()
            rss = next(int(line.split()[1]) * 1024 for line in status if line.startswith('VmRSS:'))
            fds = len(list((root / 'fd').iterdir()))
            baseline = self.baseline.setdefault(name, {'rssBytes': rss, 'fds': fds})
            if rss > 256 * MIB or rss > baseline['rssBytes'] + 128 * MIB:
                raise RuntimeError('RSS budget exceeded: ' + name)
            if fds > baseline['fds'] + 128:
                raise RuntimeError('FD growth budget exceeded: ' + name)
            result[name] = {'rssBytes': rss, 'fds': fds, 'identity': expected}
            maximum = self.maxima.setdefault(name, {'rssBytes': 0, 'fds': 0})
            maximum['rssBytes'] = max(maximum['rssBytes'], rss)
            maximum['fds'] = max(maximum['fds'], fds)
        archives = []
        for directory in sorted((self.root / 'daemon').glob('logs/*/archive')):
            if not directory.is_dir():
                continue
            for _ in range(4):
                segments = sorted(directory.glob('*.seg'))
                try:
                    total = sum(p.stat().st_size for p in segments)
                except FileNotFoundError:
                    continue
                if total > 16 * MIB:
                    raise RuntimeError('production archive retention exceeded 16 MiB')
                archives.append({'directory': str(directory.relative_to(self.root)), 'bytes': total,
                    'segments': len(segments), 'first': segments[0].name if segments else None})
                self.maxima['archiveBytes'] = max(self.maxima.get('archiveBytes', 0), total)
                break
            else:
                raise RuntimeError('archive rotation sample remained unresolved')
        if not archives:
            raise RuntimeError('owned instance log archive is missing')
        for archive in archives:
            if Path(archive['directory']).parent.name != self.state['runId']:
                continue  # Retained archive from an explicitly stopped segment.
            capture = read(self.root / Path(archive['directory']).parent / 'capture.json')
            helper = identity(capture['identity']['pid'])
            if int(helper['startTicks']) != capture['identity']['birth'] or helper['bootId'] != capture['identity']['bootId']:
                raise RuntimeError('independent log helper identity mismatch')
            helper_root = Path('/proc') / str(helper['pid'])
            lines = (helper_root / 'status').read_text().splitlines()
            rss = next(int(line.split()[1]) * 1024 for line in lines if line.startswith('VmRSS:'))
            baseline = self.baseline.setdefault('logHelper', {'rssBytes': rss})
            if rss > 256 * MIB or rss > baseline['rssBytes'] + 128 * MIB:
                raise RuntimeError('independent log helper RSS budget exceeded')
            self.maxima['logHelperRSSBytes'] = max(self.maxima.get('logHelperRSSBytes', 0), rss)
            result['logHelper'] = {'identity': helper, 'rssBytes': rss}
        # A bounded evidence writer must not quietly conceal unbounded service
        # storage. Tiny scheduled sources/three snapshots allow a fixed cap.
        for _ in range(4):
            try:
                disk = sum(p.stat().st_size for p in self.root.rglob('*') if p.is_file() and 'bin' not in p.relative_to(self.root).parts)
                break
            except FileNotFoundError:
                continue
        else:
            raise RuntimeError('state directory sample remained unresolved during rotation')
        if disk > 256 * MIB:
            raise RuntimeError('private state/evidence disk budget exceeded 256 MiB')
        self.maxima['fixtureStateBytes'] = max(self.maxima.get('fixtureStateBytes', 0), disk)
        return {'processes': result, 'archives': archives, 'fixtureStateBytes': disk}

    def sample(self):
        sample_time = time.monotonic()
        if self.last_sample_mono is not None:
            gap = sample_time - self.last_sample_mono
            self.max_sample_gap = max(self.max_sample_gap, gap)
            if gap > max(15, 3 * self.config['interval']):
                raise RuntimeError('ordinary real sampling gap exceeded supervision budget')
        self.last_sample_mono = sample_time
        point = self.call('/nodes/' + self.state['nodeId'] + '/metrics')
        timestamp = dt.datetime.fromisoformat(point['observedAt'].replace('Z', '+00:00')).timestamp()
        if point.get('stale') or (self.histories and timestamp <= self.histories[-1][0]):
            raise RuntimeError('live node sample is stale/nonmonotonic')
        self.histories.append((timestamp, point['observedAt']))
        history = self.call('/nodes/' + self.state['nodeId'] + '/metrics/history')
        items = history['items']
        if history['retention'] != 120 or len(items) > 120:
            raise RuntimeError('node history retention mismatch')
        if len(self.histories) == 120 and [p['observedAt'] for p in items] != [p[1] for p in self.histories]:
            raise RuntimeError('node history no longer matches newest 120 real samples')
        instance = self.call(self.base + '/metrics')
        instance_at = dt.datetime.fromisoformat(instance['observedAt'].replace('Z', '+00:00')).timestamp()
        if instance.get('stale') or instance_at <= self.last_instance_at or instance['runId'] != self.state['runId']:
            raise RuntimeError('live instance metric identity/timestamp mismatch')
        self.last_instance_at = instance_at
        self.dates.add(point['observedAt'][:10])
        self.sample_count += 1
        resources = self.resources()
        self.samples.info(json.dumps({'utc': utc(), 'node': point, 'instance': instance, **resources}, ensure_ascii=False))
        self.schedule_check()
        self.save()

    def fault_phase(self, number):
        # No new metrics are sampled during this declared fault window. The
        # precise last 120 timestamps must survive offline cache and restart.
        if len(self.histories) != 120:
            raise RuntimeError('fault checkpoint requires 120 actual live samples')
        expected = [p[1] for p in self.histories]
        daemon = self.state['services']['daemon']['identity']
        old_master = self.state['services']['master']['identity']
        fault_start = time.time()
        self.event('fault_started', phase=number, daemon=daemon)
        owned_signal(daemon, signal.SIGSTOP)
        try:
            self.until(lambda: not self.nodes_online(), 70, 'paused Daemon OFFLINE')
            for after_restart in (False, True):
                if after_restart:
                    self.stop_service('master')
                    self.master()
                    if self.state['services']['master']['identity'] == old_master:
                        raise RuntimeError('Master restart identity did not change')
                history = self.call('/nodes/' + self.state['nodeId'] + '/metrics/history')
                if not history.get('stale') or not history.get('diagnostic') or history['retention'] != 120:
                    raise RuntimeError('offline cache diagnostic/retention missing')
                if [p['observedAt'] for p in history['items']] != expected or not all(p.get('stale') for p in history['items']):
                    raise RuntimeError('offline/Master-restart cache differs from real sampling window')
                instance = self.call(self.base + '/metrics')
                if not instance.get('stale') or instance['runId'] != self.state['runId']:
                    raise RuntimeError('offline instance cache lost run identity')
                self.event('stale_cache_verified', phase=number, afterMasterRestart=after_restart, points=len(expected))
        finally:
            owned_signal(daemon, signal.SIGCONT)
        self.until(self.nodes_online, 40, 'same Daemon ONLINE after SIGCONT')
        instance = self.call(self.base)['instance']
        if instance['runId'] != self.state['runId'] or instance['state'] != 'RUNNING':
            raise RuntimeError('fault created a replacement instance run')
        self.schedule_check()
        self.fault_windows.append((fault_start, time.time()))
        self.phases.append({'phase': number, 'completedAt': utc(), 'sameRunId': self.state['runId']})
        self.event('fault_completed', phase=number)
        self.last_sample_mono = None
        self.save()

    def final_checks(self):
        if self.sample_count < 121 or len(self.scheduled) < 2:
            raise RuntimeError('insufficient real monitor samples/minute slots')
        # End the plan only AFTER the requested wall-clock interval has
        # elapsed. This closes admission before a paginated final audit.
        self.remove_schedule()
        self.schedule_check()
        tasks = self.all_schedule_tasks()
        slots = [t['payload']['scheduleSlot'] for t in tasks]
        if len(set(slots)) != len(slots) or any(t['state'] != 'SUCCEEDED' for t in tasks):
            raise RuntimeError('final persisted task audit found duplicate/unfinished scheduled work')
        slot_seconds = {int(dt.datetime.fromisoformat(slot.replace('Z', '+00:00')).timestamp()) for slot in slots}
        end = time.time()
        required = set(range(int(self.start_wall // 60 + 1) * 60, int(end // 60) * 60 + 1, 60))
        # run_once deliberately collapses unavailable slots. Only the declared
        # fault windows (plus one reconciliation minute) are excluded; every
        # ordinary real minute must have a durable successful task.
        required = {slot for slot in required if not any(begin <= slot <= finish + 60 for begin, finish in self.fault_windows)}
        if not required.issubset(slot_seconds):
            raise RuntimeError('ordinary real minute has no accepted successful occurrence')
        newest = max(tasks, key=lambda t: t['payload']['scheduleSlot'])
        snapshot_id = newest['result']['snapshot']['id']
        restore_path = 'endurance-restored-' + uuid.uuid4().hex[:12] + '.txt'
        plan = self.call(self.base + '/restore-plans', {'backupId': snapshot_id, 'path': restore_path,
                        'version': 'missing'}, 'POST')['plan']
        self.task(self.call(self.base + '/restores', {'planId': plan['request']['id'], 'planHash': plan['hash'],
                       'overwritePlanHash': plan['hash']}, 'POST')['task']['taskId'])
        restored = self.call(self.base + '/files/content?path=' + restore_path)['text']
        if restored != self.versions[newest['payload']['create']['version']]:
            raise RuntimeError('actual archive restore does not match admitted source version')
        with sqlite3.connect('file:' + str(self.root / 'master/master.db') + '?mode=ro', uri=True) as db:
            integrity = db.execute('PRAGMA integrity_check').fetchone()[0]
            if integrity != 'ok':
                raise RuntimeError('private SQLite integrity check failed')
            # Instance history has no public history API: read only its own
            # cache rows to prove the same production persistence boundary.
            rows = db.execute("SELECT id FROM records WHERE namespace='monitor_instance_points_v1' AND id LIKE ? ORDER BY id",
                              (self.state['instanceId'] + '-%',)).fetchall()
            if len(rows) != 120 or any(rows[i][0] >= rows[i + 1][0] for i in range(len(rows) - 1)):
                raise RuntimeError('instance persistent history is not the ordered newest 120 records')
        if self.config['seconds'] >= 86400:
            if len(self.dates) < 2 or time.time() - self.start_wall < 86400 or time.monotonic() - self.started < 86400:
                raise RuntimeError('24h/cross-UTC-date actual wall-clock proof incomplete')
            declared_fault_seconds = sum(finish - begin for begin, finish in self.fault_windows)
            minimum_samples = int((self.config['seconds'] - declared_fault_seconds) / self.config['interval'] * .99)
            if self.sample_count < minimum_samples:
                raise RuntimeError('24h ordinary sampling coverage is below 99% of configured cadence')
            archives = self.resources()['archives']
            if not any(a['first'] and int(a['first'].removesuffix('.seg')) > 1 for a in archives):
                raise RuntimeError('24h log archive has not actually rotated')
        self.event('final_checks_succeeded', slots=len(slots), ordinaryMinuteSlots=len(required),
                   faultWindows=self.fault_windows, actualRestoreSnapshot=snapshot_id, sqliteIntegrity=integrity)

    def remove_schedule(self):
        if self.client and self.state.get('scheduleId'):
            current = next((s for s in self.call('/schedules')['items'] if s['spec']['id'] == self.state['scheduleId']), None)
            if current:
                self.call('/schedules/' + self.state['scheduleId'], {'revision': current['configRevision']}, 'DELETE')
            self.state['scheduleRemoved'] = True
            self.save()

    def cleanup(self):
        # Keep the private evidence/state; delete only this schedule and stop
        # the resource whose immutable identities were recorded at admission.
        interrupted = self.stopping
        self.stopping = False
        marker = self.root / 'stop.request'
        if marker.exists():
            marker.unlink()
        try:
            self.remove_schedule()
            if self.client and self.state.get('instanceId'):
                task = self.call('/instances/' + self.state['instanceId'] + '/actions', {'action': 'stop'}, 'POST')['task']
                self.task(task['taskId'])
        except Exception as error:
            self.event('cleanup_api_failed', errorType=type(error).__name__, message=str(error))
            self.state['cleanupError'] = str(error)
        finally:
            for name in ('daemon', 'master'):
                self.stop_service(name)
            self.stopping = interrupted
        self.event('cleanup_completed', servicesAlive=[name for name, entry in self.state.get('services', {}).items() if alive(entry['identity'])])

    def recover_cleanup(self):
        self.segment_start = utc()
        self.save('CLEANING')
        error = None
        (self.root / 'stop.request').unlink(missing_ok=True)
        try:
            master = self.state.get('services', {}).get('master')
            if master and alive(master['identity']):
                owned_signal(master['identity'], signal.SIGCONT)
                self.connect_client()
                self.login()
            elif (self.root / 'master/tls.crt').exists():
                self.master()
            daemon = self.state.get('services', {}).get('daemon')
            if daemon and alive(daemon['identity']):
                owned_signal(daemon['identity'], signal.SIGCONT)
            elif (self.root / 'daemon.json').exists():
                self.start_service('daemon', [str(self.root / 'bin/blora-daemon'), '--config', str(self.root / 'daemon.json')])
            if self.client and self.state.get('nodeId'):
                self.until(self.nodes_online, 40, 'owned node for interrupted-worker cleanup')
        except Exception as exception:
            error = type(exception).__name__ + ': ' + str(exception)
            self.state['cleanupError'] = error
        finally:
            self.cleanup()
            self.save('FAILED_CLEANUP' if self.state.get('cleanupError') else 'STOPPED', finishedAt=utc(), error=error)
        return 1 if self.state.get('cleanupError') else 0

    def run(self):
        self.segment_start = utc()
        for old_result in ('finishedAt', 'error', 'activeWait'):
            self.state.pop(old_result, None)
        status, error = 'FAILED', None
        try:
            self.setup()
            self.started, self.start_wall = time.monotonic(), time.time()
            self.segment_start = utc()
            next_sample = self.started
            phases = list(self.config['faultAfter'])
            while time.monotonic() - self.started < self.config['seconds']:
                self.check()
                if phases and time.monotonic() - self.started >= phases[0]:
                    phases.pop(0)
                    self.fault_phase(len(self.phases) + 1)
                self.sample()
                next_sample += self.config['interval']
                self.wait(max(0, next_sample - time.monotonic()))
                # A deliberate fault does not create a fake backlog of samples.
                next_sample = max(next_sample, time.monotonic())
            self.final_checks()
            if len(self.phases) != len(self.config['faultAfter']):
                raise RuntimeError('required fault checkpoints did not complete')
            status = 'PASSED_24H' if self.config['seconds'] >= 86400 else 'PASSED_SHORT'
        except Interrupted as exception:
            status, error = 'STOPPED', str(exception)
        except Exception as exception:
            error = type(exception).__name__ + ': ' + str(exception)
            self.event('failed', error=error)
        finally:
            self.cleanup()
            if self.state.get('cleanupError'):
                status = 'FAILED_CLEANUP'
            self.save(status, finishedAt=utc(), error=error)
        return 0 if status.startswith('PASSED') else 1


def validate_root(path):
    root = path.resolve(strict=True)
    if not root.name.startswith('endurance-') or not (root / 'run.json').is_file() or root.is_symlink():
        raise ValueError('select an explicitly created endurance-* run directory')
    if root.stat().st_uid != os.getuid() or root.stat().st_mode & 0o077:
        raise ValueError('run directory must be owned by this user and private (0700)')
    return root


def launch(root, action='_worker'):
    # The worker has its own session and no PTY/tool-owned pipes. Its private
    # identity+lock survives turn/session expiration, and status is durable.
    with (root / 'worker.log').open('ab') as output:
        worker = subprocess.Popen([sys.executable, str(Path(__file__).resolve()), action, str(root)],
                stdin=subprocess.DEVNULL, stdout=output, stderr=output, start_new_session=True, close_fds=True)
    atomic(root / 'worker.json', identity(worker.pid))
    print(json.dumps({'runDirectory': str(root), 'worker': identity(worker.pid), 'origin': read(root / 'run.json')['origin']}, indent=2))


def main():
    os.umask(0o077)
    if sys.platform != 'linux' or not hasattr(os, 'pidfd_open') or not hasattr(signal, 'pidfd_send_signal'):
        raise RuntimeError('requires Linux /proc and pidfd identity-safe signaling')
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='action', required=True)
    start = sub.add_parser('start')
    start.add_argument('--parent', type=Path, default=REPO / '.local')
    start.add_argument('--master', type=Path, default=REPO / 'dist/blora-master')
    start.add_argument('--daemon', type=Path, default=REPO / 'dist/blora-daemon')
    start.add_argument('--seconds', type=int, default=86400)
    start.add_argument('--interval', type=float, default=5)
    start.add_argument('--fault-after', type=float, nargs='*', help='actual elapsed seconds; default 6h and 18h for 24h')
    for action in ('status', 'stop', 'resume', '_worker', '_cleanup'):
        sub.add_parser(action).add_argument('root', type=Path)
    args = parser.parse_args()
    if args.action == 'start':
        if not 180 <= args.seconds <= 7 * 86400 or not 1 <= args.interval <= 30:
            parser.error('seconds must be 180..604800; interval must be 1..30')
        phases = args.fault_after if args.fault_after is not None else ([21600, 64800] if args.seconds >= 86400 else [])
        if phases != sorted(set(phases)) or any(p < 121 * args.interval or p + 110 >= args.seconds for p in phases):
            parser.error('each fault needs 121 previous samples and 110s before completion')
        args.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
        root = Path(tempfile.mkdtemp(prefix='endurance-', dir=args.parent)).resolve()
        (root / 'bin').mkdir(mode=0o700)
        hashes = {}
        for name, source in (('blora-master', args.master), ('blora-daemon', args.daemon)):
            target = root / 'bin' / name
            shutil.copyfile(source.resolve(strict=True), target)
            target.chmod(0o700)
            hashes[name] = hashlib.sha256(target.read_bytes()).hexdigest()
        with socket.socket() as listener:
            listener.bind(('127.0.0.1', 0))
            port = listener.getsockname()[1]
        atomic(root / 'run.json', {'createdAt': utc(), 'seconds': args.seconds, 'interval': args.interval,
            'faultAfter': phases, 'origin': f'https://127.0.0.1:{port}', 'binarySHA256': hashes,
            'entrySHA256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
            'budgets': {'serviceRSSBytes': 256 * MIB, 'serviceRSSGrowthBytes': 128 * MIB, 'fdGrowth': 128,
                        'archiveBytes': 16 * MIB, 'readySnapshots': 3, 'stateEvidenceDiskBytes': 256 * MIB}})
        launch(root)
        return 0
    root = validate_root(args.root)
    if args.action in ('_worker', '_cleanup'):
        with (root / 'worker.lock').open('a') as lock:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
            runner = Runner(root)
            return runner.run() if args.action == '_worker' else runner.recover_cleanup()
    expected = read(root / 'worker.json')
    if args.action == 'status':
        state = read(root / 'checkpoint.json') if (root / 'checkpoint.json').exists() else {'status': 'STARTING'}
        for private in ('versions',):
            state.pop(private, None)
        for entry in state.get('services', {}).values():
            entry.pop('command', None)
        try:
            state['workerAlive'] = alive(expected)
        except PermissionError:
            state['workerAlive'] = 'UNKNOWN: /proc identity is not readable in this sandbox'
        state['checkpointAgeSeconds'] = time.time() - (root / 'checkpoint.json').stat().st_mtime if (root / 'checkpoint.json').exists() else None
        print(json.dumps(state, ensure_ascii=False, indent=2))
    elif args.action == 'stop':
        (root / 'stop.request').touch(mode=0o600)
        if alive(expected):
            owned_signal(expected, signal.SIGTERM)
        else:
            launch(root, '_cleanup')
        print('Stop requested; poll status for STOPPED and cleanup evidence. No process with a different identity is signaled.')
    elif args.action == 'resume':
        if alive(expected):
            raise RuntimeError('worker is still alive; inspect/stop it before resuming')
        state = read(root / 'checkpoint.json')
        if state['status'] in ('PASSED_24H', 'PASSED_SHORT') or state.get('cleanupError'):
            raise RuntimeError('completed/cleanup-error run cannot be resumed; start a fresh isolated run')
        # cleanup removes the schedule on explicit stop; create a new one on
        # resume rather than preserving a nonexistent schedule identifier.
        if state['status'] == 'STOPPED' or state.get('scheduleRemoved'):
            state.pop('scheduleId', None)
            atomic(root / 'checkpoint.json', state)
        (root / 'stop.request').unlink(missing_ok=True)
        launch(root)
    return 0


if __name__ == '__main__':
    try:
        sys.exit(main())
    except Exception as exception:
        print(type(exception).__name__ + ': ' + str(exception), file=sys.stderr)
        sys.exit(1)
