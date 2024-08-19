#!/bin/bash
mkdir -p /etc/vdev

docker run --name dcu-exporter-v2 -d --privileged
--device=/dev/kfd \
--device=/dev/mkfd \
--device=/dev/dri \
-v /etc/vdev:/etc/vdev \
-v /etc/hostname:/etc/hostname \
-e LD_LIBRARY_PATH=$LD_LIBRARY_PATH \
-p 16080:16080 dcu-exporter:v2.0.1