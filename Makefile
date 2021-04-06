#
# Copyright 2016 the original author or authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# set default shell
SHELL = bash -e -o pipefail

# Variables
VERSION                    ?= $(shell cat ./VERSION)

DOCKER_LABEL_VCS_DIRTY     = false
ifneq ($(shell git ls-files --others --modified --exclude-standard 2>/dev/null | wc -l | sed -e 's/ //g'),0)
    DOCKER_LABEL_VCS_DIRTY = true
endif
## Docker related
DOCKER_EXTRA_ARGS          ?=
DOCKER_REGISTRY            ?=
DOCKER_REPOSITORY          ?=
DOCKER_TAG                 ?= ${VERSION}$(shell [[ ${DOCKER_LABEL_VCS_DIRTY} == "true" ]] && echo "-dirty" || true)
DOCKER_TARGET              ?= prod
ADAPTER_IMAGENAME          := ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}voltha-ofagent-go

## Docker labels. Only set ref and commit date if committed
DOCKER_LABEL_VCS_URL       ?= $(shell git remote get-url $(shell git remote))
DOCKER_LABEL_VCS_REF       = $(shell git rev-parse HEAD)
DOCKER_LABEL_BUILD_DATE    ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
DOCKER_LABEL_COMMIT_DATE   = $(shell git show -s --format=%cd --date=iso-strict HEAD)

DOCKER_BUILD_ARGS ?= \
	${DOCKER_EXTRA_ARGS} \
	--build-arg org_label_schema_version="${VERSION}" \
	--build-arg org_label_schema_vcs_url="${DOCKER_LABEL_VCS_URL}" \
	--build-arg org_label_schema_vcs_ref="${DOCKER_LABEL_VCS_REF}" \
	--build-arg org_label_schema_build_date="${DOCKER_LABEL_BUILD_DATE}" \
	--build-arg org_opencord_vcs_commit_date="${DOCKER_LABEL_COMMIT_DATE}" \
	--build-arg org_opencord_vcs_dirty="${DOCKER_LABEL_VCS_DIRTY}"

DOCKER_BUILD_ARGS_LOCAL ?= ${DOCKER_BUILD_ARGS} \
	--build-arg LOCAL_PROTOS=${LOCAL_PROTOS}

# tool containers
VOLTHA_TOOLS_VERSION ?= 2.3.1

GO                = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang go
GO_JUNIT_REPORT   = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-go-junit-report go-junit-report
GOCOVER_COBERTURA = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app/src/github.com/opencord/ofagent-go -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-gocover-cobertura gocover-cobertura
GOLANGCI_LINT     = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golangci-lint golangci-lint
HADOLINT          = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-hadolint hadolint

.PHONY: docker-build local-protos local-lib-go local-voltha help
.DEFAULT_GOAL := help

## Local Development Helpers

local-protos: ## Copies a local version of the voltha-protos dependency into the vendor directory
ifdef LOCAL_PROTOS
	rm -rf vendor/github.com/opencord/voltha-protos/go
	mkdir -p vendor/github.com/opencord/voltha-protos/go
	cp -r ${GOPATH}/src/github.com/opencord/voltha-protos/go/* vendor/github.com/opencord/voltha-protos/go
endif

## Local Development Helpers
local-lib-go: ## Copies a local version of the voltha-lib-go dependency into the vendor directory
ifdef LOCAL_LIB_GO
	rm -rf vendor/github.com/opencord/voltha-lib-go/v4/pkg
	mkdir -p vendor/github.com/opencord/voltha-lib-go/v4/pkg
	cp -r ${LOCAL_LIB_GO}/pkg/* vendor/github.com/opencord/voltha-lib-go/v4/pkg/
endif


local-voltha: ## Copies a local version of the voltha-go dependency into the vendor directory
ifdef LOCAL_VOLTHA
	rm -rf vendor/github.com/opencord/voltha-go/
	mkdir -p vendor/github.com/opencord/voltha-go/
	cp -rf ${GOPATH}/src/github.com/opencord/voltha-go/ vendor/github.com/opencord/
	rm -rf vendor/github.com/opencord/voltha-go/vendor
endif

## Docker targets
build: docker-build ## Alias for 'docker-build'

## Docker targets

docker-build: local-protos local-voltha local-lib-go ## Build docker image (set BUILD_PROFILED=true to also build the profiled image)
	docker build $(DOCKER_BUILD_ARGS) --target=${DOCKER_TARGET} -t ${ADAPTER_IMAGENAME}:${DOCKER_TAG} -f docker/Dockerfile.ofagent-go .
ifdef BUILD_PROFILED
	docker build $(DOCKER_BUILD_ARGS) --target=${DOCKER_TARGET} --build-arg EXTRA_GO_BUILD_TAGS="-tags profile" -t ${ADAPTER_IMAGENAME}:${DOCKER_TAG}-profile -f docker/Dockerfile.ofagent-go .
endif

docker-push: ## Push the docker images to an external repository
	docker push ${ADAPTER_IMAGENAME}:${DOCKER_TAG}
ifdef BUILD_PROFILED
	docker push ${ADAPTER_IMAGENAME}:${DOCKER_TAG}-profile
endif

## lint and unit tests

lint-dockerfile: ## Perform static analysis on Dockerfile
	@echo "Running Dockerfile lint check..."
	@${HADOLINT} $$(find . -name "Dockerfile.*")
	@echo "Dockerfile lint check OK"

lint-mod: ## Verify the Go dependencies
	@echo "Running dependency check..."
	@${GO} mod verify
	@echo "Dependency check OK. Running vendor check..."
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Staged or modified files must be committed before running this test" && git status -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files must be cleaned up before running this test" && git status -- go.mod go.sum vendor && exit 1)
	${GO} mod tidy
	${GO} mod vendor
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Modified files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@echo "Vendor check OK."

lint: lint-mod lint-dockerfile ## Run all lint targets

test: local-lib-go ## Run unit tests
	@mkdir -p ./tests/results
	@${GO} test -mod=vendor -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	${GO_JUNIT_REPORT} < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	${GOCOVER_COBERTURA} < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	exit $$RETURN

sca: ## Runs static code analysis with the golangci-lint tool
	@rm -rf ./sca-report
	@mkdir -p ./sca-report
	@echo "Running static code analysis..."
	@${GOLANGCI_LINT} run --deadline=4m --out-format junit-xml ./... | tee ./sca-report/sca-report.xml
	@echo ""
	@echo "Static code analysis OK"

clean: ## Removes any local filesystem artifacts generated by a build
	rm -rf ./sca-report

distclean: clean ## Removes any local filesystem artifacts generated by a build or test run
	rm -rf ${VENVDIR}

mod-update: ## Update go mod files
	${GO} mod tidy
	${GO} mod vendor

# For each makefile target, add ## <description> on the target line and it will be listed by 'make help'
help: ## Print help for each Makefile target
	@echo "Usage: make [<target>]"
	@echo "where available targets are:"
	@echo
	@grep '^[[:alpha:]_-]*:.* ##' $(MAKEFILE_LIST) \
		| sort | awk 'BEGIN {FS=":.* ## "}; {printf "%-25s : %s\n", $$1, $$2};'
