package middleware

import (
	"fmt"
	"monorepo/pkg/logger"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xfiber"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Recover creates a middleware that recovers from panics and logs the error
func Recover() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Get the error and stack trace
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", r)
				}

				stack := getStackTrace(3)

				// 格式化打印 方便人类读取
				fmt.Printf("Panic: %v\nStack Trace:\n%s\n", err, stack)

				// Log the panic with necessary information
				logger.Error("Panic recovered",
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
					zap.String("ip", c.IP()),
					zap.Error(err),
					zap.String("panic_stack", stack),
				)
				// panic 错误也要标准化
				err = xfiber.StdResponse(c, nil, xerr.NewWithDetail(xerr.CodeInternalError, "%s", err.Error()))
			}
		}()

		return c.Next()
	}
}

// getStackTrace returns the current stack trace with specified skip frames
func getStackTrace(skip int) string {
	// 获取调用栈信息，跳过指定层数
	pc := make([]uintptr, 7)
	n := runtime.Callers(skip, pc)
	if n == 0 {
		return "no stack trace available"
	}

	pc = pc[:n] // 截取实际获取到的程序计数器
	frames := runtime.CallersFrames(pc)

	// 构建堆栈跟踪字符串
	var stackTrace string
	var foundFirstRepoFrame bool
	for {
		frame, more := frames.Next()
		var printFrame bool
		if !strings.Contains(frame.File, "/src/") {
			foundFirstRepoFrame = true
		} else if foundFirstRepoFrame {
			printFrame = true
		}

		if printFrame || foundFirstRepoFrame {
			stackTrace += fmt.Sprintf("\t%s:%d in %s\n", frame.File, frame.Line, frame.Function)
		}

		if !more {
			break
		}
	}

	return stackTrace
}
