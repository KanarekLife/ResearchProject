# The jobs worker (mcbench jobs -watch). CI builds and pushes it to GHCR
# (.github/workflows/worker-image.yml). It has no repository of its own, so the
# commit is baked in:
#   docker build --build-arg COMMIT=$(git rev-parse HEAD) -t mcbench-worker .
FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG COMMIT
RUN test -n "$COMMIT" && CGO_ENABLED=0 go build -ldflags "-X main.buildCommit=$COMMIT" -o /mcbench ./cmd

# GitHub's SSH host keys, as GitHub publishes them, to verify it on every push.
FROM alpine:3.22 AS hostkeys
RUN apk add --no-cache curl jq \
	&& curl -fsSL https://api.github.com/meta | jq -r '.ssh_keys[] | "github.com " + .' > /github_known_hosts \
	&& test -s /github_known_hosts

FROM alpine:3.22
RUN apk add --no-cache git openssh-client ca-certificates
COPY --from=hostkeys /github_known_hosts /etc/ssh/github_known_hosts
WORKDIR /app
COPY --from=build /mcbench /usr/local/bin/mcbench
COPY data/instructions data/instructions
COPY config.yaml deploy/entrypoint.sh ./
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["mcbench", "jobs", "-watch", "5m"]
