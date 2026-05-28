package util

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/cast"
)

var LegalNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.\-+·\x{4e00}-\x{9fa5}]+$`)

// TruncateString 截断字符串到指定长度并拼接省略号
func TruncateString(s string, maxLen int, suffix ...string) string {
	if maxLen <= 0 {
		return ""
	}

	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	ellipse := "..."
	if len(suffix) > 0 {
		ellipse = suffix[0]
	}
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + ellipse
	}

	return s
}

func NewRequestId() string {
	for i := 0; i < 10; i++ {
		_uuid, err := uuid.NewV7()
		if err != nil {
			pp.Printf("!!! NewUUIDv7 failed %v", err)
			break
		}
		return _uuid.String()
	}
	return cast.ToString(time.Now().UnixMilli())
}

// GenRandomString 生成随机字符串
func GenRandomString(isDigit bool, length int) string {
	if length <= 0 {
		return ""
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var chars string
	if isDigit {
		chars = "0123456789"
	} else {
		chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[r.Intn(len(chars))]
	}

	return string(result)
}

func RandName() string {
	return randomNames[rand.Intn(len(randomNames))] + GenRandomString(true, 4)
}

func ParseFullPhone(src string) (string, string, error) {
	if !strings.HasPrefix(src, "+") {
		return "", "", fmt.Errorf("invalid phone format")
	}
	ss := strings.Split(src, "|")
	if len(ss) != 2 {
		return "", "", fmt.Errorf("invalid phone format")
	}
	return ss[0], ss[1], nil
}
