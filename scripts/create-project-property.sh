#!/bin/sh

set -eu

unset DATA
HOST="${HOST:-}"
TOKEN="${TOKEN:-}"
PROJECT_UUID="${PROJECT_UUID:-}"

GROUP_NAME="${GROUP_NAME:-}"
PROPERTY_NAME="${PROPERTY_NAME:-}"
PROPERTY_VALUE="${PROPERTY_VALUE:-}"
PROPERTY_TYPE="${PROPERTY_TYPE:-}"
DESCRIPTION="${DESCRIPTION:-}"

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

if [ -z "${PROJECT_UUID}" ]; then
	echo -n "Enter Project UUID for DependencyTrack: "
	read -r PROJECT_UUID
	if [ -z "${PROJECT_UUID}" ]; then
		echo "Project UUID is required" >> /dev/stderr
		exit 1
	fi
fi

DATA="{\"groupName\":\"${GROUP_NAME}\",\"propertyName\":\"${PROPERTY_NAME}\",\"propertyValue\":\"${PROPERTY_VALUE}\",\"propertyType\":\"${PROPERTY_TYPE}\",\"description\":\"${DESCRIPTION}\"}"
curl --request PUT --fail-with-body "${HOST}/api/v1/project/${PROJECT_UUID}/property" --header "Authorization: Bearer ${TOKEN}" --header "Content-Type: application/json" --data "${DATA}"
