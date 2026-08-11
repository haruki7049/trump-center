#!/usr/bin/env nu

def main [] {
  # Create target dir
  mkdir target/bin

  # The path is from project root, because this script is executed by /Makefile
  go build -o target/bin cmd/trump-center/trump-center.go
}
