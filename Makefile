output = $(PWD)/output
imgscf = $(PWD)/images/cf
cf_deployment = $(PWD)/../cf-deployment
cf_ops = $(PWD)/images/cf/cf-operations

.PHONY = all cf-deps
all: cf-deps "$(output)"/STEMCELL.txt
cf-deps: "$(output)"/manifest.yml "$(output)"/runtime-config.yml "$(output)"/app-security-group.json "$(output)"/bin/deploy-cf "$(output)"/cloud-config.yml

"$(output)"/runtime-config.yml:
	cp "$(imgscf)"/configs/dns-runtime-config.yml "$(output)"/runtime-config.yml
"$(output)"/app-security-group.json:
	cp "$(imgscf)"/app-security-group.json "$(output)"/app-security-group.json
"$(output)"/bin/deploy-cf:
	mkdir -p "$(output)"/bin
	cp "$(imgscf)"/deploy-cf "$(output)"/bin/deploy-cf
"$(output)"/manifest.yml: $(wildcard "$(cf_deployment)"/**/*.yml)
	bosh int "$(cf_deployment)"/cf-deployment.yml \
		-o "$(cf_deployment)"/operations/use-compiled-releases.yml \
		\
		-o "$(cf_deployment)"/operations/experimental/skip-consul-cell-registrations.yml \
		-o "$(cf_deployment)"/operations/experimental/skip-consul-locks.yml \
		-o "$(cf_deployment)"/operations/experimental/use-bosh-dns-for-containers.yml \
		-o "$(cf_deployment)"/operations/experimental/disable-consul.yml \
		-o "$(cf_deployment)"/operations/bosh-lite.yml \
		-o "$(cf_deployment)"/operations/experimental/disable-consul-bosh-lite.yml \
		\
		-o "$(cf_ops)"/allow-local-docker-registry.yml \
		-o "$(cf_ops)"/garden-disable-app-armour.yml \
		-o "$(cf_ops)"/collocate-tcp-router.yml \
		-o "$(cf_ops)"/set-cfdev-subnet.yml \
		-o "$(cf_ops)"/lower-memory.yml \
		-o "$(cf_ops)"/remove-smoke-test.yml \
		-o "$(cf_ops)"/low-tcp-ports.yml \
		\
		-v cf_admin_password=admin \
		-v uaa_admin_client_secret=admin-client-secret \
		> "$(output)/manifest.yml"
"$(output)"/cloud-config.yml:
	bosh int "$(cf_deployment)"/iaas-support/bosh-lite/cloud-config.yml \
		-o "$(cf_ops)"/set-cloud-config-subnet.yml \
		> "$(output)"/cloud-config.yml

"$(output)"/STEMCELL.txt:
	rq -y < output/cache/manifest.yml | jq -r '.stemcells[0].version' > "$(output)"/STEMCELL.txt

