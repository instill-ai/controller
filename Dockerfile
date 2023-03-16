ARG GOLANG_VERSION
ARG UBUNTU_VERSION

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} AS build

ARG SERVICE_NAME

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG TARGETOS TARGETARCH
RUN --mount=target=. --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /${SERVICE_NAME} ./cmd/main

# Mounting points
RUN mkdir /etc/vdp
RUN mkdir /vdp

FROM --platform=$BUILDPLATFORM ubuntu:${UBUNTU_VERSION}

# Need permission of /tmp folder for internal process such as store temporary files.
RUN chown -R nobody:nogroup /tmp
# Need permission of /nonexistent folder for HuggingFace internal process.
RUN mkdir /nonexistent > /dev/null && chown -R nobody:nogroup /nonexistent

USER nobody:nogroup

ARG SERVICE_NAME

WORKDIR /${SERVICE_NAME}

COPY --from=docker:dind-rootless --chown=nobody:nogroup /usr/local/bin/docker /usr/local/bin

COPY --from=build --chown=nobody:nogroup /src/config ./config
COPY --from=build --chown=nobody:nogroup /src/assets ./assets
COPY --from=build --chown=nobody:nogroup /src/release-please ./release-please

COPY --from=build --chown=nobody:nogroup /${SERVICE_NAME} ./

COPY --from=build --chown=nobody:nogroup /etc/vdp /etc/vdp
COPY --from=build --chown=nobody:nogroup /vdp /vdp
