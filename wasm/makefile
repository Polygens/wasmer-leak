SHELL := /bin/bash
.DEFAULT_GOAL := help

build: ## Build a WASM binary for production
	cargo build --target wasm32-wasi --release
