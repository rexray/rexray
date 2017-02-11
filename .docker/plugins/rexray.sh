#!/bin/bash
set -e

# first arg is `-f` or `--some-option`
if [ "${1:0:1}" = '-' ]; then
	set -- rexray start -f "$@"
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
		var="REXRAY_${rexray_option^^}"
		val="${!var}"
		if [ "$val" ]; then
			sed -ri 's/^([\ ]*'"$rexray_option"':).*/\1 '"$val"'/' /etc/rexray/rexray.yml
		fi
	done

fi

exec "$@"
