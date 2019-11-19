SHELL := /bin/bash

# Project variables
DEP_VERSION = 0.5.0
PACKAGE = github.com/eclipse-iofog/iofog-kubelet
BINARY_NAME = iofog-kubelet
IMAGE = iofog/iofog-kubelet
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
github_repo := iofog-kubelet/iofog-kubelet
binary := iofog-kubelet
build_tags := "netgo osusergo $(VK_BUILD_TAGS)"

BRANCH ?= $(TRAVIS_BRANCH)
RELEASE_TAG ?= 0.0.0

# comment this line out for quieter things
#V := 1 # When V is set, print commands and build progress.

# Space separated patterns of packages to skip in list, test, format.
IGNORED_PACKAGES := /vendor/

.PHONY: all
all: test build

.PHONY: safebuild
# safebuild builds inside a docker container with no clingons from your $GOPATH
safebuild:
	@echo "Building..."
	$Q docker build --build-arg BUILD_TAGS=$(build_tags) -t $(IMAGE):$(VERSION) -f Dockerfile .

bin/dep: bin/dep-$(DEP_VERSION)
	@ln -sf dep-$(DEP_VERSION) bin/dep
bin/dep-$(DEP_VERSION):
	@mkdir -p bin
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | INSTALL_DIRECTORY=bin DEP_RELEASE_TAG=v$(DEP_VERSION) sh
	@mv bin/dep $@

.PHONY: vendor
vendor: bin/dep ## Install dependencies
	bin/dep ensure -v -vendor-only

.PHONY: build
build: authors
	@echo "Building..."
	$Q CGO_ENABLED=0 go build -a --tags $(build_tags) -ldflags '-extldflags "-static"' -o bin/$(binary) $(if $V,-v) $(VERSION_FLAGS) $(PACKAGE)

.PHONY: tags
tags:
	@echo "Listing tags..."
	$Q @git tag

.PHONY: release
release: build $(GOPATH)/bin/goreleaser
	goreleaser


### Code not in the repository root? Another binary? Add to the path like this.
# .PHONY: otherbin
# otherbin: .GOPATH/.ok
#   $Q go install $(if $V,-v) $(VERSION_FLAGS) $(PACKAGE)/cmd/otherbin

##### ^^^^^^ EDIT ABOVE ^^^^^^ #####

##### =====> Utility targets <===== #####

.PHONY: setup
deps: setup
	@echo "Ensuring Dependencies..."
	$Q go env
	$Q dep ensure

.PHONY: build-img
build-img:
	docker build --rm -t $(IMAGE):latest -f Dockerfile .

.PHONY: push-img
push-img:
	@echo $(DOCKER_PASS) | docker login -u $(DOCKER_USER) --password-stdin
ifeq ($(BRANCH), master)
	# Master branch
	docker push $(IMAGE):latest
	docker tag $(IMAGE):latest $(IMAGE):$(RELEASE_TAG)
	docker push $(IMAGE):$(RELEASE_TAG)
endif
ifneq (,$(findstring release,$(BRANCH)))
	# Release branch
	docker tag $(IMAGE):latest $(IMAGE):rc-$(RELEASE_TAG)
	docker push $(IMAGE):rc-$(RELEASE_TAG)
else
	# Develop and feature branches
	docker tag $(IMAGE):latest $(IMAGE)-$(BRANCH):latest
	docker push $(IMAGE)-$(BRANCH):latest
	docker tag $(IMAGE):latest $(IMAGE)-$(BRANCH):$(COMMIT_HASH)
	docker push $(IMAGE)-$(BRANCH):$(COMMIT_HASH)
endif

.PHONY: clean
clean:
	@echo "Clean..."
	$Q rm -rf bin

vet:
	@echo "go vet'ing..."
ifndef CI
	@echo "go vet'ing Outside CI..."
	$Q go vet $(allpackages)
else
	@echo "go vet'ing in CI..."
	$Q mkdir -p test
	$Q ( go vet $(allpackages); echo $$? ) | \
       tee test/vet.txt | sed '$$ d'; exit $$(tail -1 test/vet.txt)
endif

.PHONY: test
test:
	@echo "Testing..."
	$Q go test $(if $V,-v) -i $(allpackages) # install -race libs to speed up next run
ifndef CI
	@echo "Testing Outside CI..."
	$Q GODEBUG=cgocheck=2 go test $(allpackages)
else
	@echo "Testing in CI..."
	$Q mkdir -p test
	$Q ( GODEBUG=cgocheck=2 go test -v $(allpackages); echo $$? ) | \
       tee test/output.txt | sed '$$ d'; exit $$(tail -1 test/output.txt)
endif

.PHONY: list
list:
	@echo "List..."
	@echo $(allpackages)

.PHONY: cover
cover: $(GOPATH)/bin/gocovmerge
	@echo "Coverage Report..."
	@echo "NOTE: make cover does not exit 1 on failure, don't use it to check for tests success!"
	$Q rm -f .GOPATH/cover/*.out cover/all.merged
	$(if $V,@echo "-- go test -coverpkg=./... -coverprofile=cover/... ./...")
	@for MOD in $(allpackages); do \
        go test -coverpkg=`echo $(allpackages)|tr " " ","` \
            -coverprofile=cover/unit-`echo $$MOD|tr "/" "_"`.out \
            $$MOD 2>&1 | grep -v "no packages being tested depend on"; \
    done
	$Q gocovmerge cover/*.out > cover/all.merged
ifndef CI
	@echo "Coverage Report..."
	$Q go tool cover -html .GOPATH/cover/all.merged
else
	@echo "Coverage Report In CI..."
	$Q go tool cover -html .GOPATH/cover/all.merged -o .GOPATH/cover/all.html
endif
	@echo ""
	@echo "=====> Total test coverage: <====="
	@echo ""
	$Q go tool cover -func .GOPATH/cover/all.merged

.PHONY: format
format: $(GOPATH)/bin/goimports
	@echo "Formatting..."
	$Q find . -iname \*.go | grep -v \
        -e "^$$" $(addprefix -e ,$(IGNORED_PACKAGES)) | xargs goimports -w

# skaffold deploys the iofog-kubelet to the Kubernetes cluster targeted by the current kubeconfig using skaffold.
# The current context (as indicated by "kubectl config current-context") must be one of "minikube" or "docker-for-desktop".
# MODE must be set to one of "dev" (default), "delete" or "run", and is used as the skaffold command to be run.
.PHONY: skaffold
skaffold: MODE ?= dev
skaffold: PROFILE := local
skaffold: VK_BUILD_TAGS ?= no_alibabacloud_provider no_aws_provider no_azure_provider no_azurebatch_provider no_cri_provider no_huawei_provider no_hyper_provider no_vic_provider no_web_provider
skaffold:
	@if [[ ! "minikube,docker-for-desktop" =~ .*"$(kubectl_context)".* ]]; then \
		echo current-context is [$(kubectl_context)]. Must be one of [minikube,docker-for-desktop]; false; \
	fi
	@if [[ ! "$(MODE)" == "delete" ]]; then \
		GOOS=linux GOARCH=amd64 VK_BUILD_TAGS="$(VK_BUILD_TAGS)" $(MAKE) build; \
	fi
	@skaffold $(MODE) \
		-f $(PWD)/hack/skaffold/iofog-kubelet/skaffold.yml \
		-p $(PROFILE)

# e2e runs the end-to-end test suite against the Kubernetes cluster targeted by the current kubeconfig.
# It automatically deploys the iofog-kubelet with the mock provider by running "make skaffold MODE=run".
# It is the caller's responsibility to cleanup the deployment after running this target (e.g. by running "make skaffold MODE=delete").
.PHONY: e2e
e2e: KUBECONFIG ?= $(HOME)/.kube/config
e2e: NAMESPACE := default
e2e: NODE_NAME := vkubelet-mock-0
e2e: TAINT_KEY := iofog-kubelet.io/provider
e2e: TAINT_VALUE := mock
e2e: TAINT_EFFECT := NoSchedule
e2e:
	@$(MAKE) skaffold MODE=delete && kubectl delete --ignore-not-found node $(NODE_NAME)
	@$(MAKE) skaffold MODE=run
	@cd $(PWD)/test/e2e && go test -v -tags e2e ./... \
		-kubeconfig=$(KUBECONFIG) \
		-namespace=$(NAMESPACE) \
		-node-name=$(NODE_NAME) \
		-taint-key=$(TAINT_KEY) \
		-taint-value=$(TAINT_VALUE) \
		-taint-effect=$(TAINT_EFFECT)

##### =====> Internals <===== #####

.PHONY: setup
setup: clean
	@echo "Setup..."
	if ! grep "/bin" .gitignore > /dev/null 2>&1; then \
        echo "/bin" >> .gitignore; \
    fi
	if ! grep "/cover" .gitignore > /dev/null 2>&1; then \
        echo "/cover" >> .gitignore; \
    fi
	mkdir -p cover
	mkdir -p bin
	mkdir -p test
	go get -u github.com/golang/dep/cmd/dep
	go get github.com/wadey/gocovmerge
	go get golang.org/x/tools/cmd/goimports
	go get github.com/mitchellh/gox
	go get github.com/goreleaser/goreleaser

MAJOR            ?= 0
MINOR            ?= 0
PATCH            ?= 0
SUFFIX           ?= -dev
VERSION          := $(MAJOR).$(MINOR).$(PATCH)$(SUFFIX)
DATE             := $(shell date -u '+%Y-%m-%d-%H:%M UTC')
VERSION_FLAGS    := -ldflags='-X "github.com/eclipse-iofog/iofog-kubelet/versions.Version=$(VERSION)" -X "github.com/eclipse-iofog/iofog-kubelet/versions.BuildTime=$(DATE)"'

# assuming go 1.9 here!!
_allpackages = $(shell go list ./...)

# memoize allpackages, so that it's executed only once and only if used
allpackages = $(if $(__allpackages),,$(eval __allpackages := $$(_allpackages)))$(__allpackages)


Q := $(if $V,,@)


$(GOPATH)/bin/gocovmerge:
	@echo "Checking Coverage Tool Installation..."
	@test -d $(GOPATH)/src/github.com/wadey/gocovmerge || \
        { echo "Vendored gocovmerge not found, try running 'make setup'..."; exit 1; }
	$Q go install github.com/wadey/gocovmerge
$(GOPATH)/bin/goimports:
	@echo "Checking Import Tool Installation..."
	@test -d $(GOPATH)/src/golang.org/x/tools/cmd/goimports || \
        { echo "Vendored goimports not found, try running 'make setup'..."; exit 1; }
	$Q go install golang.org/x/tools/cmd/goimports

$(GOPATH)/bin/goreleaser:
	go get -u github.com/goreleaser/goreleaser

authors:
	$Q git log --all --format='%aN <%cE>' | sort -u  | sed -n '/github/!p' > GITAUTHORS
	$Q cat AUTHORS GITAUTHORS  | sort -u > NEWAUTHORS
	$Q mv NEWAUTHORS AUTHORS
	$Q rm -f NEWAUTHORS
	$Q rm -f GITAUTHORS
