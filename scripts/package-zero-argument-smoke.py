#!/usr/bin/env python3
"""Exercise extracted Linux release binaries with sibling JSON and no arguments."""
import argparse
import http.cookiejar
import json
import os
from pathlib import Path
import secrets
import signal
import socket
import subprocess
import tarfile
import tempfile
import time
import urllib.error
import urllib.request


def wait_for(check, label):
    deadline = time.monotonic() + 45
    while time.monotonic() < deadline:
        try:
            value = check()
            if value:
                return value
        except (OSError, urllib.error.URLError):
            pass
        time.sleep(.1)
    raise RuntimeError('Timed out: ' + label)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('release', type=Path)
    args = parser.parse_args()
    release = args.release.resolve()
    if not Path('/proc/self/stat').exists():
        parser.error('This packaged runtime check requires Linux')
    processes, logs = [], []
    with tempfile.TemporaryDirectory(prefix='blora-packaged-json-') as temporary:
        root = Path(temporary)
        root.chmod(0o700)
        unrelated = root / 'unrelated-cwd'
        unrelated.mkdir()
        try:
            infos = {}
            for component in ('master', 'daemon'):
                archives = list(release.glob(f'blora-{component}-*-linux-amd64.tar.gz'))
                if len(archives) != 1:
                    raise ValueError('Expected exactly one archive: ' + component)
                target = root / component
                with tarfile.open(archives[0]) as archive:
                    archive.extractall(target, filter='data')
                binary = target / ('blora-' + component)
                infos[component] = json.loads(subprocess.check_output([str(binary), '--core-update-info']))
                assert infos[component]['component'] == component
                assert infos[component]['managedRestart'] and infos[component]['preserveInstances']
                expected = json.loads((target / 'CORE-UPDATE.json').read_text())
                for key in ('version', 'revision', 'schemaFingerprint', 'protocolVersion'):
                    assert infos[component][key] == expected[key]
            assert infos['master']['version'] == infos['daemon']['version']
            assert infos['master']['revision'] == infos['daemon']['revision']
            with socket.socket() as reserve:
                reserve.bind(('127.0.0.1', 0))
                port = reserve.getsockname()[1]
            origin = f'http://127.0.0.1:{port}'
            master, daemon = root / 'master', root / 'daemon'
            password = secrets.token_urlsafe(32)

            def private_file(path, text):
                path.touch(mode=0o600)
                path.write_text(text, encoding='utf-8')

            private_file(master / 'initial-password', password)
            private_file(master / 'master.json', json.dumps({'listen': f'127.0.0.1:{port}'}))
            csrf = ''
            client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))

            def api(path, value=None):
                request = urllib.request.Request(origin + path,
                    data=None if value is None else json.dumps(value).encode(),
                    headers={'Content-Type': 'application/json', 'Origin': origin, 'X-CSRF-Token': csrf})
                with client.open(request, timeout=5) as response:
                    payload = response.read()
                    return json.loads(payload) if payload.startswith((b'{', b'[')) else payload

            def start(component):
                output = tempfile.TemporaryFile()
                logs.append(output)
                process = subprocess.Popen([str(root / component / ('blora-' + component))],
                    cwd=unrelated, stdout=output, stderr=output)
                processes.append(process)

            def stop():
                for process in reversed(processes):
                    if process.poll() is None:
                        process.send_signal(signal.SIGINT)
                for process in reversed(processes):
                    assert process.wait(timeout=40) == 0, 'Packaged management service did not stop cleanly'
                processes.clear()

            start('master')
            wait_for(lambda: api('/healthz'), 'packaged Master HTTP startup')
            assert b'<html' in api('/'), 'Missing bundled frontend'
            csrf = api('/api/v1/login', {'name': 'admin', 'password': password})['csrfToken']
            token = api('/api/v1/nodes/enrollments', {'name': 'packaged-json-node'})['token']
            private_file(daemon / 'node.enrollment', token)
            private_file(daemon / 'daemon.json', json.dumps({'masterUrl': origin,
                'enrollmentFile': 'node.enrollment', 'allowPGIDFallback': False}))
            start('daemon')

            def online():
                nodes = api('/api/v1/nodes')['items']
                return nodes[0] if len(nodes) == 1 and nodes[0]['state'] == 'ONLINE' else None

            node_id = wait_for(online, 'packaged Daemon enrollment')['nodeId']
            assert (master / 'state/master').is_dir()
            assert (daemon / 'state/daemon').is_dir()
            assert not (unrelated / 'state').exists(), 'State resolved against cwd instead of installation'
            assert api('/api/v1/system/updates')['version'] == infos['master']['version']
            status = wait_for(lambda: api(f'/api/v1/nodes/{node_id}/updates/status'), 'Daemon updater status')
            assert status['version'] == infos['daemon']['version']
            stop()
            (master / 'initial-password').unlink()
            if (daemon / 'node.enrollment').exists():
                (daemon / 'node.enrollment').unlink()
            start('master')
            wait_for(lambda: api('/healthz'), 'Master identity reuse')
            csrf = api('/api/v1/login', {'name': 'admin', 'password': password})['csrfToken']
            start('daemon')
            assert wait_for(online, 'Daemon identity reuse')['nodeId'] == node_id
            assert api('/api/v1/system/updates')['revision'] == infos['master']['revision']
            stop()
            print(json.dumps({'outcome': 'PASS', 'platform': 'linux/amd64',
                'version': infos['master']['version'], 'revision': infos['master']['revision'],
                'checks': ['no-argument-sibling-json', 'unrelated-cwd-default-state',
                    'http-bundled-web-login', 'daemon-enrollment-and-update-status',
                    'restart-without-initial-password-or-enrollment-reuses-identities'],
                'scope': 'Actual extracted release programs, private state, no production deployment'}, indent=2))
        finally:
            for process in reversed(processes):
                if process.poll() is None:
                    process.send_signal(signal.SIGINT)
            for process in reversed(processes):
                try:
                    process.wait(timeout=40)
                except subprocess.TimeoutExpired:
                    # This runner never creates managed instances. Signal only
                    # its own stable launcher; no unrelated process-tree kill.
                    process.kill()
                    process.wait(timeout=5)
            for output in logs:
                output.close()


if __name__ == '__main__':
    main()
