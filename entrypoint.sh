#!/bin/sh

[ ! -f /config/hosts ] && echo No /config/hosts file!

/prometheus-certcheck -hosts /config/hosts -interval $INTERVAL -web.address $ADDRESS -web.telemetry-path $METRICSPATH