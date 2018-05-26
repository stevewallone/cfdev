#!/usr/bin/env bash

set -ex

go build -ldflags '-X code.cloudfoundry.org/cfdev/config.cfdepsUrl=file:///Users/dgodd/.cfdev/cache/cf-deps.iso
     -X code.cloudfoundry.org/cfdev/config.cfdepsMd5=0eeb68be5c72d92a4e96bab3aca3c808
     -X code.cloudfoundry.org/cfdev/config.cfdepsSize=4699973632

     -X code.cloudfoundry.org/cfdev/config.cfdevefiUrl=file:///Users/dgodd/.cfdev/cache/cfdev-efi.iso
     -X code.cloudfoundry.org/cfdev/config.cfdevefiMd5=98e87be65a2c02f728acc7e6349331be
     -X code.cloudfoundry.org/cfdev/config.cfdevefiSize=330201088

     -X code.cloudfoundry.org/cfdev/config.vpnkitUrl=file:///Users/dgodd/.cfdev/cache/vpnkit
     -X code.cloudfoundry.org/cfdev/config.vpnkitMd5=f551649c099fd32e2914690bc2f7b84b
     -X code.cloudfoundry.org/cfdev/config.vpnkitSize=19655400

     -X code.cloudfoundry.org/cfdev/config.hyperkitUrl=file:///Users/dgodd/.cfdev/cache/hyperkit
     -X code.cloudfoundry.org/cfdev/config.hyperkitMd5=61da21b4e82e2bf2e752d043482aa966
     -X code.cloudfoundry.org/cfdev/config.hyperkitSize=3691536

     -X code.cloudfoundry.org/cfdev/config.linuxkitUrl=file:///Users/dgodd/.cfdev/cache/linuxkit
     -X code.cloudfoundry.org/cfdev/config.linuxkitMd5=da8048c89e1cfa1f2a95ea27e83ae94c
     -X code.cloudfoundry.org/cfdev/config.linuxkitSize=44150800

     -X code.cloudfoundry.org/cfdev/config.qcowtoolUrl=file:///Users/dgodd/.cfdev/cache/qcow-tool
     -X code.cloudfoundry.org/cfdev/config.qcowtoolMd5=22f3a57096ae69027c13c4933ccdd96c
     -X code.cloudfoundry.org/cfdev/config.qcowtoolSize=4104388

     -X code.cloudfoundry.org/cfdev/config.uefiUrl=file:///Users/dgodd/.cfdev/cache/UEFI.fd
     -X code.cloudfoundry.org/cfdev/config.uefiMd5=2eff1c02d76fc3bde60f497ce1116b09
     -X code.cloudfoundry.org/cfdev/config.uefiSize=2097152

     -X code.cloudfoundry.org/cfdev/config.cfdevdUrl=file:///Users/dgodd/workspace/cfdev/src/code.cloudfoundry.org/cfdev/cfdevd
     -X code.cloudfoundry.org/cfdev/config.cfdevdMd5=f9e475d0c356b485e1a1a1a8712215fa
     -X code.cloudfoundry.org/cfdev/config.cfdevdSize=4715896

     -X code.cloudfoundry.org/cfdev/config.cliVersion=0.0.20180525-200459
     -X code.cloudfoundry.org/cfdev/config.analyticsKey=WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2' code.cloudfoundry.org/cfdev


# cfdev="/Users/pivotal/workspace/cfdev"
# dir="$( cd "$( dirname "$0" )" && pwd )"
# cfdev="$dir"/../../..
# cache_dir="$HOME"/.cfdev/cache
#
# export GOPATH="$cfdev"
# pkg="code.cloudfoundry.org/cfdev/config"
#
# export GOOS=darwin
# export GOARCH=amd64
#
# go build code.cloudfoundry.org/cfdevd
# cfdevd="$PWD"/cfdevd
#
# cfdepsUrl="$cfdev/output/cf-deps.iso"
# if [ ! -f "$cfdepsUrl" ]; then
#   cfdepsUrl="$cache_dir/cf-deps.iso"
# fi
# cfdevefiUrl="$cfdev/output/cfdev-efi.iso"
# if [ ! -f "$cfdevefiUrl" ]; then
#   cfdevefiUrl="$cache_dir/cfdev-efi.iso"
# fi
#
# go build \
#   -ldflags \
#     "-X $pkg.cfdepsUrl=file://$cfdepsUrl
#      -X $pkg.cfdepsMd5=$(md5 $cfdepsUrl | awk '{ print $4 }')
#      -X $pkg.cfdepsSize=$(wc -c < $cfdepsUrl | tr -d '[:space:]')
#
#      -X $pkg.cfdevefiUrl=file://$cfdevefiUrl
#      -X $pkg.cfdevefiMd5=$(md5 $cfdevefiUrl | awk '{ print $4 }')
#      -X $pkg.cfdevefiSize=$(wc -c < $cfdevefiUrl | tr -d '[:space:]')
#
#      -X $pkg.vpnkitUrl=file://$cache_dir/vpnkit
#      -X $pkg.vpnkitMd5=$(md5 "$cache_dir"/vpnkit | awk '{ print $4 }')
#      -X $pkg.vpnkitSize=$(wc -c < "$cache_dir"/vpnkit | tr -d '[:space:]')
#
#      -X $pkg.hyperkitUrl=file://$cache_dir/hyperkit
#      -X $pkg.hyperkitMd5=$(md5 "$cache_dir"/hyperkit | awk '{ print $4 }')
#      -X $pkg.hyperkitSize=$(wc -c < "$cache_dir"/hyperkit | tr -d '[:space:]')
#
#      -X $pkg.linuxkitUrl=file://$cache_dir/linuxkit
#      -X $pkg.linuxkitMd5=$(md5 "$cache_dir"/linuxkit | awk '{ print $4 }')
#      -X $pkg.linuxkitSize=$(wc -c < "$cache_dir"/linuxkit | tr -d '[:space:]')
#
#      -X $pkg.qcowtoolUrl=file://$cache_dir/qcow-tool
#      -X $pkg.qcowtoolMd5=$(md5 "$cache_dir"/qcow-tool | awk '{ print $4 }')
#      -X $pkg.qcowtoolSize=$(wc -c < "$cache_dir"/qcow-tool | tr -d '[:space:]')
#
#      -X $pkg.uefiUrl=file://$cache_dir/UEFI.fd
#      -X $pkg.uefiMd5=$(md5 "$cache_dir"/UEFI.fd | awk '{ print $4 }')
#      -X $pkg.uefiSize=$(wc -c < "$cache_dir"/UEFI.fd | tr -d '[:space:]')
#
#      -X $pkg.cfdevdUrl=file://$cfdevd
#      -X $pkg.cfdevdMd5=$(md5 "$cfdevd" | awk '{ print $4 }')
#      -X $pkg.cfdevdSize=$(wc -c < "$cfdevd" | tr -d '[:space:]')
#
#      -X $pkg.cliVersion=0.0.$(date +%Y%m%d-%H%M%S)
#      -X $pkg.analyticsKey=WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2" \
#      code.cloudfoundry.org/cfdev
#
#
