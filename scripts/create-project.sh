#!/bin/sh

set -eu

unset DATA
HOST="${HOST:-}"
TOKEN="${TOKEN:-}"
PROJECT_NAME="${PROJECT_NAME:-}"
PROJECT_VERSION="${PROJECT_VERSION:-}"

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

if [ -z "${PROJECT_NAME}" ]; then
	echo -n "Enter Project Name for DependencyTrack: "
	read -r PROJECT_NAME
	if [ -z "${PROJECT_NAME}" ]; then
		echo "Project Name is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${PROJECT_VERSION}" ]; then
	echo -n "Enter Project Version for Project in DependencyTrack: "
	read -r PROJECT_VERSION
	if [ -z "${PROJECT_VERSION}" ]; then
		echo "Project Version is required" >> /dev/stderr
		exit 1
	fi
fi

DATA="{\"name\":\"${PROJECT_NAME}\",\"version\":\"${PROJECT_VERSION}\",\"accessTeams\":[],\"active\":true,\"isLatest\":true,\"parent\":null,\"tags\":[{\"Name\":\"project_data_test_tag\"}]}"
# TODO: The "isLatest" key is introduced in API v4.12. For compatibility, this is removed, though still mentioned for reference.
DATA=$(echo $DATA | jq 'del(.isLatest)')
curl --fail-with-body --request PUT "${HOST}/api/v1/project" --header "Authorization: Bearer ${TOKEN}" --header "Content-Type: application/json" --data "${DATA}"
