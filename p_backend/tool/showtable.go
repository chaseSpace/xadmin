package main

import (
	"fmt"
	"log"
	"os"

	"monorepo/config"
	"monorepo/pkg/db"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "" {
		log.Fatal("usage: go run ./tool/showtable.go <table>")
	}
	table := os.Args[1]
	cfg := config.MustLoadConfig().App.Database
	database := db.GetDatabase()
	if !database.Migrator().HasTable(table) {
		log.Fatalf("table %s does not exist", table)
	}
	columns, err := database.Migrator().ColumnTypes(table)
	if err != nil {
		log.Fatalf("inspect table %s: %v", table, err)
	}
	fmt.Printf("table: %s (%s)\\n", table, cfg.Type)
	for _, column := range columns {
		nullable, _ := column.Nullable()
		length, _ := column.Length()
		fmt.Printf("- %s %s nullable=%v length=%d\\n", column.Name(), column.DatabaseTypeName(), nullable, length)
	}
}
