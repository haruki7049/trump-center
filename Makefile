.PHONY: build clean run

build:
	nu ./scripts/build.nu

clean:
	nu ./scripts/clean.nu

run: build
	nu ./scripts/run.nu
