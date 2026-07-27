#image indakum(os base,langiage,library) enitu containeril store akum,container has its own os,each container has its onw fie systemand independet,small computer virtual,
# syntax=docker/dockerfile:1

# ---- builder ----
    #fempty file sustem,linux,go compiler
FROM golang:1.25-alpine AS builder

# Install ca-certificates for https access (if needed)
RUN apk add --no-cache ca-certificates git
#diretory
WORKDIR /app

# copy go.mod and go.sum first to take advantage of Docker cache
#library,depency,layer layer ayita docker build aka,cache aki vekum frequntlu use akan-dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy the rest of the source
COPY . .

# build statically (disable cgo),compile akitu binary file,server namil tsore akum
ENV CGO_ENABLED=0
RUN go build -o /app/server ./main.go    # adjust path if your main is elsewhere

# ---- final image ----,linux image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# copy the binary from builder,multistage build-first stage binary then 2nd stage linux image,layer-os,dependecy,source code,compiled bonary file
#oru layer mathram change varutha
COPY --from=builder /app/server .

# optional: create non-root user (recommended for security)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# expose port (match what your app listens on)
EXPOSE 8080

# default command
ENTRYPOINT ["./server"]