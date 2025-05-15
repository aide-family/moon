out ?= $(out)
target ?= $(target)

.PHONY: gen-plugin
gen-plugin:
	@if [ -z "$(out)" ]; then echo "out is required"; echo "usage: make gen-plugin out=<out> target=<target>"; exit 1; fi
	@if [ -z "$(target)" ]; then echo "target is required"; echo "usage: make gen-plugin out=<out> target=<target>"; exit 1; fi
	go build -buildmode=plugin -o $(out)_plugin.so $(target)