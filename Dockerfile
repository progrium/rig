FROM golang:alpine
RUN apk add --no-cache nodejs npm bash sudo
RUN go install golang.org/x/tools/gopls@latest
RUN adduser -D -s /bin/bash -G root claude && \
    echo "claude ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
RUN npm install -g @anthropic-ai/claude-code
WORKDIR /go/src/github.com/progrium/rig
COPY go.mod go.sum ./
RUN go mod download
COPY web/system/package.json web/system/package-lock.json ./web/system/
RUN cd web/system && npm ci
COPY . .
RUN mkdir -p /home/claude/.claude
RUN cp CLAUDE.md /home/claude/.claude/CLAUDE.md
RUN cp .claude.json /home/claude/.claude.json
RUN cp .credentials.json /home/claude/.claude/.credentials.json
RUN chown -R claude:root /home/claude/.claude /home/claude/.claude.json
RUN cd web/system && npm run compile-web
RUN CGO_ENABLED=0 go build -o /bin/rig ./cmd/rig
WORKDIR /src
RUN go mod init main \
    && go mod edit -replace github.com/progrium/rig=/go/src/github.com/progrium/rig \
    && go mod tidy
WORKDIR /
ENTRYPOINT ["/bin/rig", "serve"]
