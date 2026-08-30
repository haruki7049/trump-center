.PHONY: build clean run test fmt lint update

build:
	@./scripts/build.nu

clean:
	@./scripts/clean.nu

run: build
	./target/bin/trump-center

test:
	@./scripts/test.nu

fmt:
	@./scripts/fmt.nu

lint:
	@./scripts/lint.nu

update:
	@./scripts/update.nu
