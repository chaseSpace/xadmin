package bootstrap

import (
	"log"
	"strings"
	"time"
)

func InitProgramTimezone(timezone string) {
	nextTimezone := strings.TrimSpace(timezone)
	if nextTimezone == "" {
		log.Fatalf("timezone is empty")
	}
	loc, err := time.LoadLocation(nextTimezone)
	if err != nil {
		log.Fatalf("Failed to load timezone %q: %v", nextTimezone, err)
	}
	time.Local = loc
}
