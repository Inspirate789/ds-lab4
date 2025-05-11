# Choose whether to use migrations
ARG MODE=with_migrations

FROM golang:1.23.1-bookworm AS build
WORKDIR /build
ENV CGO_ENABLED=0

# Install dependencies
COPY go.* .
RUN go mod download

# Get path to main.go
ARG MAIN_PATH

# Build the binary
# '--mount=target=.': use bind mounting from the build context
# '--mount=type=cache,target=/root/.cache/go-build': use go’s compiler cache
RUN --mount=target=. \
    --mount=type=cache,target=/root/.cache/go-build \
    go build \
    -trimpath -ldflags "-s -w -extldflags '-static'" \
    -o /app $MAIN_PATH
RUN chmod +x /app

FROM scratch AS stage_with_migrations
ARG MIGRATIONS_FOLDER
ONBUILD COPY $MIGRATIONS_FOLDER /migrations

FROM scratch AS stage_without_migrations

FROM stage_${MODE} AS app
# Add label to image
ARG PIPELINE_ID
LABEL version="$PIPELINE_ID"
# Copy the binary
COPY --from=build /app /app
# Get path to config
ARG CONFIG_PATH
# Create environment
COPY $CONFIG_PATH /app.yaml
# Run the binary
ENTRYPOINT ["/app", "--config=/app.yaml"]
