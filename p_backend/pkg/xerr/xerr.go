package xerr

import (
	"errors"
	"fmt"
	"runtime"

	"monorepo/pkg/xi18n"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type XErr struct {
	code        Code
	bizCode     string // business error code for i18n lookup
	inputDetail string
	fmtArgs     []any
	stack       []uintptr
}

func New(code Code) *XErr {
	e := &XErr{
		code: code,
	}
	e.setCallerStack()
	return e
}

func NewWithDetail(code Code, detail string, args ...any) *XErr {
	e := &XErr{
		code:        code,
		inputDetail: fmt.Sprintf(detail, args...),
	}
	e.setCallerStack()
	return e
}

// NewBiz creates an error with a business error code for i18n translation.
func NewBiz(code Code, bizCode string, args ...any) *XErr {
	e := &XErr{
		code:    code,
		bizCode: bizCode,
		fmtArgs: args,
	}
	e.setCallerStack()
	return e
}

func NewWithError(code Code, err error, detail string) *XErr {
	e := &XErr{
		code:        code,
		inputDetail: fmt.Sprintf("%s: %v", detail, err),
	}
	e.setCallerStack()
	return e
}

func (e *XErr) Error() string {
	detail := e.inputDetail
	if detail == "" && e.bizCode != "" {
		detail = xi18n.MsgZh(e.bizCode, e.fmtArgs...)
	}
	if detail == "" {
		detail = e.code.Detail()
	}
	return fmt.Sprintf("[%d] - %s", e.code, detail)
}

func (e *XErr) setCallerStack() {
	if e.code == CodeSuccess {
		return
	}
	const skip = 3  // 跳过3层栈帧
	const depth = 5 // 栈跟踪的最大深度
	stack := make([]uintptr, depth)
	n := runtime.Callers(skip, stack)
	e.stack = stack[:n]
}

func (e *XErr) FormatStack() string {
	if e.stack == nil {
		return ""
	}

	frames := runtime.CallersFrames(e.stack)
	var stackInfo string

	for {
		frame, more := frames.Next()
		stackInfo += fmt.Sprintf("\t%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}

	return stackInfo
}

func (e *XErr) Code() Code {
	return e.code
}

func (e *XErr) BizCode() string {
	return e.bizCode
}

func (e *XErr) FmtArgs() []any {
	return e.fmtArgs
}

func (e *XErr) Detail() string {
	if e.inputDetail != "" {
		return e.inputDetail
	}
	if e.bizCode != "" {
		return xi18n.MsgZh(e.bizCode, e.fmtArgs...)
	}
	return e.code.Detail()
}

type Code int

const (
	// 常规错误码
	CodeSuccess            Code = 200
	CodeInternalError      Code = 500
	CodeBadRequest         Code = 400
	CodeNotFound           Code = 404
	CodeUnauthorized       Code = 401
	CodeForbidden          Code = 403
	CodeTimeout            Code = 408
	CodeConflict           Code = 409
	CodeTooManyRequests    Code = 429
	CodeServiceUnavailable Code = 503

	// 外部组件错误码
	CodeDBError         Code = 1000
	CodeRedisError      Code = 1001
	CodeMongoDBError    Code = 1002
	CodeKafkaError      Code = 1003
	CodeParamError      Code = 1004
	CodeIllegalPhoneNum Code = 1005
	CodeIllegalEmail    Code = 1006
	CodeDataNotFound    Code = 1404
)

var codeDetails = map[Code]string{
	CodeSuccess:            "CodeSuccess",
	CodeInternalError:      "CodeInternalError",
	CodeBadRequest:         "CodeBadRequest",
	CodeNotFound:           "CodeNotFound",
	CodeUnauthorized:       "CodeUnauthorized",
	CodeForbidden:          "CodeForbidden",
	CodeTimeout:            "CodeTimeout",
	CodeConflict:           "CodeConflict",
	CodeTooManyRequests:    "CodeTooManyRequests",
	CodeServiceUnavailable: "CodeServiceUnavailable",

	CodeDBError:         "CodeDBError",
	CodeRedisError:      "CodeRDSError",
	CodeMongoDBError:    "CodeMDBError",
	CodeKafkaError:      "CodeKafkaError",
	CodeParamError:      "CodeParamError",
	CodeDataNotFound:    "CodeDataNotFound",
	CodeIllegalPhoneNum: "非法手机号格式",
	CodeIllegalEmail:    "非法邮箱格式",
}

func (c Code) Detail() string {
	if detail, exists := codeDetails[c]; exists {
		return detail
	}
	return "undefined error code"
}

// ToXErr 进程内传递的err，转化为xerr，返回值永远非空
func ToXErr(err error) *XErr {
	if err == nil {
		return New(CodeSuccess)
	}
	var xerr *XErr
	if errors.As(err, &xerr) {
		return xerr
	}
	return NewWithDetail(CodeInternalError, "%s", err.Error())
}

func WrapDB(err error, detail ...string) error {
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return WrapDBE(err, detail...)
}

func WrapDBNotFound(err error, msg string, arg ...any) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewWithDetail(CodeDataNotFound, msg, arg...)
	}
	return WrapDBE(err)
}

func WrapDBE(err error, detail ...string) error {
	if err == nil {
		return nil
	}
	if len(detail) == 0 {
		detail = append(detail, "DB: ")
	}
	return NewWithDetail(CodeDBError, "%s", detail[0]+": "+err.Error())
}

func WrapDBDuplicate(err error, detail string) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		return WrapDBE(err, "failed to save")
	}
	return NewWithDetail(CodeConflict, "%s", detail)
}

func WrapDBUpdateMiss(db *gorm.DB, detail ...string) error {
	if db.Error != nil {
		return WrapDBE(db.Error, "update failed")
	}
	if db.RowsAffected == 0 {
		if len(detail) > 0 {
			return NewWithDetail(CodeDataNotFound, "%s", detail[0])
		}
		return NewWithDetail(CodeDataNotFound, "data not found")
	}
	return nil
}

func WrapRedis(err error, detail ...string) error {
	if err == nil || errors.Is(err, redis.Nil) {
		return nil
	}
	if len(detail) == 0 {
		detail = append(detail, "Redis")
	}
	return NewWithDetail(CodeRedisError, "%s", detail[0]+": "+err.Error())
}
