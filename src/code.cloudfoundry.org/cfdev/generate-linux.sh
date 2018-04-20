#!/usr/bin/env bash

set -ex

export GOPATH="$( cd "$( dirname "$0" )/../../../" && pwd )"
pkg="code.cloudfoundry.org/cfdev/config"

export GOOS=linux
export GOARCH=amd64

go build \
  -ldflags \
    "-X $pkg.cfdepsUrl=https://s3.amazonaws.com/cfdev-ci/cf-oss-deps/cf-oss-dep-0.29.0.iso
     -X $pkg.cfdepsMd5=5aa54a6595bddcafc40209a47de89a9b
     -X $pkg.cfdepsSize=4037984256

     -X $pkg.gdnUrl=https://github.com/cloudfoundry/garden-runc-release/releases/download/v1.12.1/gdn-1.12.1
     -X $pkg.gdnMd5=2abeba3fc15a6015684c05c3cc2a90f5
     -X $pkg.gdnSize=31656409

     -X $pkg.analyticsKey=WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2" \
     code.cloudfoundry.org/cfdev

     # -X $pkg.gdnUrl=https://github.com/cloudfoundry/garden-runc-release/releases/download/v1.12.1/gdn-1.12.1
     # -X $pkg.gdnMd5=2abeba3fc15a6015684c05c3cc2a90f5
     # -X $pkg.gdnSize=31656409
     #
     # -X $pkg.gdnUrl=https://github.com/cloudfoundry/garden-runc-release/releases/download/v1.9.4/gdn-1.9.4
     # -X $pkg.gdnMd5=fd3a127a7644785e3974bed4b60c90bb
     # -X $pkg.gdnSize=24430365

