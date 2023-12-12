# Build stage
FROM registry.access.redhat.com/ubi9/go-toolset@sha256:c3a9c5c7fb226f6efcec2424dd30c38f652156040b490c9eca5ac5b61d8dc3ca as build-env
ENV APP_ROOT=/opt/app-root
WORKDIR $APP_ROOT/src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o tas-installer

# Setup stage 
FROM registry.access.redhat.com/ubi9/go-toolset@sha256:c3a9c5c7fb226f6efcec2424dd30c38f652156040b490c9eca5ac5b61d8dc3ca as setup
WORKDIR /opt/app-root/src
USER root

ADD https://github.com/securesign/sigstore-ocp/archive/refs/heads/release-1.0.gamma.tar.gz .

RUN chmod +r release-1.0.gamma.tar.gz && \
    tar -xzvf release-1.0.gamma.tar.gz 

# Deploy stage
FROM registry.access.redhat.com/ubi9/ubi-minimal@sha256:3e313209ac617a92b50350286752311d99ea2dafc429ef0e5311889294b0bc21 as deploy
WORKDIR /opt/app-root/src

LABEL description="TAS installer aims to provide a quick and easy way of installing the Trusted Artifact Signer stack."
LABEL io.k8s.description="TAS installer aims to provide a quick and easy way of installing the Trusted Artifact Signer stack."
LABEL io.k8s.display-name="TAS Installer container image for Red Hat Trusted Artifact Signer"
LABEL io.openshift.tags="tas-installer trusted-signer"
LABEL summary="Provides the tas-installer binary for installing TAS"
LABEL com.redhat.component="tas-installer"

COPY --from=build-env /opt/app-root/src/tas-installer .
COPY --from=setup /opt/app-root/src/sigstore-ocp-release-1.0.gamma/charts ./charts

ENTRYPOINT [ "./tas-installer" ]
