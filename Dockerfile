# syntax=docker/dockerfile:1

# Multistage build to generate the smallest possible runtime image.

##
## BUILD
##
# Start with the stable Debian image (bullseye in 01/2022) corresponding
# to the version of Go we are using.
#
# By labeling this stage, we can later prune the intermediate images if we
# want using something like:
#  docker image prune --filter label=stage=builder
#  podman image prune --filter label=stage=builder
FROM golang:1.18.3-bullseye AS build
LABEL stage=builder

# Create a working directory. This also tells Docker to use this directory
# as the default destination for all subsequent commands. This way we
# don't have to type out full file paths, but rather can use relative paths
# based on this directory.
WORKDIR /app

# Copy go.mod and go.sum so `go mod download` will know what to get
COPY go.mod ./
COPY go.sum ./

# Get all of the dependencies. Note, this works exactly the same as if we
# were running `go` locally on our own machine, but this time the Go
# modules are installed into a directory inside the image.
RUN go mod download

# At this point we have an image with all necessary dependencies installed.
# Copy the source code into the image.
# Note, there is some optimization that has happened here. Each COPY command
# changes the dependency graph for Docker layers. Since our code changes
# more often than our dependencies, copying go.mod and go.sum above means
# we are more likely able to use a cache up to this step.
COPY *.go ./

# Build the application and stash it in the root of the image.
# NOTE: `CGO_ENABLED=0` generates a completely stand alone application which
# is what we want since we are using the distroless/static image below.
RUN CGO_ENABLED=0 go build -o /ping-pong-ding-dong

##
## Deploy
##
# This uses a "Distroless" image. This is a project from Google to create
# images that contain only an application and its runtime dependencies.
# These do not contain package managers, shells, or any other programs that
# you would expect to find in a standard Linux distribution.
#
# https://github.com/GoogleContainerTools/distroless
#
# The gcr.io/distroless/static image is for statically compiled appliations
# that do not require `libc`.
#
# https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md
#
FROM gcr.io/distroless/static

COPY --from=build /ping-pong-ding-dong /ping-pong-ding-dong

EXPOSE 8080

USER nonroot:nonroot

# Specify what to execute when we start this image.
# NOTE: using ENTRYPOINT instead of CMD means that this cannot be overridden
# by CLI options when starting the container.
ENTRYPOINT ["/ping-pong-ding-dong"]
