#!/bin/sh

set -eu

unset DATA
HOST="${HOST:-}"
TOKEN="${TOKEN:-}"
TEAM_NAME="${TEAM_NAME:-}"

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

if [ -z "${TEAM_NAME}" ]; then
	echo -n "Enter Team Name for DependencyTrack: "
	read -r TEAM_NAME
	if [ -z "${TEAM_NAME}" ]; then
		echo "Team Name is required" >> /dev/stderr
		exit 1
	fi
fi

DATA="{\"name\":\"${TEAM_NAME}\"}"
curl --request PUT "${HOST}/api/v1/team" --header "Authorization: Bearer ${TOKEN}" --header "Content-Type: application/json" --data "${DATA}"
