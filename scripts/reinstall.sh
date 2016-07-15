#!/usr/bin/env bash

glide install
cf uninstall-plugin v3_beta

set -xe

go build
cf install-plugin v3-cli-plugin -f