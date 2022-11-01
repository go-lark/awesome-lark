all: generate

generate:
	( cd ./cmd/build && go build -o /tmp/awesome-lark-generate )
	cd ../..
	/tmp/awesome-lark-generate
