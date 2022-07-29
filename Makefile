PWD=$(shell pwd)
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort )
TOOLS_MOD_DIR := ./internal/tools
ADDLICENSE=addlicense
ALL_SRC := $(shell find . -name '*.go' -o -name '*.sh' -o -name 'Dockerfile' -type f | sort)
GIT_SHA=$(shell git rev-parse --short HEAD)
NAMESPACE=bindplane-dev
OUTDIR=./build

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ifeq ($(GOARCH), amd64)
GOARCH_FULL=amd64_v1
endif

.PHONY: gomoddownload
gomoddownload:
	go mod download

.PHONY: install-git-hooks
install-git-hooks:
	cp scripts/git_hooks/* .git/hooks/

.PHONY: install-tools
install-tools: install-git-hooks
	cd $(TOOLS_MOD_DIR) && go install github.com/securego/gosec/v2/cmd/gosec
	cd $(TOOLS_MOD_DIR) && go install github.com/google/addlicense
	cd $(TOOLS_MOD_DIR) && go install github.com/swaggo/swag/cmd/swag
	cd $(TOOLS_MOD_DIR) && go install github.com/99designs/gqlgen
	cd $(TOOLS_MOD_DIR) && go install github.com/mgechev/revive
	cd $(TOOLS_MOD_DIR) && go install github.com/uw-labs/lichen
	cd $(TOOLS_MOD_DIR) && go install honnef.co/go/tools/cmd/staticcheck
	cd $(TOOLS_MOD_DIR) && go install github.com/client9/misspell/cmd/misspell
	cd $(TOOLS_MOD_DIR) && go install github.com/ory/go-acc

.PHONY: install-ui
install-ui:
	cd ui && npm install

.PHONY: install
install: install-tools install-ui

.PHONY: ci
ci:
	cd ui && npm ci

# dev runs go serve, ui proxy server, and ui graphql generator
.PHONY: dev
dev:
	./ui/node_modules/.bin/concurrently -c blue,magenta,cyan -n sv,ui,gq "go run ./cmd/bindplane/main.go serve --force-console-color --env development" "cd ui && npm start" "cd ui && npm run generate:watch"

.PHONY: test
test: prep
	go test ./... -race -cover -timeout 60s

.PHONY: test-with-cover
test-with-cover: prep
	BINDPLANE_TEST_IMAGE="observiq/bindplane-amd64:$(GIT_SHA)" go-acc --tags=integration --output=coverage.out --ignore=generated --ignore=mocks ./...

show-coverage: test-with-cover
	# Show coverage as HTML in the default browser.
	go tool cover -html=coverage.out

.PHONY: bench
bench:
	go test -benchmem -run=^$$ -bench ^* ./...

.PHONY: tidy
tidy:
	$(MAKE) for-all CMD="rm -fr go.sum"
	$(MAKE) for-all CMD="go mod tidy"

.PHONY: lint
lint:
	revive -formatter friendly -exclude "internal/graphql/schema.*" -set_exit_status ./...
	cd ui && npm run lint && cd ..

.PHONY: vet
vet:
	GOOS=darwin go vet ./...
	GOOS=linux go vet ./...
	GOOS=windows go vet ./...

.PHONY: secure
secure: prep
	gosec -exclude-generated -exclude-dir internal/tools ./...

.PHONY: generate
generate:
	go generate ./...
	@$(MAKE) add-license

.PHONY: swagger
swagger:
	swag init --parseDependency --parseInternal -g model/rest.go -o docs/swagger/
	@$(MAKE) add-license

.PHONY: for-all
for-all:
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && $${CMD} ); \
	done

# TODO(jsirianni): Add secure: https://github.com/observIQ/bindplane/issues/478
.PHONY: ci-check
ci-check: vet test lint check-license scan-licenses

.PHONY: check-license
check-license:
	@ADDLICENSEOUT=`$(ADDLICENSE) -check $(ALL_SRC) 2>&1`; \
		if [ "$$ADDLICENSEOUT" ]; then \
			echo "$(ADDLICENSE) FAILED => add License errors:\n"; \
			echo "$$ADDLICENSEOUT\n"; \
			echo "Use 'make add-license' to fix this."; \
			exit 1; \
		else \
			echo "Check License finished successfully"; \
		fi

.PHONY: add-license
add-license:
	@ADDLICENSEOUT=`$(ADDLICENSE) -y "" -c "observIQ, Inc." $(ALL_SRC) 2>&1`; \
		if [ "$$ADDLICENSEOUT" ]; then \
			echo "$(ADDLICENSE) FAILED => add License errors:\n"; \
			echo "$$ADDLICENSEOUT\n"; \
			exit 1; \
		else \
			echo "Add License finished successfully"; \
		fi

.PHONY: scan-licenses
scan-licenses:
	lichen --config=./license.yaml $$(find build/bindplane* | xargs)

# TLS will run the tls generation script only when the
# tls directory is missing
tls:
	mkdir tls
	docker run \
		-v ${PWD}/scripts/generate-dev-certificates.sh:/generate-dev-certificates.sh \
		-v ${PWD}/tls:/tls \
		--entrypoint=/bin/sh \
		alpine/openssl /generate-dev-certificates.sh

.PHONY: docker-http
docker-http:
	docker run -d -p 3010:3001 \
		--name "bindplane-server-${GIT_SHA}-http" \
		-e BINDPLANE_CONFIG_SESSIONS_SECRET=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		-e BINDPLANE_CONFIG_LOG_OUTPUT=stdout \
		"observiq/bindplane-$(GOARCH):${GIT_SHA}" \
		--host 0.0.0.0 \
		--port "3001" \
		--server-url http://localhost:3010 \
		--remote-url ws://localhost:3010 \
		--secret-key 403dd8ff-72a9-4401-9a66-e54b37d6e0ce
	docker logs "bindplane-server-${GIT_SHA}-http"

	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile set docker-http \
		--server-url http://localhost:3010 --remote-url ws://localhost:3010
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile use docker-http

.PHONY: docker-https
docker-https: tls
	docker run -d \
		-p 3011:3001 \
		--name "bindplane-server-${GIT_SHA}-https" \
		-e BINDPLANE_CONFIG_SESSIONS_SECRET=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		-e BINDPLANE_CONFIG_LOG_OUTPUT=stdout \
		-v "${PWD}/tls:/tls" \
		"observiq/bindplane-$(GOARCH):latest" \
			--tls-cert /tls/bindplane.crt --tls-key /tls/bindplane.key \
			--host 0.0.0.0 \
			--port "3001" \
			--server-url https://localhost:3011 \
			--remote-url wss://localhost:3011 \
			--secret-key 403dd8ff-72a9-4401-9a66-e54b37d6e0ce
	docker logs "bindplane-server-${GIT_SHA}-https"

	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile set docker-https \
		--server-url https://localhost:3011 --remote-url wss://localhost:3011 \
		--tls-ca tls/bindplane-ca.crt
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile use docker-https

.PHONY: docker-https-mtls
docker-https-mtls: tls
	docker run -d \
		-p 3012:3001 \
		--name "bindplane-server-${GIT_SHA}-https-mtls" \
		-e BINDPLANE_CONFIG_SESSIONS_SECRET=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		-e BINDPLANE_CONFIG_LOG_OUTPUT=stdout \
		-v "${PWD}/tls:/tls" \
		"observiq/bindplane-$(GOARCH):latest" \
			--tls-cert /tls/bindplane.crt --tls-key /tls/bindplane.key --tls-ca /tls/bindplane-ca.crt --tls-ca /tls/test-ca.crt \
			--host 0.0.0.0 \
			--port "3001" \
			--server-url https://localhost:3012 \
			--remote-url wss://localhost:3012 \
			--secret-key 403dd8ff-72a9-4401-9a66-e54b37d6e0ce
	docker logs  "bindplane-server-${GIT_SHA}-https-mtls"

	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile set docker-https-mtls \
		--server-url https://localhost:3012 --remote-url wss://localhost:3012 \
		--tls-cert tls/bindplane-client.crt --tls-key ./tls/bindplane-client.key --tls-ca tls/bindplane-ca.crt
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile use docker-https-mtls

.PHONY: docker-all
docker-all: docker-clean docker-http docker-https docker-https-mtls

.PHONY: docker-clean
docker-clean:
	docker ps -a | grep bindplane-server | awk '{print $$1}' | xargs -I{} docker rm --force {}

# Call 'release-test' first.
.PHONY: inspec-continer-image
inspec-continer-image: prep docker-http
	docker exec -u root bindplane-server-${GIT_SHA}-http apt-get update -qq
	docker exec -u root bindplane-server-${GIT_SHA}-http apt-get install -qq -y procps net-tools
	cinc-auditor exec test/inspec/docker/integration.rb -t "docker://bindplane-server-${GIT_SHA}-http"

.PHONY: run
run: docker-http

# Called by commands such as 'vet', useful when the ui has not
# been built before (in ci)
prep: ui/build
ui/build:
	mkdir ui/build
	touch ui/build/index.html

.PHONY: ui-test
ui-test:
	cd ui && CI=true npm run test --watchAll

# ui-build builds the static site to be embeded into the Go binary.
# make install should be called before, if you are not up to date.
.PHONY: ui-build
ui-build:
	cd ui && npm run build

# goreleaser will call ui-build to ensure the static site
# is up to date. goreleaser will not call `make install`.
.PHONY: build
build:
	goreleaser build --rm-dist --skip-validate --single-target --snapshot

.PHONY: clean
clean:
	rm -rf $(OUTDIR)

.PHONY: release-test
release-test:
	goreleaser release --rm-dist --skip-publish --skip-validate --snapshot

# Kitchen prep will build a release and ensure the required
# gems are installed for using Kitchen with GCE
.PHONY: kitchen-prep
kitchen-prep: release-test
	sudo cinc gem install --no-user-install kitchen-google
	sudo cinc gem install --no-user-install kitchen-sync
	mkdir -p dist/kitchen
	cp dist/bindplane_*amd64.deb dist/kitchen
	cp dist/bindplane_*amd64.rpm dist/kitchen

# Assumes you have a ssh key pair at ~/.ssh/id_rsa && ~/.ssh/id_rsa.pub
# Assumes you are authenticated to GCP with Gcloud SDK
#
# Run all tests:
#   make kitchen
# Run tests against specific OS:
#   make kitchen ARGS=sles
.PHONY: kitchen
kitchen:
	kitchen test -c 10 $(ARGS)

.PHONY: kitchen-clean
kitchen-clean:
	kitchen destroy -c 10

ALLDOC=$(shell find . \( -name "*.md" -o -name "*.yaml" \) | grep -v ui/node_modules)

.PHONY: misspell
misspell:
	misspell -error $(ALLDOC)

.PHONY: misspell-fix
misspell-fix:
	misspell -w $(ALLDOC)

