# Welcome to the WeCarry.App server code!

## Requirements

This software needs docker-compose version 1.24.0 at least.

## Setting it up

In an appropriate directory:
* git clone git@github.com:silinternational/wecarry-api.git
* cd wecarry-api
* cp .env.example .env
* make

Note that the data provided in `.env.example` will not allow
all features to work. In particular, the `ROLLBAR_TOKEN` must 
be valid. The fake AWS data will work for file uploads to the minIO
container, but obviously not for a real AWS S3 bucket.

## Installation Troubleshooting

Default versions which come with your operating system
may have earlier versions of these installed.

Make sure to uninstall docker-compose packages, and then following the
instructions here: https://docs.docker.com/compose/install/
