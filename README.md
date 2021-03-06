# Project Sandstorm Mailer -- Sending Daemon

**The project has been sponsored by [Swisscom](https://www.swisscom.ch), and initiated by [Web Essentials](http://www.web-essentials.asia/). Thanks for your support!**

This is a tool which adjusts Neos in a way such that it can be used to send Newsletters.

Is is comprised of two parts:

- a [Go Daemon](https://github.com/sandstorm/mailer-daemon) **this package** which does the actual Mail-Sending
- a [Neos package](https://github.com/sandstorm/Sandstorm.Newsletter) which provides the User Interface

This repository is containing the Go Daemon which does the mail-sending.

## Installation of the mailer-sender

* The compiled Mailer Deb Package can be found in "sandstorm-newsletter_1.0_i386.deb" next to this README.
* When installing this package, it will create a system user and group `newslettersender` under which the newsletter sender will execute.
* The mailer log can be found in `/var/log/sandstorm-newsletter-sender/mailer.log`.
* In order to start the sender, run `service sandstorm-newsletter-sender start`. If it crashes, it restarts automatically.
* You can configure the defaults in `/etc/default/sandstorm-newsletter-sender`, currently with the following ones being supported:

  * AUTH_TOKEN -- must be set to a random "password"; and the exact same key must be defined inside Neos Setting as "Sandstorm.Newsletter.auth_token"
                  so that Neos can connect to the mailer.
  * REDIS_URL (defaults to localhost:6379)
  * SMTP_URL (defaults to localhost:25)

  * The backend server currently listens on localhost:3000 -- this currently cannot be changed (but it should be changeable)

## Development Setup

The following instructions have been tested using *Go 1.6*.

```
# Ensure you have a working go installation and gopath is set:
echo $GOPATH

# clone the mailer daemon and its dependencies
go get github.com/sandstorm/mailer-daemon

# Try to run the mailer. should start up, but display an error about AUTH_TOKEN not being set.
$GOPATH/bin/mailer-daemon

# To run the tests, do the following:
cd $GOPATH/src/github.com/sandstorm/mailer-daemon/
./runAllGoTests.sh
```

To run the mailer during development, do the following:

* Ensure you have redis installed and on your path
* We suggest to use MailHog (https://github.com/mailhog/MailHog/releases) to test the mail-sending

* Then, use the `runMailer.sh` script:

```
cd $GOPATH/src/github.com/sandstorm/mailer-daemon/
./runMailer.sh
```

## Generating Production Build

```
# initial setup
cd $GOPATH/src/github.com/sandstorm/mailer-daemon/distributionScripts
brew install gnu-tar
bundle install
go get github.com/mitchellh/gox
$GOPATH/bin/gox -build-toolchain


# actual build:
./build-deb-package.sh
```
