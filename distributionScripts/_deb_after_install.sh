#!/bin/bash

addgroup --system newslettersender
adduser --system --ingroup newslettersender newslettersender
chown -R newslettersender:newslettersender /var/log/sandstorm-newsletter-sender/

if [ ! -f /etc/default/sandstorm-newsletter-sender ]; then

PASS=`openssl rand -base64 32`

    cat >/etc/default/sandstorm-newsletter-sender <<EOL

# you can also set this to something else, the following token was generated on installation
export AUTH_TOKEN=${PASS}

# export REDIS_URL=localhost:6379
# export SMTP_URL=localhost:25

EOL
fi