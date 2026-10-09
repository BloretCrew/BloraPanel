#!/usr/bin/env python3
"""Read-only verification of component manifests, source and matching Web copies."""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import subprocess
import tarfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]


def digest(value):
    return hashlib.sha256(value).hexdigest()


def payload(path):
    values, modes = {}, {}
    if path.suffix == '.zip':
        with zipfile.ZipFile(path) as archive:
            for item in archive.infolist():
                assert not item.is_dir(), 'Unexpected directory entry'
                assert item.filename not in values, 'Duplicate ZIP member'
                assert item.external_attr >> 16 & 0o170000 == 0o100000, 'Nonregular ZIP member'
                values[item.filename] = archive.read(item)
                modes[item.filename] = item.external_attr >> 16 & 0o777
    else:
        with tarfile.open(path) as archive:
            for item in archive:
                assert item.isfile(), 'Nonregular TAR member'
                assert item.name not in values, 'Duplicate TAR member'
                values[item.name] = archive.extractfile(item).read()
                modes[item.name] = item.mode
    for name in values:
        p = PurePosixPath(name)
        assert not p.is_absolute() and '..' not in p.parts, 'Unsafe archive path'
        assert not any(part in {'.local', '.git', 'node_modules', '.codex', '.agents'} for part in p.parts), 'Private archive path'
    return values, modes


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('release', type=Path)
    parser.add_argument('--version', required=True)
    parser.add_argument('--revision', required=True)
    parser.add_argument('--source-archive', type=Path)
    args = parser.parse_args()
    assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip() == args.revision
    assert not subprocess.check_output(['git', 'status', '--porcelain'], cwd=ROOT, text=True).strip(), 'Source tree is modified'
    tracked_inputs = set(subprocess.check_output(['git', 'ls-files', '-z'], cwd=ROOT).decode().strip('\0').split('\0'))
    expected = {}
    for platform in ('linux-amd64', 'windows-amd64'):
        for component in ('master', 'daemon'):
            suffix = '.zip' if platform.startswith('windows') else '.tar.gz'
            expected[f'blora-{component}-{args.version}-{platform}{suffix}'] = (component, platform)
    expected[f'blora-web-{args.version}.tar.gz'] = ('web', 'browser')
    expected[f'blora-sdk-{args.version}.tar.gz'] = ('sdk', 'cross-platform')
    sums = {}
    for line in (args.release / 'SHA256SUMS').read_text().splitlines():
        sha, name = line.split('  ', 1)
        assert name not in sums, 'Duplicate external checksum'
        sums[name] = sha
    assert set(expected) <= set(sums), 'Missing component checksum'
    core_metadata = None
    if (args.release / 'CORE-UPDATE.json').exists():
        core_payload = (args.release / 'CORE-UPDATE.json').read_bytes()
        assert digest(core_payload) == sums.get('CORE-UPDATE.json'), 'Core update metadata checksum mismatch'
        core_metadata = json.loads(core_payload)
        assert core_metadata['version'] == args.version and core_metadata['revision'] == args.revision
        probe = json.loads(subprocess.check_output([str(ROOT / 'dist/blora-master'), '--core-update-info']))
        for key in ('formatVersion', 'version', 'revision', 'protocolVersion', 'schemaVersion', 'schemaFingerprint', 'preserveInstances'):
            assert core_metadata[key] == probe[key], 'Core update metadata differs from binary: ' + key
        assert probe['managedRestart'] and core_metadata['peerProtocolMin'] <= probe['protocolVersion'] <= core_metadata['peerProtocolMax']
    web_copies, records, count = [], [], 0
    for name, (component, platform) in expected.items():
        archive = args.release / name
        sha = digest(archive.read_bytes())
        assert sha == sums[name], 'External checksum mismatch: ' + name
        values, modes = payload(archive)
        assert all(p in tracked_inputs for p in values if p.startswith('docs/')), 'Untracked/private documentation in release archive'
        manifest = json.loads(values.pop('MANIFEST.json'))
        assert (manifest['version'], manifest['component'], manifest['platform']) == (args.version, component, platform)
        entries = manifest['files']
        assert len({e['path'] for e in entries}) == len(entries)
        assert set(values) == {e['path'] for e in entries}, 'Manifest membership mismatch'
        for entry in entries:
            path, value = entry['path'], values[entry['path']]
            assert len(value) == entry['size'] and digest(value) == entry['sha256'], 'Internal checksum mismatch'
            assert modes[path] == int(entry['mode'], 8), 'File mode mismatch'
            source = ROOT / path
            if path.startswith('web/dist/'):
                source = ROOT / path
            elif component == 'web' and path not in {'LICENSE', 'SOURCE-REVISION.txt', 'THIRD-PARTY-NOTICES.txt'}:
                source = ROOT / 'web/dist' / path
            elif path.startswith('blora-') or path.startswith('tools/blora-'):
                source = ROOT / 'dist' / Path(path).name
            elif path == 'THIRD-PARTY-NOTICES.txt':
                source = ROOT / 'docs/licenses' / path
            elif path == 'CORE-UPDATE.json':
                assert core_metadata is not None and value == core_payload, 'Missing or differing external core metadata'
                source = args.release / path
            if path not in {'START.txt', 'SOURCE-REVISION.txt'}:
                assert source.is_file() and source.read_bytes() == value, 'Current input mismatch: ' + path
        assert values['LICENSE'] == (ROOT / 'LICENSE').read_bytes()
        assert (f'Source revision: {args.revision}\n').encode() in values['SOURCE-REVISION.txt']
        assert b'Modified working tree: false\n' in values['SOURCE-REVISION.txt']
        if component in {'master', 'daemon'}:
            assert 'README.zh-CN.md' in values and 'docs/licenses/THIRD-PARTY-NOTICES.txt' in values
            if core_metadata is not None:
                assert values.get('CORE-UPDATE.json') == core_payload, 'Core package metadata mismatch'
        else:
            assert 'THIRD-PARTY-NOTICES.txt' in values
        if component == 'master':
            web_copies.append({p.removeprefix('web/dist/'): digest(v) for p, v in values.items() if p.startswith('web/dist/')})
        elif component == 'web':
            web_copies.append({p: digest(v) for p, v in values.items() if p not in {'LICENSE', 'SOURCE-REVISION.txt', 'THIRD-PARTY-NOTICES.txt'}})
        count += len(entries)
        records.append({'name': name, 'sha256': sha, 'files': len(entries)})
    actual_web = {p.relative_to(ROOT / 'web/dist').as_posix(): digest(p.read_bytes()) for p in (ROOT / 'web/dist').rglob('*') if p.is_file()}
    assert len(web_copies) == 3 and all(copy == actual_web for copy in web_copies), 'Web copies differ'
    source_count = None
    if args.source_archive:
        with tarfile.open(args.source_archive) as archive:
            members = [m for m in archive if m.isfile()]
            relative = {m.name.split('/', 1)[1]: m for m in members}
            tracked = set(subprocess.check_output(['git', 'ls-files', '-z'], cwd=ROOT).decode().strip('\0').split('\0'))
            assert set(relative) == tracked, 'Source archive does not contain all tracked inputs'
            for path, member in relative.items():
                assert archive.extractfile(member).read() == (ROOT / path).read_bytes(), 'Source archive differs: ' + path
            source_count = len(relative)
    print(json.dumps({'version': args.version, 'revision': args.revision, 'componentArchives': records,
                      'manifestFiles': count, 'identicalWebCopies': 3, 'webFiles': len(actual_web),
                      'sourceFiles': source_count, 'license': 'GPL-3.0-only', 'verified': True}, indent=2))


if __name__ == '__main__':
    main()
