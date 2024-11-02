#!/bin/sh

unset DATA
HOST="${HOST:-}"
USERNAME="${USERNAME:-}"
PASSWORD="${PASSWORD:-}"

if [ -z "${HOST}" ]; then
	echo -n "Enter Host for DependencyTrack: "
	read -r HOST
	if [ -z "${HOST}" ]; then
		echo "Host is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${USERNAME}" ]; then
	echo -n "Enter Username to sign in as: "
	read -r USERNAME
	if [ -z "${USERNAME}" ]; then
		echo "Username is required" >> /dev/stderr
		exit 1
	fi
fi

if [ -z "${PASSWORD}" ]; then
	echo -n "Enter Password for ${USERNAME}: "
	read -r PASSWORD
	if [ -z "${PASSWORD}" ]; then
		echo "Current password is required" >> /dev/stderr
		exit 1
	fi
fi

DATA="username=${USERNAME}&password=${PASSWORD}"

curl --request POST --fail-with-body "${HOST}/api/v1/user/login" --header "Content-Type: application/x-www-form-urlencoded" --data "${DATA}"
