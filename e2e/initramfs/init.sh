#!/bin/sh
busybox modprobe i6300esb nowayout=1 heartbeat=3
busybox mknod /dev/watchdog c 10 130

busybox mount -t proc none /proc

busybox sleep 2

config=$(busybox sed -e 's/.*config=//g' /proc/cmdline)
echo
echo
echo
echo "=== Starting ${config}"
/usr/src/voff -config "${config}"

echo '=== Terminated'
read