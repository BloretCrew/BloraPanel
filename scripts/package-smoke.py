#!/usr/bin/env python3
"""Verify extracted Linux release startup and independent SDK build locally."""
import argparse
import http.cookiejar
import json
import os
from pathlib import Path
import secrets
import shutil
import signal
import socket
import ssl
import subprocess
import tarfile
import tempfile
import time
import uuid
import urllib.request
import urllib.error


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('release', type=Path)
    parser.add_argument('--master-crash', action='store_true', help='SIGKILL the owned Master during a real restart task')
    parser.add_argument('--daemon-crash', action='store_true', help='Also SIGKILL the owning Daemon while its instance is running')
    parser.add_argument('--daemon-task-crash', action='store_true', help='SIGKILL the owning Daemon in the STOPPING phase and verify interruption')
    parser.add_argument('--scheduled-backup-crash', action='store_true', help='SIGKILL Master and the owning Daemon after a real scheduled backup occurrence is accepted')
    parser.add_argument('--state-restore', action='store_true', help='Back up stopped owned state, change it, restore, and verify identities/permissions/files')
    parser.add_argument('--upgrade-from', type=Path, help='Initialize and seed with this older Linux release, then upgrade the same stopped state to the current release')
    parser.add_argument('--rollback-release', type=Path, help='Use this compatible older Linux release after restoring the stopped snapshot')
    args = parser.parse_args()
    if args.rollback_release and not args.state_restore:
        parser.error('--rollback-release requires --state-restore')
    if args.upgrade_from and not args.state_restore:
        parser.error('--upgrade-from requires --state-restore')
    if args.state_restore and (args.master_crash or args.daemon_crash or args.daemon_task_crash or args.scheduled_backup_crash):
        parser.error('Run stopped-state restoration separately from crash scenarios')
    if args.scheduled_backup_crash and (args.master_crash or args.daemon_crash or args.daemon_task_crash):
        parser.error('Run scheduled-backup crash recovery separately from instance lifecycle crash scenarios')
    release = args.release.resolve()
    root = Path(tempfile.mkdtemp(prefix='blora-release-smoke-'))
    processes, logs = [], []
    print(f'Test directory: {root}', flush=True)
    try:
        for component in ('master', 'daemon', 'sdk'):
            pattern = f'blora-{component}-*.tar.gz' if component == 'sdk' else f'blora-{component}-*-linux-amd64.tar.gz'
            matches = list(release.glob(pattern))
            if len(matches) != 1:
                raise ValueError(f'Expected one {component} archive')
            with tarfile.open(matches[0]) as archive:
                archive.extractall(root / component, filter='data')
        initial_master, initial_daemon = root / 'master', root / 'daemon'
        if args.upgrade_from:
            older = args.upgrade_from.resolve()
            if older == release:
                raise ValueError('Upgrade source must differ from the current release')
            for component in ('master', 'daemon'):
                matches = list(older.glob(f'blora-{component}-*-linux-amd64.tar.gz'))
                if len(matches) != 1:
                    raise ValueError(f'Expected one upgrade-source {component} archive')
                with tarfile.open(matches[0]) as archive:
                    archive.extractall(root / ('upgrade-from-' + component), filter='data')
            initial_master, initial_daemon = root / 'upgrade-from-master', root / 'upgrade-from-daemon'
        env = dict(os.environ, GOCACHE='/tmp/blora-go-grant-idem', GOMODCACHE='/tmp/blora-go-mod', GOPATH='/tmp/blora-go', npm_config_cache='/tmp/blora-npm-cache')
        for directory in (root / 'sdk/sdk', root / 'sdk/sdk/examples/reference-app'):
            for command in (['npm', 'ci', '--offline'], ['npm', 'run', 'build' if directory.name == 'sdk' else 'package']):
                subprocess.run(command, cwd=directory, env=env, check=True, timeout=120)
        key = root / 'publisher.pem'
        subprocess.run(['openssl', 'genpkey', '-algorithm', 'ED25519', '-out', str(key)], check=True)
        key.chmod(0o600)
        subprocess.run([str(root / 'sdk/tools/blora-extension-sign'), '--package', str(root / 'sdk/sdk/examples/reference-app/reference.blora-extension.json'), '--key', str(key), '--out', str(root / 'signed.json')], check=True)
        password = secrets.token_urlsafe(32)
        password_file = root / 'initial-password'
        password_file.touch(mode=0o600)
        password_file.write_text(password)
        state = root / 'state'
        binary = str(initial_master / 'blora-master')
        subprocess.run([binary, '--state-dir', str(state), '--init', '--password-file', str(password_file)], cwd=initial_master, check=True, timeout=30)
        with socket.socket() as listener:
            listener.bind(('127.0.0.1', 0))
            port = listener.getsockname()[1]
        origin = f'https://127.0.0.1:{port}'

        def start(name, command, cwd):
            log = (root / f'{name}.log').open('wb')
            logs.append(log)
            processes.append(subprocess.Popen(command, cwd=cwd, stdout=log, stderr=log))

        start('master', [binary, '--state-dir', str(state), '--listen', f'127.0.0.1:{port}', '--origin', origin, '--static-dir', 'web/dist'], initial_master)
        client = urllib.request.build_opener(urllib.request.HTTPSHandler(context=ssl.create_default_context(cafile=str(state / 'tls.crt'))), urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
        csrf = ''

        def call(path, value=None, request_id=None, method=None):
            request = urllib.request.Request(origin + path, data=None if value is None else json.dumps(value).encode(), headers={'Content-Type': 'application/json', 'Origin': origin, 'X-CSRF-Token': csrf, 'Idempotency-Key': request_id or str(uuid.uuid4())}, method=method)
            with client.open(request, timeout=5) as response:
                return response.read()

        for attempt in range(100):
            try:
                call('/healthz')
                break
            except OSError:
                if attempt == 99:
                    raise
                time.sleep(.1)
        assert b'<html' in call('/')
        csrf = json.loads(call('/api/v1/login', {'name': 'admin', 'password': password}))['csrfToken']
        for index in range(2):
            token = json.loads(call('/api/v1/nodes/enrollments', {'name': f'package-node-{index}'}))['token']
            ticket = root / f'node-{index}.ticket'
            ticket.touch(mode=0o600)
            ticket.write_text(token)
            config = root / f'node-{index}.json'
            config.touch(mode=0o600)
            config.write_text(json.dumps({'stateDir': str(root / f'node-{index}'), 'masterUrl': origin, 'caFile': str(state / 'tls.crt'), 'enrollmentFile': str(ticket), 'allowPGIDFallback': True}))
            start(f'daemon-{index}', [str(initial_daemon / 'blora-daemon'), '--config', str(config)], initial_daemon)
        for attempt in range(100):
            nodes = json.loads(call('/api/v1/nodes'))['items']
            if len(nodes) == 2 and all(node['state'] == 'ONLINE' for node in nodes):
                break
            if attempt == 99:
                raise RuntimeError('Packaged Daemons did not become ONLINE')
            time.sleep(.2)
        print('PASS: independent SDK build/signing, Master initialization/TLS/static/login, two packaged Daemons ONLINE', flush=True)
        if args.state_restore:
            owner = nodes[0]
            owner_index = int(owner['name'].removeprefix('package-node-'))
            directory = root / f'node-{owner_index}' / 'restore-resource'
            directory.mkdir(mode=0o700)
            marker = directory / 'restore-marker.txt'
            marker.write_text('snapshot resource content\n', encoding='utf-8')
            instance = json.loads(call('/api/v1/instances', {'nodeId': owner['nodeId'], 'name': 'snapshot-instance',
                'config': {'mode': 'native', 'directory': str(directory), 'command': ['/bin/sh', '-c', 'exit 0'],
                           'stopSeconds': 2, 'killSeconds': 2, 'escalate': True}}))['instance']
            file_endpoint = '/api/v1/instances/' + instance['instanceId'] + '/files/actions'
            file_request = {'action': 'mkdir', 'path': 'task-created', 'target': 'task-created', 'targetVersion': 'missing'}
            file_key = str(uuid.uuid4())
            file_task = json.loads(call(file_endpoint, file_request, file_key))['task']
            deadline = time.monotonic() + 20
            while time.monotonic() < deadline:
                file_task = json.loads(call('/api/v1/tasks/' + file_task['taskId']))['task']
                if file_task['state'] == 'SUCCEEDED':
                    break
                if file_task['state'] in ('FAILED', 'INTERRUPTED', 'CANCELLED'):
                    raise RuntimeError('Snapshot seed file task did not succeed')
                time.sleep(.1)
            assert file_task['state'] == 'SUCCEEDED' and (directory / 'task-created').is_dir()
            member_password = secrets.token_urlsafe(32)
            member = json.loads(call('/api/v1/users', {'name': 'snapshot-member', 'password': member_password, 'admin': False}))['user']
            resource = {'kind': 'instance', 'id': instance['instanceId'], 'nodeId': owner['nodeId']}
            call('/api/v1/grants', {'userId': member['userId'], 'resource': resource, 'action': 'instance.read'})
            expected_ids = {node['nodeId'] for node in nodes}
            certificate = (state / 'tls.crt').read_bytes()

            def stop_owned():
                for process in reversed(processes):
                    if process.poll() is None:
                        process.send_signal(signal.SIGINT)
                for process in reversed(processes):
                    if process.wait(timeout=15) != 0:
                        raise RuntimeError('Cannot form a stopped snapshot after an unclean exit')

            def start_owned(master_dir, daemon_dir):
                start('restore-master-' + str(len(processes)), [str(master_dir / 'blora-master'), '--state-dir', str(state),
                    '--listen', f'127.0.0.1:{port}', '--origin', origin, '--static-dir', 'web/dist'], master_dir)
                for index in range(2):
                    start(f'restore-daemon-{index}-' + str(len(processes)), [str(daemon_dir / 'blora-daemon'), '--config', str(root / f'node-{index}.json')], daemon_dir)
                deadline = time.monotonic() + 30
                while time.monotonic() < deadline:
                    try:
                        current = json.loads(call('/api/v1/nodes'))['items']
                        if {node['nodeId'] for node in current} == expected_ids and all(node['state'] == 'ONLINE' for node in current):
                            return
                    except OSError:
                        pass
                    time.sleep(.2)
                raise RuntimeError('Restored node identities did not return ONLINE')

            if args.upgrade_from:
                # The seed was created by the older extracted programs. Start
                # the current extracted programs against exactly those stopped
                # directories/configurations before making any restore copy.
                stop_owned()
                start_owned(root / 'master', root / 'daemon')
                assert (state / 'tls.crt').read_bytes() == certificate
                assert b'<html' in call('/')
                csrf = json.loads(call('/api/v1/login', {'name': 'admin', 'password': password}))['csrfToken']
                users = json.loads(call('/api/v1/users'))['items']
                assert any(user['userId'] == member['userId'] and user['name'] == 'snapshot-member' for user in users)
                upgraded = json.loads(call('/api/v1/instances/' + instance['instanceId']))['instance']
                assert upgraded['name'] == 'snapshot-instance' and upgraded['state'] == 'STOPPED'
                assert marker.read_text(encoding='utf-8') == 'snapshot resource content\n'
                assert json.loads(call('/api/v1/instances/' + instance['instanceId'] + '/files/content?path=restore-marker.txt'))['text'] == 'snapshot resource content\n'
                replayed_task = json.loads(call(file_endpoint, file_request, file_key))['task']
                assert replayed_task['taskId'] == file_task['taskId'] and replayed_task['state'] == 'SUCCEEDED'
                assert (directory / 'task-created').is_dir()
                member_client = urllib.request.build_opener(urllib.request.HTTPSHandler(context=ssl.create_default_context(cafile=str(state / 'tls.crt'))), urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
                request = urllib.request.Request(origin + '/api/v1/login', data=json.dumps({'name': 'snapshot-member', 'password': member_password}).encode(), headers={'Content-Type': 'application/json', 'Origin': origin})
                with member_client.open(request, timeout=5) as response:
                    member_csrf = json.loads(response.read())['csrfToken']
                with member_client.open(origin + '/api/v1/instances', timeout=5) as response:
                    assert [item['instanceId'] for item in json.loads(response.read())['items']] == [instance['instanceId']]
                request = urllib.request.Request(origin + '/api/v1/instances/' + instance['instanceId'] + '/actions', data=b'{"action":"start"}', headers={'Content-Type': 'application/json', 'Origin': origin, 'X-CSRF-Token': member_csrf, 'Idempotency-Key': str(uuid.uuid4())})
                try:
                    member_client.open(request, timeout=5).close()
                except urllib.error.HTTPError as error:
                    assert error.code == 403
                    error.close()
                else:
                    raise AssertionError('Upgraded read-only grant unexpectedly allowed start')
                print('PASS: older packaged Master/Daemons upgraded in place to current package; TLS, both node identities, login, read-only grant, task receipt and resource bytes retained', flush=True)

            # All resources are stopped; directories and configurations belong
            # solely to this randomly-created test root. No live DB is copied.
            stop_owned()
            backup = root / 'stopped-snapshot'
            backup.mkdir(mode=0o700)
            for name in ('state', 'node-0', 'node-1'):
                shutil.copytree(root / name, backup / name)
            for index in range(2):
                shutil.copy2(root / f'node-{index}.json', backup / f'node-{index}.json')
            start_owned(root / 'master', root / 'daemon')
            call('/api/v1/users', {'name': 'after-snapshot', 'password': secrets.token_urlsafe(32), 'admin': False})
            marker.write_text('changed after backup\n', encoding='utf-8')
            assert any(user['name'] == 'after-snapshot' for user in json.loads(call('/api/v1/users'))['items'])
            stop_owned()
            retained = root / 'after-snapshot-state'
            retained.mkdir(mode=0o700)
            for name in ('state', 'node-0', 'node-1'):
                (root / name).rename(retained / name)
                shutil.copytree(backup / name, root / name)
            for index in range(2):
                shutil.copy2(backup / f'node-{index}.json', root / f'node-{index}.json')
            master_dir, daemon_dir = root / 'master', root / 'daemon'
            if args.rollback_release:
                for component in ('master', 'daemon'):
                    matches = list(args.rollback_release.resolve().glob(f'blora-{component}-*-linux-amd64.tar.gz'))
                    if len(matches) != 1:
                        raise ValueError(f'Expected one rollback {component} archive')
                    with tarfile.open(matches[0]) as archive:
                        archive.extractall(root / ('rollback-' + component), filter='data')
                master_dir, daemon_dir = root / 'rollback-master', root / 'rollback-daemon'
            assert (state / 'tls.crt').read_bytes() == certificate
            start_owned(master_dir, daemon_dir)
            assert b'<html' in call('/')
            assert marker.read_text(encoding='utf-8') == 'snapshot resource content\n'
            users = json.loads(call('/api/v1/users'))['items']
            assert not any(user['name'] == 'after-snapshot' for user in users)
            assert any(user['userId'] == member['userId'] and user['name'] == 'snapshot-member' for user in users)
            restored = json.loads(call('/api/v1/instances/' + instance['instanceId']))['instance']
            assert restored['name'] == 'snapshot-instance' and restored['state'] == 'STOPPED'
            assert json.loads(call('/api/v1/instances/' + instance['instanceId'] + '/files/content?path=restore-marker.txt'))['text'] == 'snapshot resource content\n'
            replayed_task = json.loads(call(file_endpoint, file_request, file_key))['task']
            assert replayed_task['taskId'] == file_task['taskId'] and replayed_task['state'] == 'SUCCEEDED'
            assert (directory / 'task-created').is_dir()
            member_client = urllib.request.build_opener(urllib.request.HTTPSHandler(context=ssl.create_default_context(cafile=str(state / 'tls.crt'))), urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
            request = urllib.request.Request(origin + '/api/v1/login', data=json.dumps({'name': 'snapshot-member', 'password': member_password}).encode(), headers={'Content-Type': 'application/json', 'Origin': origin})
            with member_client.open(request, timeout=5) as response:
                member_csrf = json.loads(response.read())['csrfToken']
            with member_client.open(origin + '/api/v1/instances', timeout=5) as response:
                assert [item['instanceId'] for item in json.loads(response.read())['items']] == [instance['instanceId']]
            request = urllib.request.Request(origin + '/api/v1/instances/' + instance['instanceId'] + '/actions', data=b'{"action":"start"}', headers={'Content-Type': 'application/json', 'Origin': origin, 'X-CSRF-Token': member_csrf, 'Idempotency-Key': str(uuid.uuid4())})
            try:
                member_client.open(request, timeout=5).close()
            except urllib.error.HTTPError as error:
                assert error.code == 403
                error.close()
            else:
                raise AssertionError('Restored read-only grant unexpectedly allowed start')
            print('PASS: stopped-state backup/restore retained TLS, both node identities, user login, read-only grant, task receipt and resource bytes; later mutations absent', flush=True)
            if args.rollback_release:
                print('PASS: compatible older packaged Master/Daemons opened the restored snapshot with the same identities and permissions', flush=True)
            return
        if args.scheduled_backup_crash:
            owner = nodes[0]
            owner_index = int(owner['name'].removeprefix('package-node-'))
            source = root / 'scheduled-backup-source'
            source.mkdir(mode=0o700)
            # A moderately large real file makes it likely that SIGKILL lands
            # while the accepted task is still crossing Master/Daemon storage.
            (source / 'snapshot.bin').write_bytes(os.urandom(32 << 20))
            instance = json.loads(call('/api/v1/instances', {
                'nodeId': owner['nodeId'],
                'name': 'scheduled-crash-backup',
                'config': {'mode': 'native', 'directory': str(source), 'command': ['/bin/sh', '-c', 'exit 0'],
                           'stopSeconds': 2, 'killSeconds': 2, 'escalate': True},
            }))['instance']
            resource = {'kind': 'instance', 'id': instance['instanceId'], 'nodeId': owner['nodeId']}
            backup_args = {'path': 'snapshot.bin', 'compression': 'store', 'consistency': {'mode': 'files'}}
            schedule = json.loads(call('/api/v1/schedules', {
                'resource': resource,
                'action': 'backup.create',
                'args': backup_args,
                'cron': '* * * * *',
                'timezone': 'UTC',
                'misfire': 'run_once',
                'overlap': 'wait_one',
                'enabled': True,
            }, str(uuid.uuid4())))['schedule']
            schedule_id = schedule['spec']['id']

            def read_schedule():
                current = json.loads(call('/api/v1/schedules'))['items']
                return next((item for item in current if item['spec']['id'] == schedule_id), None)

            deadline = time.monotonic() + 70
            accepted = None
            while time.monotonic() < deadline:
                accepted = read_schedule()
                if accepted and accepted.get('lastTaskId'):
                    break
                time.sleep(.1)
            if not accepted or not accepted.get('lastTaskId'):
                raise RuntimeError('The real minute schedule did not accept an occurrence within 70 seconds')
            scheduled_task_id = accepted['lastTaskId']

            def read_scheduled_task():
                try:
                    return json.loads(call('/api/v1/tasks/' + scheduled_task_id))['task']
                except urllib.error.HTTPError as error:
                    if error.code == 404:
                        error.close()
                        return None
                    raise

            task_before = read_scheduled_task()
            if not task_before or task_before.get('taskId') != scheduled_task_id or not task_before.get('requestId'):
                raise RuntimeError(f'Scheduled task receipt was not readable before crash: {task_before}')
            request_id = task_before['requestId']
            old_startup = owner['startupId']
            # Kill both independently owned processes without a shutdown path.
            processes[0].kill()
            processes[owner_index + 1].kill()
            if processes[0].wait(timeout=5) != -signal.SIGKILL:
                raise RuntimeError('Master did not terminate from SIGKILL')
            if processes[owner_index + 1].wait(timeout=5) != -signal.SIGKILL:
                raise RuntimeError('Owning Daemon did not terminate from SIGKILL')
            start('scheduled-crash-master-restarted', [binary, '--state-dir', str(state), '--listen', f'127.0.0.1:{port}', '--origin', origin, '--static-dir', 'web/dist'], root / 'master')
            start('scheduled-crash-daemon-restarted', [str(root / 'daemon/blora-daemon'), '--config', str(root / f'node-{owner_index}.json')], root / 'daemon')

            def cluster_recovered():
                try:
                    current = json.loads(call('/api/v1/nodes'))['items']
                    by_id = {node['nodeId']: node for node in current}
                    return len(current) == 2 and all(node['state'] == 'ONLINE' for node in current) and by_id[owner['nodeId']]['startupId'] != old_startup
                except (KeyError, OSError, urllib.error.URLError):
                    return False

            recovery_deadline = time.monotonic() + 30
            while time.monotonic() < recovery_deadline and not cluster_recovered():
                time.sleep(.1)
            if not cluster_recovered():
                raise RuntimeError('Master and owning Daemon did not recover after scheduled task crash')
            after_restart = read_schedule()
            if not after_restart or after_restart.get('lastTaskId') != scheduled_task_id:
                raise RuntimeError(f'Master/Daemon restart changed accepted task identity: {after_restart}')
            # Prevent a later minute from adding a second valid occurrence to
            # this test while leaving the already-accepted task untouched.
            disabled = {
                'revision': after_restart['configRevision'],
                'resource': resource,
                'action': 'backup.create',
                'args': backup_args,
                'cron': '* * * * *',
                'timezone': 'UTC',
                'misfire': 'run_once',
                'overlap': 'wait_one',
                'enabled': False,
            }
            call('/api/v1/schedules/' + schedule_id, disabled, str(uuid.uuid4()), method='PUT')
            terminal = {'SUCCEEDED', 'FAILED', 'INTERRUPTED', 'CANCELLED'}
            task_after = None
            deadline = time.monotonic() + 45
            while time.monotonic() < deadline:
                task_after = read_scheduled_task()
                if task_after and task_after.get('state') in terminal:
                    break
                time.sleep(.1)
            if not task_after or task_after.get('state') not in terminal:
                raise RuntimeError(f'Accepted task remained ambiguous after Master/Daemon recovery: {task_after}')
            if task_after.get('taskId') != scheduled_task_id or task_after.get('requestId') != request_id:
                raise RuntimeError(f'Accepted task receipt changed after recovery: before={task_before} after={task_after}')
            recovered_schedule = read_schedule()
            if not recovered_schedule or recovered_schedule.get('lastTaskId') != scheduled_task_id:
                raise RuntimeError(f'Restart created another scheduled occurrence: {recovered_schedule}')
            backups = json.loads(call('/api/v1/instances/' + instance['instanceId'] + '/backups'))['items']
            if task_after['state'] == 'SUCCEEDED' and len(backups) != 1:
                raise RuntimeError(f'Succeeded scheduled task did not produce exactly one archive: {backups}')
            print(f"PASS: actual minute backup accepted as {scheduled_task_id}; Master and owning Daemon both SIGKILLed; startup identity changed; same task/request receipt recovered terminal as {task_after['state']}; schedule did not create another occurrence; archives={len(backups)}", flush=True)
            return
        if args.master_crash or args.daemon_crash or args.daemon_task_crash:
            run_root = root / 'run'
            run_root.mkdir(mode=0o700)
            instance = json.loads(call('/api/v1/instances', {'nodeId': nodes[0]['nodeId'], 'name': 'owned-crash-run', 'config': {'mode': 'native', 'directory': str(run_root), 'command': ['/bin/sh', '-c', "printf x >> generations; trap '' TERM; sleep 20 & wait"], 'stopSeconds': 3, 'killSeconds': 2, 'escalate': True}}))['instance']
            endpoint = '/api/v1/instances/' + instance['instanceId']

            def until(check, seconds=20):
                deadline = time.monotonic() + seconds
                while time.monotonic() < deadline:
                    if check():
                        return
                    time.sleep(.1)
                raise RuntimeError('Crash recovery condition timed out')

            def action(name, key=None):
                return json.loads(call(endpoint + '/actions', {'action': name}, key))['task']

            def task(task_id):
                return json.loads(call('/api/v1/tasks/' + task_id))['task']

            try:
                initial = action('start')
                until(lambda: task(initial['taskId'])['state'] == 'SUCCEEDED')
                until(lambda: (run_root / 'generations').read_text() == 'x')
                if args.daemon_crash:
                    owner = nodes[0]
                    index = int(owner['name'].removeprefix('package-node-'))
                    original_run = json.loads(call(endpoint))['instance']['runId']
                    old_startup = owner['startupId']
                    processes[index + 1].kill()
                    assert processes[index + 1].wait(timeout=5) == -signal.SIGKILL
                    start('daemon-restarted', [str(root / 'daemon/blora-daemon'), '--config', str(root / f'node-{index}.json')], root / 'daemon')
                    def daemon_ready():
                        current_nodes = json.loads(call('/api/v1/nodes'))['items']
                        return any(node['nodeId'] == owner['nodeId'] and node['state'] == 'ONLINE' and node['startupId'] != old_startup for node in current_nodes)
                    until(daemon_ready)
                    recovered = json.loads(call(endpoint))['instance']
                    assert recovered['runId'] == original_run and recovered['state'] == 'RUNNING'
                    assert (run_root / 'generations').read_text() == 'x'
                    print('PASS: Daemon SIGKILL; new startup identity recovered original run without another process start', flush=True)
                key = str(uuid.uuid4())
                restarting = action('restart', key)
                if args.daemon_task_crash:
                    until(lambda: task(restarting['taskId'])['phase'] == 'STOPPING')
                    owner = nodes[0]
                    index = int(owner['name'].removeprefix('package-node-'))
                    original_run = json.loads(call(endpoint))['instance']['runId']
                    process = processes[-1] if args.daemon_crash else processes[index + 1]
                    process.kill()
                    assert process.wait(timeout=5) == -signal.SIGKILL
                    start('daemon-task-restarted', [str(root / 'daemon/blora-daemon'), '--config', str(root / f'node-{index}.json')], root / 'daemon')
                    until(lambda: task(restarting['taskId'])['state'] == 'INTERRUPTED')
                    interrupted = task(restarting['taskId'])
                    assert interrupted['phase'] == 'daemon_restarted_reconcile_required'
                    assert action('restart', key)['taskId'] == restarting['taskId']
                    assert task(restarting['taskId'])['state'] == 'INTERRUPTED'
                    assert json.loads(call(endpoint))['instance']['runId'] == original_run
                    assert (run_root / 'generations').read_text() == 'x'
                    print('PASS: Daemon SIGKILL in STOPPING; original task INTERRUPTED, same-key retry does not replay restart', flush=True)
                    return
                until(lambda: task(restarting['taskId'])['state'] == 'RUNNING')
                processes[0].kill()
                assert processes[0].wait(timeout=5) == -signal.SIGKILL
                start('master-restarted', [binary, '--state-dir', str(state), '--listen', f'127.0.0.1:{port}', '--origin', origin, '--static-dir', 'web/dist'], root / 'master')
                def ready():
                    try:
                        return json.loads(call('/api/v1/nodes'))['items'] and all(node['state'] == 'ONLINE' for node in json.loads(call('/api/v1/nodes'))['items'])
                    except OSError:
                        return False
                until(ready)
                assert action('restart', key)['taskId'] == restarting['taskId']
                until(lambda: task(restarting['taskId'])['state'] == 'SUCCEEDED')
                until(lambda: (run_root / 'generations').read_text() == 'xx')
                before = json.loads(call(endpoint))['instance']['runId']
                assert action('restart', key)['taskId'] == restarting['taskId']
                assert json.loads(call(endpoint))['instance']['runId'] == before
                assert (run_root / 'generations').read_text() == 'xx'
                print('PASS: Master SIGKILL during restart; original task/run retained; exactly two process starts', flush=True)
            finally:
                stopped = action('stop')
                until(lambda: task(stopped['taskId'])['state'] == 'SUCCEEDED')
    finally:
        for process in reversed(processes):
            if process.poll() is None:
                process.send_signal(signal.SIGINT)
        for process in reversed(processes):
            try:
                process.wait(timeout=15)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
        for log in logs:
            log.close()
        print('Owned test processes stopped; private diagnostic directory retained.', flush=True)


if __name__ == '__main__':
    main()
