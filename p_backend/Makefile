SHELL := /bin/bash
ENV ?= dev
BIN := ./zzz/xadmin
SQL_DIR ?=
TABLE ?=

.PHONY: help before_run pb build pull run restart stop logs \
	fmt test vet check-no-automigrate check-sql-docs check-api-docs check-api-tests check-docs \
	verify verify-fast verify-auth execsql showtable

help:
	@echo "Common targets:"
	@echo "  make run ENV=dev         # generate pb + build + run server"
	@echo "  make restart ENV=dev     # restart server"
	@echo "  make stop                # stop server"
	@echo "  make logs                # tail server log"
	@echo "  make verify              # format + test + docs/rules checks"
	@echo "  make verify-fast         # fast checks for daily dev loop"
	@echo "  make verify-auth         # run auth API tests only"
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

run: pull pb build
	@echo "==> Running server (ENV=$(ENV))"
	-@pkill -f "$(BIN)" || true
	@APP_ENV=$(ENV) nohup $(BIN) > nohup.log 2>&1 &
	@echo "Server started, PID: $$(pgrep -f '$(BIN)')"
	@sleep 2
	@tail -n 20 nohup.log

restart: stop run

stop:
	@echo "==> Stopping server"
	-@pkill -f "$(BIN)" || true

logs:
	@tail -n 100 -f nohup.log

fmt:
	@echo "==> gofmt"
	@gofmt -w ./cmd ./internal ./pkg ./test

test:
	@echo "==> go test"
	@go test ./...

vet:
	@echo "==> go vet"
	@go vet ./...

check-no-automigrate:
	@echo "==> Checking AutoMigrate usage"
	@if rg -n "AutoMigrate" ./cmd ./internal ./pkg ./test; then \
		echo "Found forbidden AutoMigrate usage"; \
		exit 1; \
	fi

check-sql-docs:
	@echo "==> Checking SQL docs"
	@test -d docs/model_sql_pg
	@rg -n "CREATE TABLE|ALTER TABLE|CREATE INDEX|DROP INDEX" docs/model_sql_pg

check-api-docs:
	@echo "==> Checking API docs"
	@test -d docs/api_docs
	@rg -n "^#|^##|GET |POST |PUT |DELETE |PATCH " docs/api_docs/*.MD

check-api-tests:
	@echo "==> Checking API tests"
	@test -d test
	@rg --files test -g "*_test.go"

check-docs: check-sql-docs check-api-docs check-api-tests

verify: fmt test vet check-no-automigrate check-docs
	@echo "==> Verify passed"

verify-fast: fmt verify-auth check-no-automigrate check-docs
	@echo "==> Verify-fast passed"

verify-auth:
	@echo "==> Running auth API tests"
	@go test ./test -run Auth -v

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
