#!/usr/bin/env python3
"""Build deterministic local release archives from explicit project inputs."""
import argparse
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import re
import shutil
import stat
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]
IGNORED = {"node_modules", ".git", ".local", ".codex", ".agents", "__pycache__"}


def tree(relative):
    base = ROOT / relative
    if not base.is_dir() or base.is_symlink():
        raise ValueError(f"Missing regular input directory: {relative}")
    result = []
    for directory, directories, files in os.walk(base, followlinks=False):
        directories[:] = sorted(name for name in directories if name not in IGNORED)
        for name in directories:
            if (Path(directory) / name).is_symlink():
                raise ValueError("Symlink input directories are not packaged")
        result.extend(Path(directory, name).relative_to(ROOT).as_posix() for name in sorted(files))
    return result


def content(relative):
    source = ROOT / relative
    if source.is_symlink() or not source.is_file():
        raise ValueError(f"Missing regular input file: {relative}")
    if any(part in IGNORED for part in source.relative_to(ROOT).parts):
        raise ValueError("Private/runtime directories are not package inputs")
    before = source.stat()
    value = source.read_bytes()
    after = source.stat()
    if (before.st_ino, before.st_size, before.st_mtime_ns) != (after.st_ino, after.st_size, after.st_mtime_ns) or len(value) != before.st_size:
        raise ValueError(f"Input changed during packaging: {relative}")
    return value


def executable(relative, platform):
    output = subprocess.check_output(["go", "version", "-m", str(ROOT / relative)], text=True)
    settings = {}
    for line in output.splitlines():
        fields = line.split()
        if len(fields) == 2 and fields[0] == "build" and "=" in fields[1]:
            key, value = fields[1].split("=", 1)
            settings[key] = value
    target_os, target_arch = platform.split("-")
    if settings.get("GOOS") != target_os or settings.get("GOARCH") != target_arch:
        raise ValueError(f"Binary platform mismatch: {relative}")


def archive(destination, entries, version, component, platform):
    records = []
    payload = []
    for name, source, mode in sorted(entries):
        if name.startswith("/") or ".." in Path(name).parts:
            raise ValueError("Unsafe archive path")
        value = source if isinstance(source, bytes) else content(source)
        records.append({"path": name, "size": len(value), "mode": f"{mode:04o}", "sha256": hashlib.sha256(value).hexdigest()})
        payload.append((name, value, mode))
    if len({name for name, _, _ in payload}) != len(payload):
        raise ValueError("Duplicate archive path")
    manifest = {"formatVersion": 1, "version": version, "component": component, "platform": platform, "files": records}
    payload.append(("MANIFEST.json", (json.dumps(manifest, ensure_ascii=False, sort_keys=True, indent=2) + "\n").encode(), 0o644))
    if destination.suffix == ".zip":
        with zipfile.ZipFile(destination, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as target:
            for name, value, mode in payload:
                info = zipfile.ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
                info.compress_type = zipfile.ZIP_DEFLATED
                info.external_attr = (stat.S_IFREG | mode) << 16
                info.create_system = 3
                target.writestr(info, value)
    else:
        with destination.open("wb") as raw, gzip.GzipFile(filename="", mode="wb", fileobj=raw, mtime=0) as compressed:
            with tarfile.open(fileobj=compressed, mode="w", format=tarfile.PAX_FORMAT) as target:
                for name, value, mode in payload:
                    info = tarfile.TarInfo(name)
                    info.size, info.mode, info.mtime = len(value), mode, 0
                    info.uid = info.gid = 0
                    info.uname = info.gname = ""
                    target.addfile(info, io.BytesIO(value))


def run():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", default="development")
    parser.add_argument("--output", type=Path, default=ROOT / "dist/releases")
    options = parser.parse_args()
    if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]{0,63}", options.version):
        parser.error("version must be a simple 1..64 character release label")
    options.output.mkdir(parents=True, exist_ok=True)
    revision = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip()
    modified = bool(subprocess.check_output(["git", "status", "--porcelain", "--untracked-files=normal"], cwd=ROOT, text=True).strip())
    if "-beta." in options.version and modified:
        raise ValueError("Beta packages require a clean committed source tree")
    source_notice = (f"Blora Panel {options.version}\n"
                     f"Source revision: {revision}\n"
                     f"Modified working tree: {str(modified).lower()}\n"
                     f"Source: https://github.com/BloretCrew/BloraPanel/tree/{revision}\n"
                     "License: GPL-3.0-only; see LICENSE. Third-party terms remain separate.\n").encode()
    final = options.output / options.version
    staging = Path(tempfile.mkdtemp(prefix=".blora-package-", dir=options.output))
    try:
        documentation = ["LICENSE", "README.md", "README.zh-CN.md", "master.example.json", "daemon.example.json", "sdk/README.md"] + tree("docs")
        common = [(name, name, 0o644) for name in documentation] + [("SOURCE-REVISION.txt", source_notice, 0o644)]
        web = [(name, name, 0o644) for name in tree("web/dist")]
        image = [(name, name, 0o644) for name in ["Dockerfile.isolated", "go.mod", "go.sum"] + tree("cmd/exec-helper") + tree("internal/containerterm/helper")]
        for platform in ("linux-amd64", "windows-amd64"):
            extension = ".exe" if platform.startswith("windows") else ""
            suffix = ".zip" if extension else ".tar.gz"
            for component in ("master", "daemon"):
                binary = f"blora-{component}{extension}"
                executable(f"dist/{binary}", platform)
                invocation = f".\\{binary}" if extension else f"./{binary}"
                command = invocation
                setup = ("For Master, copy master.example.json to master.json beside the executable, edit paths and create its private passwordFile.\n"
                         "The first normal start initializes the administrator automatically.\n") if component == "master" else (
                         "For Daemon, copy daemon.example.json to daemon.json beside the executable; configure its one-time enrollmentFile.\n")
                start = (f"Blora {component} / {platform} / {options.version}\n\n"
                         "This is a prerelease/development build, not a stable-release or full platform acceptance claim.\n"
                         "Blora Panel is licensed GPL-3.0-only; see LICENSE and the third-party notices.\n"
                         "Use platform-appropriate private paths and restrictive state/config permissions.\n"
                         + setup + f"Start without arguments: {command}\n"
                         "Nginx owns public HTTPS; the default backend is loopback HTTP at 127.0.0.1:37861.\n"
                         "See docs/operations/OPERATIONS.md for identities, proxy configuration, backups and upgrades.\n"
                         "Windows binaries are console programs, not native SCM service executables.\n"
                         "MANIFEST.json records every payload file; verify the external SHA256SUMS before use.\n").encode()
                entries = common + [(binary, f"dist/{binary}", 0o755), ("START.txt", start, 0o644)] + (web if component == "master" else image)
                archive(staging / f"blora-{component}-{options.version}-{platform}{suffix}", entries, options.version, component, platform)
        notices = ("THIRD-PARTY-NOTICES.txt", "docs/licenses/THIRD-PARTY-NOTICES.txt", 0o644)
        archive(staging / f"blora-web-{options.version}.tar.gz", [(name.removeprefix("web/dist/"), source, mode) for name, source, mode in web] + [notices, ("LICENSE", "LICENSE", 0o644), ("SOURCE-REVISION.txt", source_notice, 0o644)], options.version, "web", "browser")
        sdk_sources = tree("sdk")
        sdk = [(name, name, 0o644) for name in sdk_sources] + [notices, ("LICENSE", "LICENSE", 0o644), ("SOURCE-REVISION.txt", source_notice, 0o644)]
        for platform, extension in (("linux-amd64", ""), ("windows-amd64", ".exe")):
            source = f"dist/blora-extension-sign{extension}"
            executable(source, platform)
            sdk.append((f"tools/blora-extension-sign{extension}", source, 0o755))
        sdk.append(("START.txt", b"Run npm ci and npm run build in sdk, then npm ci and npm run package in sdk/examples/reference-app.\nUse tools/blora-extension-sign with your private publisher key to sign the resulting package.\nNo credentials, node_modules or runtime state are included.\n", 0o644))
        archive(staging / f"blora-sdk-{options.version}.tar.gz", sdk, options.version, "sdk", "cross-platform")
        hashes = "".join(f"{hashlib.sha256(path.read_bytes()).hexdigest()}  {path.name}\n" for path in sorted(staging.iterdir()))
        (staging / "SHA256SUMS").write_text(hashes, encoding="utf-8")
        if final.is_symlink():
            raise ValueError("Release destination must not be a symlink")
        if final.exists():
            expected = {path.name for path in staging.iterdir()}
            if not final.is_dir() or {path.name for path in final.iterdir()} != expected or any(not (final / path.name).is_file() or (final / path.name).is_symlink() or (final / path.name).read_bytes() != path.read_bytes() for path in staging.iterdir()):
                raise ValueError("Release label already exists with different content; choose a new version or output directory")
            print(f"Identical release already exists: {final}")
        else:
            staging.rename(final)
            print(f"Release ready: {final}")
    finally:
        if staging.exists():
            shutil.rmtree(staging)


if __name__ == "__main__":
    run()
