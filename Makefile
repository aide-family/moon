# app subdirs only (basename), e.g. rabbit, marksman
APPS := $(notdir $(shell find app -maxdepth 1 -mindepth 1 -type d))

.PHONY: $(APPS)
# run the app binary in development mode
$(APPS):
	cd app/$@ && make dev

.PHONY: all
# run all apps in development mode
all:
	@cd magicbox && make proto;
	@for app in $(APPS); do \
		echo "=========build $$app =========="; \
		cd app/$$app && make all; \
		cd -; \
	done

.PHONY: gen
# generate the gen files
gen:
	@for app in $(APPS); do \
		echo "=========generate $$app =========="; \
		cd app/$$app && make gen; \
		cd -; \
	done

.PHONY: build-exe
# build the exe files
build-exe:
	@cd magicbox && make proto;
	@mkdir -p package
	@for app in $(APPS); do \
		echo "=========build-exe $$app =========="; \
		cd app/$$app; \
		make clean; \
		make build-exe; \
		cp .env bin/; \
		mv bin $$app; \
		zip -r $$app-exe.zip $$app/; \
		rm -rf $$app; \
		cd -; \
		mv app/$$app/$$app-exe.zip package/; \
		rm -rf app/$$app/$$app-exe.zip; \
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