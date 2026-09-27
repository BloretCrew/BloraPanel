#!/usr/bin/env python3
"""Collect upstream notice texts from pinned local release inputs, offline."""
import hashlib
import json
import os
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[1]


def notice_files(directory):
    if directory.is_symlink():
        raise ValueError(f'Symlink dependency directory is not a notice input: {directory.name}')
    files = sorted(path for path in directory.iterdir() if path.is_file() and
                   any(word in path.name.lower() for word in ('license', 'notice', 'copying', 'copyright')))
    if any(path.is_symlink() for path in files):
        raise ValueError(f'Symlink notice is not a release input: {directory.name}')
    return files


def main():
    sections = []
    inventory = []

    def append(name, version, origin, files, declared=None):
        if not files:
            raise ValueError(f'Missing upstream license text: {name}@{version}')
        record = {'name': name, 'version': version, 'origin': origin, 'files': []}
        if declared:
            record['declaredLicense'] = declared
        text = [f'\n{"=" * 72}\n{name} @ {version}\nSource: {origin}\n']
        if declared:
            text.append(f'Package-declared license: {declared}\n')
        for path in files:
            if path.is_symlink() or not path.is_file():
                raise ValueError(f'Notice must be a regular file: {path.name}')
            data = path.read_bytes()
            record['files'].append({'name': path.name, 'sha256': hashlib.sha256(data).hexdigest()})
            text.append(f'\n--- {path.name} ---\n{data.decode("utf-8")}\n')
        inventory.append(record)
        sections.append(''.join(text))

    modules = set()
    for component in ('master', 'daemon', 'extension-sign'):
        for suffix in ('', '.exe'):
            binary = ROOT / 'dist' / f'blora-{component}{suffix}'
            output = subprocess.check_output(['go', 'version', '-m', str(binary)], text=True)
            for line in output.splitlines():
                fields = line.split()
                if fields and fields[0] == 'dep':
                    modules.add((fields[1], fields[2]))
                if fields and fields[0] == '=>':
                    raise ValueError('Resolve replaced Go modules explicitly before packaging notices')
    cache = Path(subprocess.check_output(['go', 'env', 'GOMODCACHE'], text=True).strip())
    for name, version in sorted(modules):
        escaped = ''.join('!' + char.lower() if char.isupper() else char for char in name)
        directory = cache / f'{escaped}@{version}'
        append(name, version, f'https://pkg.go.dev/{name}@{version}', notice_files(directory))
    go_root = Path(subprocess.check_output(['go', 'env', 'GOROOT'], text=True).strip())
    go_version = subprocess.check_output(['go', 'env', 'GOVERSION'], text=True).split()[0]
    license_root = Path(os.environ.get('BLORA_GO_LICENSE_DIR', str(go_root)))
    if not (license_root / 'LICENSE').is_file() and 'BLORA_GO_LICENSE_DIR' not in os.environ:
        # Fedora splits these files into the installed package license tree.
        license_root = Path('/usr/share/licenses/golang')
    append('Go toolchain and standard library', go_version,
           f'https://go.googlesource.com/go/+/refs/tags/{go_version}/',
           [license_root / name for name in ('LICENSE', 'PATENTS') if (license_root / name).is_file()])

    lock = json.loads((ROOT / 'web/package-lock.json').read_text())
    skipped = []
    for relative, metadata in sorted(lock['packages'].items()):
        if not relative or metadata.get('dev'):
            continue
        directory = ROOT / 'web' / relative
        if not directory.is_dir():
            if not metadata.get('optional'):
                raise ValueError(f'Run npm ci: missing {relative}')
            skipped.append(relative)
            continue
        package = json.loads((directory / 'package.json').read_text())
        name, version = package['name'], package['version']
        if version != metadata['version']:
            raise ValueError(f'Installed package differs from lock: {name}')
        files = notice_files(directory)
        origin = metadata.get('resolved', f'https://www.npmjs.com/package/{name}/v/{version}')
        if not files and (name, version) in {
                ('@xterm/addon-serialize', '0.14.0'), ('@xterm/headless', '6.0.0')}:
            # These tarballs omit LICENSE. Use the companion's upstream text
            # only when both packages match the exact verified source commit.
            companion = ROOT / 'web/node_modules/@xterm/xterm'
            companion_package = json.loads((companion / 'package.json').read_text())
            commit = 'f447274f430fd22513f6adbf9862d19524471c04'
            if (package.get('commit') != commit or package.get('license') != 'MIT' or
                    companion_package.get('version') != '6.0.0' or
                    companion_package.get('commit') != commit or
                    companion_package.get('license') != 'MIT'):
                raise ValueError(f'Unverified companion license source: {name}@{version}')
            license_file = companion / 'LICENSE'
            expected_hash = 'b569f629d00f2626a8100df2a1798210535621e42164dfd426a6fe5aac7b0ccd'
            if license_file.is_symlink() or hashlib.sha256(license_file.read_bytes()).hexdigest() != expected_hash:
                raise ValueError(f'Companion license text changed: {name}@{version}')
            files = [license_file]
            origin = f'https://github.com/xtermjs/xterm.js/blob/{commit}/LICENSE'
        elif not files and name == '@vue/devtools-api' and version == '6.6.4':
            files = [ROOT / 'docs/licenses/vue-devtools-6.6.4-LICENSE.txt']
            origin = 'https://github.com/vuejs/vue-devtools/blob/v6.6.4/LICENSE'
        elif not files and name.startswith('@rolldown/binding-'):
            parent = ROOT / 'web/node_modules/rolldown'
            if json.loads((parent / 'package.json').read_text())['version'] != version:
                raise ValueError('Rolldown native binding and parent versions differ')
            files = notice_files(parent)
            origin += ' (shared license and third-party notices from matching rolldown package)'
        append(name, version, origin, files, metadata.get('license'))

    header = ('Blora Panel — Third-party license texts and notices\n\n'
              'Generated by scripts/third-party-notices.py from compiled Go release dependencies,\n'
              'the Go toolchain, and installed non-dev entries of web/package-lock.json.\n'
              'The latter also includes build tools declared as dependencies; their listing\n'
              'does not imply that every listed package is embedded in the browser bundle.\n'
              'Original upstream notices below are retained without changing their terms.\n'
              'This file does not select a license for Blora Panel itself.\n')
    destination = ROOT / 'docs/licenses'
    destination.mkdir(exist_ok=True)
    (destination / 'THIRD-PARTY-NOTICES.txt').write_text(header + ''.join(sections), encoding='utf-8')
    (destination / 'inventory.json').write_text(json.dumps({'formatVersion': 1, 'packages': inventory,
        'uninstalledOptionalPackages': skipped}, ensure_ascii=False, sort_keys=True, indent=2) + '\n', encoding='utf-8')
    print(f'Collected {len(inventory)} dependency notice records; {len(skipped)} uninstalled optional packages excluded')


if __name__ == '__main__':
    main()
