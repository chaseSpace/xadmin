package resource

import (
	"strings"
	"testing"
)

func TestFileExistsFilterConditionEscapesExistsColumn(t *testing.T) {
	if !strings.Contains(fileExistsFilterCondition, `"exists" = ?`) {
		t.Fatalf("expected escaped exists column, got condition: %s", fileExistsFilterCondition)
	}
	if strings.Contains(fileExistsFilterCondition, " exists = ?") || strings.Contains(fileExistsFilterCondition, "`exists`") {
		t.Fatalf("expected no unescaped exists column, got condition: %s", fileExistsFilterCondition)
	}
}

func TestFileKeywordFilterConditionIsGrouped(t *testing.T) {
	if !strings.HasPrefix(fileKeywordFilterCondition, "(") || !strings.HasSuffix(fileKeywordFilterCondition, ")") {
		t.Fatalf("expected grouped keyword condition, got condition: %s", fileKeywordFilterCondition)
	}
}
