#!/bin/sh

set -eu

unset DATA
HOST="${HOST:-}"
USERNAME="${USERNAME:-}"
PASSWORD="${PASSWORD:-}"
NEW_PASSWORD="${NEW_PASSWORD:-}"

if [ -z "${HOST}" ]; then
	echo -n "Enter Host for DependencyTrack: "
	read -r HOST
	if [ -z "${HOST}" ]; then
		echo "Host is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${USERNAME}" ]; then
	echo -n "Enter Username to change password: "
	read -r USERNAME
	if [ -z "${USERNAME}" ]; then
		echo "Username is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${PASSWORD}" ]; then
	echo -n "Enter Current password for ${USERNAME}: "
	read -r PASSWORD
	if [ -z "${PASSWORD}" ]; then
		echo "Current password is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${NEW_PASSWORD}" ]; then
	echo -n "Enter New password for ${USERNAME}: "
	read -r NEW_PASSWORD
	if [ -z "${NEW_PASSWORD}" ]; then
		echo "New password is required" >> /dev/stderr
		exit 1
	fi
fi

DATA="username=${USERNAME}&password=${PASSWORD}&newPassword=${NEW_PASSWORD}&confirmPassword=${NEW_PASSWORD}"

curl --fail-with-body "${HOST}/api/v1/user/forceChangePassword" --request POST --header "Content-Type: application/x-www-form-urlencoded" --data "${DATA}"
