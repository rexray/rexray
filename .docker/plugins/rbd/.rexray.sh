#!/bin/bash
# shellcheck shell=dash
set -e

# Workaround Docker bug that mount /sys as ro
mount /sys -o remount,rw

# first arg is `-f` or `--some-option`
if [ "$(echo "$1" | \
	awk  '{ string=substr($0, 1, 1); print string; }' )" = '-' ]; then
	set -- rexray start -f --nopid "$@"
fi

#set default rexray options
REXRAY_FSTYPE="${REXRAY_FSTYPE:-ext4}"
REXRAY_LOGLEVEL="${REXRAY_LOGLEVEL:-warn}"
REXRAY_PREEMPT="${REXRAY_PREEMPT:-false}"

if [ "$1" = 'rexray' ]; then

	for rexray_option in \
		fsType \
		loglevel \
		preempt \
	; do
		val=$(eval echo "\$REXRAY_$(echo $rexray_option | \
			awk '{print toupper($0)}')")
		if [ "$val" ]; then
			sed -ri 's/^([\ ]*'"$rexray_option"':).*/\1 '"$val"'/' \
				/etc/rexray/rexray.yml
		fi
	done

fi

exec "$@"
