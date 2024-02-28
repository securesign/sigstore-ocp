
CHART_VERSION ?= ""
CHART_URL ?= "./charts/trusted-artifact-signer"

LDFLAGS=-X securesign/sigstore-ocp/tas-installer/cmd.helmChartVersion=$(CHART_VERSION) \
        -X securesign/sigstore-ocp/tas-installer/cmd.helmChartUrl=$(CHART_URL)

PLATFORMS=darwin linux windows
ARCHITECTURES=amd64 arm64

.PHONY: build
build: build-tas-installer

.PHONY: test
test: test-tas-installer

.PHONY: cross
cross: cross-tas-installer

.PHONY: build-tas-installer
build-tas-installer:
	CGO_ENABLED=0 go build -C ./tas-installer -trimpath -ldflags "$(LDFLAGS)" -o ../tas-install

.PHONY: test-tas-installer
test-tas-installer:
	cd ./tas-installer && go test ./...

.PHONY: cross-tas-installer
cross-tas-installer:
	$(foreach GOOS, $(PLATFORMS),\
		$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); \
	go build -C ./tas-installer -trimpath -ldflags "$(LDFLAGS)" -o ../tas-install-$(GOOS)-$(GOARCH))))

.PHONY: clean
clean:
	rm -f tas-install
	rm -f tas-install-*
