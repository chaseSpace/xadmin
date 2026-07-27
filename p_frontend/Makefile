SHELL := /bin/bash

ENV ?= dev
override ENV := $(or $(strip $(ENV)),dev)

NGINX_DIR ?= /usr/share/nginx/html_xadmin_$(ENV)

ifeq ($(ENV),prod)
VITE_MODE := production
else ifeq ($(ENV),dev)
VITE_MODE := development
else
VITE_MODE := $(ENV)
endif

.PHONY: help pull install build deploy up

help:
	@echo "Common targets:"
	@echo "  make up ENV=prod NGINX_DIR=/usr/share/nginx/html  # pull + install + build dist + deploy to nginx dir"
	@echo "  make up                                           # defaults to ENV=dev"
	@echo "Variables:"
	@echo "  ENV       dev, beta, prod, or another Vite mode; default: dev"

pull:
	@echo "==> Git pull"
	@git pull --ff-only

install:
	@echo "==> Installing frontend dependencies"
	@HUSKY=0 pnpm install --frozen-lockfile --ignore-scripts

build:
	@echo "==> Building frontend (ENV=$(ENV), mode=$(VITE_MODE))"
	@pnpm exec vite build --mode "$(VITE_MODE)"

deploy:
ifeq ($(ENV),dev)
	@echo "==> Skip deploy when ENV=dev; build output stays in dist."
else
	@echo "==> Deploying dist to $(NGINX_DIR)"
	@test -d dist || (echo "dist directory not found, run make build first" && exit 1)
	@command -v rsync >/dev/null 2>&1 || (echo "rsync is required to deploy dist" && exit 1)
	@mkdir -p "$(NGINX_DIR)"
	@rsync -a --delete dist/ "$(NGINX_DIR)/"
	@echo "Deployed to $(NGINX_DIR)"
endif

up: pull install build deploy
