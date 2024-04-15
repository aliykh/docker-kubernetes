############################
# STEP 1 build executable binary
############################
FROM golang:1.22-alpine as builder

# Ensure ca-certficates are up to date
RUN update-ca-certificates

RUN apk add make

# Set the working directory to the root of your Go module
WORKDIR /myapp

# Add cache for faster builds
ENV GOCACHE=$HOME/.cache/go-build
RUN --mount=type=cache,target=$GOCACHE

# use modules
COPY go.mod .

RUN go mod download && go mod verify

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /bin/app ./cmd

############################
# STEP 2 build a small image
############################
# using base nonroot image
# user:group is nobody:nobody, uid:gid = 65534:65534
FROM gcr.io/distroless/static

# Copy our static executable
COPY --from=builder /bin/app /bin/app
EXPOSE 8080 80
# Run the hello binary.
ENTRYPOINT ["/bin/app"]


