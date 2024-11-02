#!/bin/sh

set -eu

unset DATA
HOST="${HOST:-}"
TOKEN="${TOKEN:-}"

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

DATA="[{\"groupName\":\"vuln-source\",\"propertyName\":\"nvd.enabled\",\"propertyValue\":false}]"
curl --request POST --fail-with-body "${HOST}/api/v1/configProperty/aggregate" --header "Authorization: Bearer ${TOKEN}" --header "Content-Type: application/json" --data "${DATA}"