SHELL := /bin/bash
ENV ?= dev
override ENV := $(or $(strip $(ENV)),dev)
BIN ?= ./zzz/xadmin_$(ENV)
NOHUP ?= ./nohup_$(ENV).log
SQL_DIR ?=
TABLE ?=

.PHONY: help before_run pb build pull up stop logs \
	fmt test vet execsql showtable

help:
	@echo "Common targets:"
	@echo "  make up ENV=dev          # generate pb + build + start or restart server"
	@echo "  make stop                # stop server"
	@echo "  make logs                # tail server log"
	@echo "  make execsql ENV=prod    # execute model SQL from config/<ENV>/app.yaml"
	@echo "  make showtable TABLE=xx  # show table ddl and columns"

before_run:
	@mkdir -p .ignore_logs .ignore_acme zzz

pb:
	@echo "==> Generating protobuf"
	@bash ./shell/gen_pb.sh

pull:
	@echo "==> Git pull"
	@git pull

build: before_run
	@echo "==> Building server binary"
	@go build -o $(BIN) ./cmd/server/main.go

up: pull pb build
	@echo "==> Running server (ENV=$(ENV))"
	-@pkill -f "$(BIN)" || true
	@APP_ENV=$(ENV) nohup $(BIN) > $(NOHUP) 2>&1 &
	@echo "Server started, PID: $$(pgrep -f '$(BIN)')"
	@sleep 2
	@tail -n 20 $(NOHUP)

stop:
	@echo "==> Stopping server"
	-@pkill -f "$(BIN)" || true

logs:
	@tail -n 100 -f $(NOHUP)

fmt:
	@echo "==> gofmt"
	@gofmt -w ./cmd ./internal ./pkg ./test

test:
	@echo "==> go test"
	@go test ./...

vet:
	@echo "==> go vet"
	@go vet ./...

execsql:
	@echo "==> Execute SQL"
	@echo "    make execsql ENV=dev"
	@echo "    make execsql ENV=prod                          (direct network)"
	@echo "    make execsql ENV=prod HOST=127.0.0.1 PORT=15432  (via socks5 proxy)"
	@echo "    Proxy: go run tool/socksfwd.go 127.0.0.1:7890 <db_host>:5432 15432"
	@echo ""
	@bash ./shell/import_model_sql.sh --env "$(ENV)" $(if $(SQL_DIR),--sql-dir "$(SQL_DIR)",) $(if $(HOST),--host "$(HOST)",) $(if $(PORT),--port "$(PORT)",)

showtable:
	@echo "==> Show table structure"
	@if [ -z "$(TABLE)" ]; then \
		echo "Usage: make showtable TABLE=admin_user_session"; \
		exit 1; \
	fi
	@APP_ENV="$(ENV)" go run ./tool/showtable.go "$(TABLE)"
