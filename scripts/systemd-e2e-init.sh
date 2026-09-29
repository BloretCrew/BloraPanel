#!/bin/sh
set -eu
# Docker already supplied a private cgroup namespace. Refuse an unexpected
# mount or host-visible cgroup path before making this container subtree writable.
test "$(cat /proc/self/cgroup)" = '0::/'
test -f /.dockerenv
test -f /sys/fs/cgroup/cgroup.controllers
mount -o remount,rw /sys/fs/cgroup
printf '%s\n' disposable-blora-systemd-e2e > /run/blora-systemd-e2e
# Activate accounting before docker exec adds a process to the namespace root.
# Otherwise enabling a new domain controller can violate cgroup's no-internal-
# processes rule. This slice belongs only to this disposable container.
mkdir -p /run/systemd/system/basic.target.wants
printf '[Unit]\nDescription=Blora isolated cgroup probe\nBefore=basic.target\n[Slice]\nCPUAccounting=yes\nCPUQuota=100%%\nMemoryAccounting=yes\nTasksAccounting=yes\n' > /run/systemd/system/blorae2e.slice
ln -s ../blorae2e.slice /run/systemd/system/basic.target.wants/blorae2e.slice
exec /usr/lib/systemd/systemd --system --unit=basic.target --log-target=console
