#!/usr/bin/env python3
"""Linux real-binary online-update drill against a private HTTPS Release fixture.

Uses only task-owned processes/state. This is mechanism evidence, not a public
Release, production deployment or Windows runtime certification.
"""
import argparse
import hashlib
import http.cookiejar
import http.server
import io
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
import threading
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parents[1]
OLD_REVISION, NEW_REVISION = "a" * 40, "b" * 40
OLD_VERSION, NEW_VERSION = "0.1.0-beta.1", "0.1.0-beta.2"


def encode(value):
    return json.dumps(value, sort_keys=True).encode()


def wait_for(check, description, seconds=50):
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        try:
            value = check()
            if value:
                return value
        except (OSError, urllib.error.URLError):
            pass
        time.sleep(.1)
    raise RuntimeError("Timed out: " + description)


def bundle(binary, component, metadata):
    files = {
        "blora-" + component: (binary.read_bytes(), 0o755),
        "CORE-UPDATE.json": (encode(metadata), 0o644),
        "SOURCE-REVISION.txt": (("Source revision: " + NEW_REVISION +
                                 "\nModified working tree: false\n").encode(), 0o644),
    }
    if component == "master":
        files["web/dist/index.html"] = (b"<html>UPDATED-RELEASE-FIXTURE</html>", 0o644)
    manifest = {"formatVersion": 1, "version": NEW_VERSION, "component": component,
                "platform": "linux-amd64", "files": [
                    {"path": name, "size": len(data), "mode": f"{mode:04o}",
                     "sha256": hashlib.sha256(data).hexdigest()}
                    for name, (data, mode) in files.items()]}
    files["MANIFEST.json"] = (encode(manifest), 0o644)
    output = io.BytesIO()
    with tarfile.open(fileobj=output, mode="w:gz") as archive:
        for name, (data, mode) in files.items():
            item = tarfile.TarInfo(name)
            item.size, item.mode = len(data), mode
            archive.addfile(item, io.BytesIO(data))
    return output.getvalue()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--report", type=Path)
    args = parser.parse_args()
    if os.name != "posix" or not Path("/proc/self/stat").exists():
        parser.error("This drill requires Linux; use Windows-specific device checks there")
    checks, processes, logs = [], [], []
    server = None
    instance_id, call, pid, birth = None, None, None, None
    work = Path(tempfile.mkdtemp(prefix="blora-core-update-"))
    os.chmod(work, 0o700)
    try:
        binaries = {}
        for version, revision, label in ((OLD_VERSION, OLD_REVISION, "old"),
                                         (NEW_VERSION, NEW_REVISION, "new")):
            for component in ("master", "daemon"):
                binary = work / f"{label}-{component}"
                subprocess.run(["go", "build", "-ldflags",
                                f"-X blora.dev/panel/internal/coreupdate.Version={version} "
                                f"-X blora.dev/panel/internal/coreupdate.Revision={revision}",
                                "-o", str(binary), "./cmd/" + component], cwd=ROOT, check=True)
                binaries[label, component] = binary
        info = json.loads(subprocess.check_output([str(binaries["new", "master"]), "--core-update-info"]))
        metadata = {key: info[key] for key in ("formatVersion", "version", "revision",
                    "protocolVersion", "schemaVersion", "schemaFingerprint", "preserveInstances")}
        metadata.update(peerProtocolMin=1, peerProtocolMax=1)
        assets = {"CORE-UPDATE.json": encode(metadata)}
        for component in ("master", "daemon"):
            name = f"blora-{component}-{NEW_VERSION}-linux-amd64.tar.gz"
            assets[name] = bundle(binaries["new", component], component, metadata)
        assets["SHA256SUMS"] = "".join(hashlib.sha256(data).hexdigest() + "  " + name + "\n"
                                       for name, data in assets.items()).encode()
        fixture = {"version": NEW_VERSION, "revision": NEW_REVISION, "corrupt": False}

        class Releases(http.server.BaseHTTPRequestHandler):
            def do_GET(self):
                path = urllib.parse.urlparse(self.path).path
                if path == "/api/releases":
                    data = encode([{"id": 1, "tag_name": "v" + fixture["version"],
                                    "html_url": endpoint + "/release", "draft": False,
                                    "prerelease": True, "assets": [
                                        {"name": name, "size": len(value),
                                         "browser_download_url": endpoint + "/assets/" + name}
                                        for name, value in assets.items()]}])
                elif path.startswith("/api/git/ref/tags/"):
                    data = encode({"object": {"type": "commit", "sha": fixture["revision"]}})
                elif path.startswith("/assets/") and path[8:] in assets:
                    data = assets[path[8:]]
                    if fixture["corrupt"] and path.endswith(".tar.gz"):
                        data = bytes([data[0] ^ 1]) + data[1:]
                else:
                    self.send_error(404)
                    return
                self.send_response(200)
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)

            def log_message(self, *_):
                pass

        cert, key = work / "fixture.crt", work / "fixture.key"
        subprocess.run(["openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes",
                        "-days", "1", "-subj", "/CN=127.0.0.1",
                        "-addext", "subjectAltName=IP:127.0.0.1", "-keyout", str(key),
                        "-out", str(cert)], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Releases)
        tls = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        tls.load_cert_chain(cert, key)
        server.socket = tls.wrap_socket(server.socket, server_side=True)
        endpoint = f"https://127.0.0.1:{server.server_port}"
        threading.Thread(target=server.serve_forever, daemon=True).start()
        with socket.socket() as reserve:
            reserve.bind(("127.0.0.1", 0))
            port = reserve.getsockname()[1]
        origin = f"http://127.0.0.1:{port}"
        source = {"repository": "https://github.com/BloretCrew/BloraPanel", "channel": "beta",
                  "apiUrl": endpoint + "/api"}
        password = secrets.token_urlsafe(32)
        for component in ("master", "daemon"):
            install = work / component
            install.mkdir(mode=0o700)
            shutil.copy2(binaries["old", component], install / ("blora-" + component))
        master, daemon = work / "master", work / "daemon"
        (master / "web/dist").mkdir(parents=True)
        (master / "web/dist/index.html").write_text("<html>INITIAL-FIXTURE</html>")
        (master / "initial-password").write_text(password)
        (master / "initial-password").chmod(0o600)
        (master / "master.json").write_text(json.dumps({"listen": f"127.0.0.1:{port}", "updates": source}))
        environment = dict(os.environ, SSL_CERT_FILE=str(cert))

        def start(component):
            output = (work / (component + ".log")).open("wb")
            logs.append(output)
            process = subprocess.Popen([str(work / component / ("blora-" + component))],
                                       cwd="/tmp", env=environment, stdout=output, stderr=output)
            processes.append(process)
            return process

        csrf = ""
        client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))

        def api(path, value=None, method=None, request_id=None):
            request = urllib.request.Request(origin + path, method=method,
                        data=None if value is None else encode(value), headers={"Origin": origin,
                        "Content-Type": "application/json", "X-CSRF-Token": csrf,
                        "Idempotency-Key": request_id or str(uuid.uuid4())})
            with client.open(request, timeout=5) as response:
                data = response.read()
                return json.loads(data) if data.startswith((b"{", b"[")) else data

        call = api
        master_process = start("master")
        wait_for(lambda: api("/healthz"), "Master startup")
        csrf = api("/api/v1/login", {"name": "admin", "password": password})["csrfToken"]
        token = api("/api/v1/nodes/enrollments", {"name": "online-update-fixture"})["token"]
        (daemon / "node.enrollment").write_text(token)
        (daemon / "node.enrollment").chmod(0o600)
        (daemon / "daemon.json").write_text(json.dumps({"masterUrl": origin, "allowPGIDFallback": True,
                                                       "enrollmentFile": "node.enrollment", "updates": source}))
        daemon_process = start("daemon")

        def online_node():
            nodes = api("/api/v1/nodes")["items"]
            return next((n for n in nodes if n["state"] == "ONLINE"), None)

        node = wait_for(online_node, "Daemon enrollment")
        directory = work / "owned-instance"
        directory.mkdir(mode=0o700)
        instance_id = api("/api/v1/instances", {"nodeId": node["nodeId"], "name": "update-survivor",
            "config": {"mode": "native", "directory": str(directory), "command": ["/bin/sh", "-c",
                "echo $$ > instance.pid; while :; do printf tick >> heartbeat; sleep 0.05; done"],
                "stopSeconds": 2, "killSeconds": 2, "escalate": True}})["instance"]["instanceId"]

        def task(action):
            result = api(f"/api/v1/instances/{instance_id}/actions", {"action": action})["task"]
            def finished():
                current = api("/api/v1/tasks/" + result["taskId"])["task"]
                if current["state"] in ("FAILED", "INTERRUPTED", "CANCELLED"):
                    raise RuntimeError("Instance action failed: " + current["state"])
                return current if current["state"] == "SUCCEEDED" else None
            return wait_for(finished, "instance " + action)

        task("start")
        wait_for(lambda: (directory / "instance.pid").exists(), "owned instance pid")
        pid = int((directory / "instance.pid").read_text())
        birth = Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()[19]
        run_id = api(f"/api/v1/instances/{instance_id}")["instance"]["runId"]

        def assert_survivor():
            assert (directory / "instance.pid").read_text().strip() == str(pid)
            assert Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()[19] == birth
            assert api(f"/api/v1/instances/{instance_id}")["instance"]["runId"] == run_id
            size = (directory / "heartbeat").stat().st_size
            wait_for(lambda: (directory / "heartbeat").stat().st_size > size, "continued business heartbeat")

        for component in ("master", "daemon"):
            base = "/api/v1/system/updates" if component == "master" else f"/api/v1/nodes/{node['nodeId']}/updates"
            status_path = base if component == "master" else base + "/status"
            api(base + "/check", {})
            preview = wait_for(lambda: api(status_path).get("preview"), component + " preview")
            assert preview["compatible"] and preview["revision"] == NEW_REVISION, preview
            request_id = str(uuid.uuid4())
            api(base + "/apply", {"revision": NEW_REVISION}, request_id=request_id)
            def updated():
                status = api(status_path)
                job = status.get("job", {})
                if job.get("state") in ("failed", "interrupted"):
                    raise RuntimeError("Update failed: " + job.get("detail", ""))
                return status if status["version"] == NEW_VERSION and job.get("state") == "completed" else None
            wait_for(updated, component + " online activation", 90)
            wait_for(online_node, "node reconnection")
            assert_survivor()
            replay = api(base + "/apply", {"revision": NEW_REVISION}, request_id=request_id)
            assert replay.get("job", replay)["state"] == "completed", replay
            checks.append(component + "-real-release-switch-instance-survives-and-receipt-replays")
            print("PASS:", checks[-1], flush=True)
        assert master_process.poll() is None and daemon_process.poll() is None
        assert b"UPDATED-RELEASE-FIXTURE" in api("/")
        fixture.update(version="0.1.0-beta.3", revision="c" * 40, corrupt=True)
        bad_metadata = dict(metadata, version=fixture["version"], revision=fixture["revision"])
        original = assets[f"blora-master-{NEW_VERSION}-linux-amd64.tar.gz"]
        assets.clear()
        assets["CORE-UPDATE.json"] = encode(bad_metadata)
        assets[f"blora-master-{fixture['version']}-linux-amd64.tar.gz"] = original
        assets["SHA256SUMS"] = "".join(hashlib.sha256(data).hexdigest() + "  " + name + "\n"
                                       for name, data in assets.items()).encode()
        base = "/api/v1/system/updates"
        api(base + "/check", {})
        wait_for(lambda: (p := api(base).get("preview")) and p["revision"] == fixture["revision"], "tamper preview")
        api(base + "/apply", {"revision": fixture["revision"]})
        failed = wait_for(lambda: (s := api(base))["job"]["state"] == "failed" and s, "tampered download rejection")
        assert failed["version"] == NEW_VERSION and "checksum" in failed["job"]["detail"].lower(), failed
        assert_survivor()
        checks.append("tampered-release-refused-old-service-and-instance-retained")
        print("PASS:", checks[-1], flush=True)
        task("stop")
        instance_id = None
        if args.report:
            args.report.parent.mkdir(parents=True, exist_ok=True)
            args.report.write_text(json.dumps({"outcome": "PASS", "platform": "linux/amd64",
                "checks": checks, "scope": "Private HTTPS fixture, real Master/Daemon and native process; no public release or Windows runtime claim"}, indent=2) + "\n")
    finally:
        if instance_id and call:
            try:
                cleanup_task = call(f"/api/v1/instances/{instance_id}/actions", {"action": "kill"})["task"]
                wait_for(lambda: call("/api/v1/tasks/" + cleanup_task["taskId"])["task"]["state"]
                         in ("SUCCEEDED", "FAILED", "INTERRUPTED", "CANCELLED"), "owned cleanup", 10)
            except OSError:
                pass
            # If the test management service is unavailable, clean only the
            # exact recorded, task-owned process group after a birth check.
            if pid and birth:
                try:
                    current = Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()[19]
                    if current == birth and os.getpgid(pid) == pid:
                        os.killpg(pid, signal.SIGKILL)
                except (ProcessLookupError, FileNotFoundError):
                    pass
        for process in reversed(processes):
            if process.poll() is None:
                process.send_signal(signal.SIGINT)
        for process in reversed(processes):
            try:
                process.wait(timeout=35)
            except subprocess.TimeoutExpired:
                process.kill()  # Exact task-owned launcher only.
                process.wait()
        if server:
            server.shutdown()
            server.server_close()
        for output in logs:
            output.close()
        shutil.rmtree(work)


if __name__ == "__main__":
    main()
