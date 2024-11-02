#!/bin/sh

set -eu

HOST="${HOST:-}"
TOKEN="${TOKEN:-}"
TEAM_UUID="${TEAM_UUID:-}"
PERMISSION="${PERMISSION:-}"

if [ -z "${HOST}" ]; then
	echo -n "Enter Host for DependencyTrack: "
	read -r HOST
	if [ -z "${HOST}" ]; then
		echo "Host is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${TOKEN}" ]; then
	echo -n "Enter Token for DependencyTrack: "
	read -r TOKEN
	if [ -z "${TOKEN}" ]; then
		echo "Token is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${TEAM_UUID}" ]; then
	echo -n "Enter Team UUID for DependencyTrack: "
	read -r TEAM_UUID
	if [ -z "${TEAM_UUID}" ]; then
		echo "Team UUID is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${PERMISSION}" ]; then
	echo -n "Enter Permission to assign to team in DependencyTrack: "
	read -r PERMISSION
	if [ -z "${PERMISSION}" ]; then
		echo "Permission Name is required" >> /dev/stderr
		exit 1
	fi
fi

curl --fail-with-body --request POST "${HOST}/api/v1/permission/${PERMISSION}/team/${TEAM_UUID}" --header "Authorization: Bearer ${TOKEN}"