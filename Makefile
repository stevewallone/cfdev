.PHONY: all
all: cfdev # cf-deps.iso cfdev-efi.iso cfdev

vpath %.iso output

cf-deps.iso: ./scripts/build-cf-deps-iso $(wildcard src/builder/**/*) $(wildcard ../bosh-deployment/**/*) $(wildcard ../cf-deployment/**/*) $(wildcard ../cf-mysql-deployment/**/*)
	./scripts/build-cf-deps-iso

cfdev-efi.iso: ./scripts/build-image $(wildcard linuxkit/**/*)
	./scripts/build-image

cfdev: $(wildcard src/code.cloudfoundry.org/{cfdev,cfdevd}/**/*.go)
	(cd src/code.cloudfoundry.org/cfdev && ./generate-plugin.sh)
