# Use the base image purely for building
FROM golang:1.22rc1 as builder

WORKDIR /build

# Copy the module meta files and download module dependencies as a build step,
# so that later layers can re-use the dependencies until the code changes.
COPY go.mod go.sum ./
RUN go mod download -x

# Copy the rest of the app code over
COPY . ./

# Build the app binary
#
# Some notes:
#  - CGO_ENABLED=0 makes sure that we don't link against C libraries that won't
#  be available later
#  - GOOS=linux describes our target OS for RUNTIME, not build
#  - The -mod=readonly flag ensures that go.mod/go.sum are immutable at build
#  - The -trimpath flag removes directory paths (like `/build`) from the
#  resulting binary, so logs and traces will be cleaner
#  - We use the `-o` flag to name our build so that the resulting binary file
#  name is deterministic
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -trimpath -v -o app

# Use the scratch image for runtime
# (Yay "Multi-Stage Builds"!)
FROM scratch

# Copy the built binary from the builder stage
COPY --from=builder /build/app /app

EXPOSE 80

ENTRYPOINT ["/app"]
