FROM golang:alpine
RUN apk add --no-cache nodejs npm
WORKDIR /go/src/github.com/progrium/rig
COPY go.mod go.sum ./
RUN go mod download
COPY web/system/package.json web/system/package-lock.json ./web/system/
RUN cd web/system && npm ci
COPY . .
RUN cd web/system && npm run compile-web
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/rig ./cmd/rig
WORKDIR /src
RUN go mod init main \
    && go mod edit -replace github.com/progrium/rig=/go/src/github.com/progrium/rig \
    && go mod tidy
WORKDIR /
ENTRYPOINT ["/bin/rig", "serve"]
