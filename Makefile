.PHONY: build clean run test fmt lint update

build:
	@nu ./scripts/build.nu

clean:
	@nu ./scripts/clean.nu

run: build
	./target/bin/trump-center

test:
	@nu ./scripts/test.nu

fmt:
	@nu ./scripts/fmt.nu

lint:
	@nu ./scripts/lint.nu

update:
	@nu ./scripts/update.nu
