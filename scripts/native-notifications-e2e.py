#!/usr/bin/env python3
"""Verify real Linux notification rendering/clicks in a disposable X11/DBus session.

Requires a built dist/blora-devfixture, web/dist, web/node_modules and the image
built from scripts/native-notifications.Dockerfile. No host display/bus is used.
"""
import argparse
import json
import os
import pathlib
import signal
import subprocess
import tempfile
import time
import uuid


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--image', default='blora-native-notifications:e2e')
    parser.add_argument('--output', type=pathlib.Path, required=True)
    parser.add_argument('--listen', default='127.0.0.1:9444')
    args = parser.parse_args()
    if args.listen not in ('127.0.0.1:9443', '127.0.0.1:9444'):
        parser.error('fixture supports loopback ports 9443 or 9444 only')
    root = pathlib.Path(__file__).resolve().parent.parent
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False, mode=0o700)
    parent = pathlib.Path(tempfile.mkdtemp(prefix='blora-native-notifications-'))
    name = 'blora-native-notifications-' + uuid.uuid4().hex
    started = time.monotonic()
    process = None
    created = False
    with (output / 'fixture.log').open('w') as log:
        os.chmod(log.name, 0o600)
        try:
            process = subprocess.Popen([str(root / 'dist/blora-devfixture'), '--listen', args.listen,
                                        '--state-parent', str(parent)], cwd=root, stdout=log, stderr=log)
            deadline = time.monotonic() + 45
            while True:
                credentials = list(parent.glob('fixture-*/browser-credentials.json'))
                if credentials:
                    break
                if process.poll() is not None or time.monotonic() >= deadline:
                    raise RuntimeError('private fixture failed to become ready; inspect private log')
                time.sleep(0.25)
            # Host networking is needed only for the loopback fixture. No host
            # display, DBus, Docker socket or devices are mounted. The checkout
            # is read-only; only this run's private evidence directory is writable.
            subprocess.run([
                'docker', 'create', '--name', name, '--network', 'host', '--init',
                '--env', 'BLORA_NATIVE_NOTIFICATION_E2E=1',
                '--env', 'BLORA_E2E_CREDENTIALS=/probe/credentials.json',
                '--env', 'BLORA_NATIVE_OUTPUT=/evidence/results',
                '--mount', f'type=bind,src={root},dst=/workspace,readonly',
                '--mount', f'type=bind,src={credentials[0]},dst=/probe/credentials.json,readonly',
                '--mount', f'type=bind,src={output},dst=/evidence',
                args.image, 'xvfb-run', '-a', '-s', '-screen 0 1440x960x24',
                'dbus-run-session', '--', 'sh', '/workspace/scripts/native-notifications-session.sh',
            ], check=True, stdout=subprocess.DEVNULL, timeout=30)
            created = True
            subprocess.run(['docker', 'start', '--attach', name], check=True, timeout=120)
            status = subprocess.check_output(['docker', 'inspect', '--format', '{{.State.ExitCode}}', name],
                                             text=True, timeout=10).strip()
            if status != '0':
                raise RuntimeError(f'native notification container exited {status}')
            if not list(output.glob('results/**/native-popup.png')):
                raise RuntimeError('native desktop screenshot was not produced')
        finally:
            try:
                if created:
                    subprocess.run(['docker', 'rm', '--force', name], check=True,
                                   stdout=subprocess.DEVNULL, timeout=30)
            finally:
                if process is not None:
                    code = process.poll()
                    if code is None:
                        process.send_signal(signal.SIGINT)
                        try:
                            code = process.wait(timeout=30)
                        except subprocess.TimeoutExpired:
                            process.kill()
                            process.wait(timeout=10)
                            raise RuntimeError('private fixture failed graceful shutdown')
                    if code != 0:
                        raise RuntimeError(f'private fixture exited {code}')
    print(json.dumps({'status': 'PASSED', 'elapsedSeconds': time.monotonic() - started,
                      'cleanedUp': True, 'output': str(output)}))


if __name__ == '__main__':
    main()
