#!/bin/bash

set -e
cd ../
$GOPATH/bin/gox -osarch='linux/386'
rm -Rf bundle_work
mkdir bundle_work
cd bundle_work
mkdir -p usr/local/bin
mkdir -p var/log/sandstorm-newsletter-sender

mv ../mailer-daemon_linux_386 usr/local/bin/sandstorm-newsletter-sender

mkdir -p etc/init
cp ../distributionScripts/_upstart_init_script.conf etc/init/sandstorm-newsletter-sender.conf


BUNDLE_GEMFILE=../distributionScripts/Gemfile bundle exec fpm -t deb -s dir --version 1.1.0 -p ../ -n sandstorm-newsletter -a i386 --deb-user newslettersender --deb-group newslettersender --after-install ../distributionScripts/_deb_after_install.sh .
