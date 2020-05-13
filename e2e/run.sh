#!/bin/sh
set -e

apk add --no-cache --no-scripts linux-lts

mkinitfs \
    -P /usr/src/initramfs/features \
    -i /usr/src/initramfs/init.sh \
    -F "base voff" \
    -o /tmp/initramfs \
    $(basename /lib/modules/*)

set +e
exitcode=0

./run.exp 5seconds.yaml
[ $? -ne 0 ] && "echo ::error::5 seconds test failed" && exitcode=1

./run.exp forever.yaml
[ $? -ne 4 ] && "echo ::error::Forever timeout test failed" && exitcode=1

echo "Test suite complete"
exit $exitcode