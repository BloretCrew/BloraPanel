#!/usr/bin/env python3
"""Read-only Linux RSS/archive sampling for one explicitly selected devfixture.

Run alongside performance.spec.ts. Output is JSON Lines; no credentials or
process arguments are emitted. Samples alone do not establish test success.
"""
import argparse
import datetime
import json
import pathlib
import time


def sample(root):
    processes = []
    for process in pathlib.Path('/proc').iterdir():
        if not process.name.isdigit():
            continue
        try:
            args = (process / 'cmdline').read_bytes().split(b'\0')
            service = pathlib.Path(args[0].decode()).name
            if service not in ('blora-master', 'blora-daemon'):
                continue
            prefix = str(root).encode() + b'/'
            if not any(arg.startswith(prefix) or b'=' + prefix in arg for arg in args):
                continue
            status = (process / 'status').read_text().splitlines()
            rss = next(int(line.split()[1]) * 1024 for line in status if line.startswith('VmRSS:'))
            processes.append({'pid': int(process.name), 'service': service, 'rssBytes': rss})
        except (OSError, UnicodeError, StopIteration, ValueError):
            continue
    archives = []
    for directory in sorted(root.glob('*/terminals/*')):
        if not directory.is_dir():
            continue
        # Rotation can retire a segment during this read. Retry rather than
        # treating a disappearing file as a zero-sized successful sample.
        for _ in range(3):
            segments = sorted(directory.glob('*.seg'))
            try:
                size = sum(segment.stat().st_size for segment in segments)
            except FileNotFoundError:
                continue
            archives.append({'directory': str(directory.relative_to(root)),
                             'segments': len(segments), 'bytes': size,
                             'first': segments[0].name if segments else None})
            break
        else:
            archives.append({'directory': str(directory.relative_to(root)), 'error': 'rotation_race'})
    return {'utc': datetime.datetime.now(datetime.timezone.utc).isoformat(),
            'processes': processes, 'archives': archives}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('fixture', type=pathlib.Path)
    parser.add_argument('--samples', type=int, default=1)
    parser.add_argument('--interval', type=float, default=60)
    args = parser.parse_args()
    root = args.fixture.resolve(strict=True)
    if not root.name.startswith('fixture-') or root.parent.name != '.local':
        parser.error('select an existing .local/fixture-* directory')
    if not 1 <= args.samples <= 120 or not 1 <= args.interval <= 3600:
        parser.error('samples must be 1..120 and interval 1..3600 seconds')
    for index in range(args.samples):
        if index:
            time.sleep(args.interval)
        print(json.dumps(sample(root)), flush=True)


if __name__ == '__main__':
    main()
