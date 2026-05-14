VSCODE_URL	?= https://github.com/progrium/vscode-web/releases/download/v1/vscode-web-1.108.2.zip

dev:
	docker build -t rig . && docker run -v $(shell pwd)/web/editors:/go/src/github.com/progrium/rig/web/editors -p 8080:8080 rig
.PHONY: dev

web/vscode:
	curl -sL $(VSCODE_URL) -o web/vscode.zip
	mkdir -p .tmp
	unzip web/vscode.zip -d .tmp
	mv .tmp/dist/vscode web/vscode
	rm -rf .tmp
	rm web/vscode.zip

toolkit-go:
	git clone https://github.com/tractordev/toolkit-go

go.work: toolkit-go
	go work init . ./toolkit-go