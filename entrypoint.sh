#!/bin/sh

[ ! -f /conf/hosts ] && touch /conf/hosts

echo Polling every $INTERVAL minutes
/prometheus-certcheck -hosts /conf/hosts -interval $INTERVAL -web.address $ADDRESS -web.telemetry-path $METRICSPATH