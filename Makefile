# app subdirs only (basename), e.g. rabbit, marksman
APPS := $(notdir $(shell find app -maxdepth 1 -mindepth 1 -type d))

.PHONY: all
# run all apps in development mode
all:
	@for app in $(APPS); do \
		echo "=========generate $$app =========="; \
		cd app/$$app && make all; \
		cd -; \
	done

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help