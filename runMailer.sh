#!/usr/bin/env bash

which redis-server > /dev/null 2>&1
if [ "$?" == "1" ]; then
	brew update & brew install redis
fi

killall -KILL "redis-server *:6379" > /dev/null 2>&1
rm dump.rdb mailer.log > /dev/null 2>&1

set -e
go build
redis-server --port 6379 & VERBOSITY=high AUTH_TOKEN=A_LONG_RANDOM_STRING SMTP_URL=localhost:1025 ./mailer-daemon

