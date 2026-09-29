#!/usr/bin/env python3
"""Run only task-owned systemd units in a disposable, private Docker namespace.

Requires Docker, an existing image containing systemd and a compiled Linux
systeminfo test executable. Never mounts the host bus, cgroup tree or repo.
"""
import argparse
import json
import pathlib
import subprocess
import time
import uuid


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=pathlib.Path, required=True)
    parser.add_argument("--runtime-binary", type=pathlib.Path,
                        help="also run the real delegated-cgroup runtime test")
    parser.add_argument("--image", default="mcr.microsoft.com/playwright:v1.63.0-noble")
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    runtime_binary = args.runtime_binary.resolve(strict=True) if args.runtime_binary else None
    runtime_mount = (["--mount", f"type=bind,src={runtime_binary},dst=/probe/runtime.test,readonly"]
                     if runtime_binary else [])
    init = pathlib.Path(__file__).with_name("systemd-e2e-init.sh").resolve(strict=True)
    name = "blora-systemd-e2e-" + uuid.uuid4().hex
    created = False
    started = time.monotonic()

    def run_test(path, target, environment=()):
        command = ["docker", "exec", *environment, name, path]
        listed = subprocess.check_output(command + [f"-test.list=^{target}$"], text=True, timeout=10)
        if target not in listed.splitlines():
            raise RuntimeError(f"test binary does not contain {target}")
        result = subprocess.run(command + [f"-test.run=^{target}$", "-test.v", "-test.timeout=90s"],
                                capture_output=True, text=True, timeout=100)
        print(result.stdout, end="", flush=True)
        print(result.stderr, end="", flush=True)
        result.check_returncode()
        if f"--- PASS: {target} " not in result.stdout or "--- SKIP:" in result.stdout:
            raise RuntimeError(f"{target} did not actually pass; skips are not acceptance")

    try:
        # SYS_ADMIN is confined to the container's own mount/PID/cgroup/network
        # namespaces. No privileged mode, host devices, sockets or mounts.
        subprocess.run([
            "docker", "create", "--name", name, "--network", "none", "--cgroupns", "private",
            "--cap-add", "SYS_ADMIN", "--security-opt", "seccomp=unconfined",
            "--tmpfs", "/run", "--tmpfs", "/run/lock", "--tmpfs", "/tmp",
            "--env", "container=docker", "--env", "BLORA_SYSTEMD_E2E=1",
            "--mount", f"type=bind,src={binary},dst=/probe/systeminfo.test,readonly",
            "--mount", f"type=bind,src={init},dst=/probe/init.sh,readonly",
            *runtime_mount,
            "--entrypoint", "/bin/sh", args.image, "/probe/init.sh",
        ], check=True, stdout=subprocess.DEVNULL, timeout=30)
        created = True
        subprocess.run(["docker", "start", name], check=True, stdout=subprocess.DEVNULL, timeout=30)
        deadline = time.monotonic() + 30
        while True:
            state = json.loads(subprocess.check_output([
                "docker", "inspect", "--format", "{{json .State}}", name
            ], timeout=10))
            if not state["Running"]:
                subprocess.run(["docker", "logs", name], timeout=10, check=False)
                raise RuntimeError(f"isolated systemd exited: {state['ExitCode']}")
            ready = subprocess.run([
                "docker", "exec", name, "systemctl", "show", "--property=SystemState", "--value"
            ], capture_output=True, text=True, timeout=10)
            if ready.returncode == 0 and ready.stdout.strip() in ("running", "degraded"):
                break
            if time.monotonic() >= deadline:
                raise RuntimeError("isolated systemd did not become ready in 30 seconds")
            time.sleep(0.5)
        run_test("/probe/systeminfo.test", "TestRealSystemdServiceAndTimerLifecycle")
        if runtime_binary:
            delegated = "/sys/fs/cgroup/blorae2e.slice/blora-runtime-e2e"
            subprocess.run([
                "docker", "exec", name, "python3", "-c",
                "import json,pathlib; print(json.dumps({d:{f:(pathlib.Path(d)/f).read_text().strip() for f in ['cgroup.controllers','cgroup.subtree_control','cgroup.type','cgroup.procs']} for d in ['/sys/fs/cgroup','/sys/fs/cgroup/blorae2e.slice']}))",
            ], check=True, timeout=10)
            subprocess.run([
                "docker", "exec", name, "/bin/sh", "-c",
                'printf "+cpu +memory +pids" > /sys/fs/cgroup/blorae2e.slice/cgroup.subtree_control',
            ], check=True, timeout=10)
            subprocess.run(["docker", "exec", name, "mkdir", delegated], check=True, timeout=10)
            subprocess.run([
                "docker", "exec", name, "/bin/sh", "-c",
                'printf "+cpu +memory +pids" > "$1/cgroup.subtree_control"', "sh", delegated,
            ], check=True, timeout=10)
            run_test("/probe/runtime.test", "TestDelegatedCgroupRealLaunch",
                     ["--env", f"BLORA_TEST_CGROUP_ROOT={delegated}"])
            run_test("/probe/runtime.test", "TestDelegatedCgroupResourceLimits",
                     ["--env", f"BLORA_TEST_CGROUP_ROOT={delegated}"])
            run_test("/probe/runtime.test", "TestDelegatedCgroupMemoryExhaustion",
                     ["--env", f"BLORA_TEST_CGROUP_ROOT={delegated}"])
            run_test("/probe/runtime.test", "TestDelegatedCgroupPidsExhaustion",
                     ["--env", f"BLORA_TEST_CGROUP_ROOT={delegated}"])
            subprocess.run(["docker", "exec", name, "rmdir", delegated], check=True, timeout=10)
    finally:
        if created:
            # Docker removes only this randomly named container, even on failure.
            subprocess.run(["docker", "rm", "--force", name], check=True, timeout=30,
                           stdout=subprocess.DEVNULL)
    print(json.dumps({"result": "PASSED", "cleanedUp": True,
                      "seconds": time.monotonic() - started}), flush=True)


if __name__ == "__main__":
    main()
