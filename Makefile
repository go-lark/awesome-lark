all: generate

generate:
	( cd .github && go build -o /tmp/awesome-lark-generate main.go )
	/tmp/awesome-lark-generate