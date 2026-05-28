package xfiber

import (
	"fmt"
	"monorepo/pkg/xerr"
	"monorepo/pkg/xi18n"
	"monorepo/proto/xadminpb/commpb"
	"reflect"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/k0kubun/pp/v3"
	"google.golang.org/protobuf/encoding/protojson"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// BindHandler2RouterGroup 将指定 handler 结构体绑定的所有方法以 POST 方法绑定到指定的路由分组中
func BindHandler2RouterGroup(prefix string, parentGroup fiber.Router, handler interface{}) {
	groupRouter := parentGroup.Group(prefix)

	handlerType := reflect.TypeOf(handler)
	handlerValue := reflect.ValueOf(handler)

	errorType := reflect.TypeOf((*error)(nil)).Elem() // 获取 error 接口类型

	// 遍历 handler 结构体的所有方法
	for i := 0; i < handlerType.NumMethod(); i++ {
		method := handlerType.Method(i)
		methodValue := handlerValue.Method(i)

		// (receiver) SomeMethod(*xfiber.Ctx, req) (rsp, err)
		if method.Type.NumIn() != 3 {
			panic(fmt.Errorf("invalid number of input arguments for method %s", method.Name))
		}
		if method.Type.NumOut() != 2 {
			panic(fmt.Errorf("invalid number of output arguments for method %s", method.Name))
		}
		if !method.Type.Out(1).Implements(errorType) {
			panic(fmt.Errorf("method %s does not return an error on second parameter", method.Name))
		}
		paramType := method.Type.In(2)

		// 生成路由路径：将 Handle 后的部分转换为小写并作为路径
		routePath := generateRoutePath(method.Name)

		fullPath := parentGroup.(*fiber.Group).Prefix + prefix + routePath
		pp.Printf("[Route binding] POST - method %s.%s --> %s\n", handlerType.String(), method.Name, fullPath)

		// 将方法绑定到路由组，使用 POST 方法
		groupRouter.Post(routePath, func(c *fiber.Ctx) error {
			paramValue := reflect.New(paramType.Elem()) // init type

			// 使用 protojson 解析 JSON 到 protobuf 对象
			if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).
				Unmarshal(c.Body(), paramValue.Interface().(proto2.Message)); err != nil {
				return StdResponse(c, nil, fmt.Errorf("error unmarshaling request body: %v", err))
			}

			// 手动调用验证方法（pb生成了 Validate 方法）
			if v, ok := paramValue.Interface().(interface{ Validate() error }); ok {
				if err := v.Validate(); err != nil {
					return StdResponse(c, nil, err)
				}
			}
			// 利用反射调用实际的方法（这点性能损失换取的整洁架构完全值得）
			results := methodValue.Call([]reflect.Value{reflect.ValueOf(c), paramValue})

			//fmt.Printf("1111111 len:%d\n", len(results))
			fullMethod := handlerType.String() + "." + method.Name
			// 获取返回值（rsp, err）
			var err error
			if !results[1].IsNil() {
				err = results[1].Interface().(error)
				fmt.Printf("\n+++++++ API_FAILED_HERE ++++++ %s -- req:{%v}, err: %s\n", fullMethod, paramValue.Interface(), err.Error())
			}
			return StdResponse(c, results[0].Interface(), err)
		})
	}
}

// generateRoutePath 从方法名生成路由路径
func generateRoutePath(methodName string) string {
	return "/" + methodName
}

func StdResponse(c *fiber.Ctx, data any, err error) error {
	e := xerr.ToXErr(err)

	var b *anypb.Any
	if data != nil {
		b, _ = anypb.New(data.(proto2.Message))
	}

	msg := e.Detail()
	if e.BizCode() != "" {
		msg = xi18n.Msg(c.UserContext(), e.BizCode(), e.FmtArgs()...)
	}

	buf, _ := (protojson.MarshalOptions{
		EmitDefaultValues: true,
		EmitUnpopulated:   true,
		UseEnumNumbers:    true,
		UseProtoNames:     true,
	}).Marshal(&commpb.StdResp{
		Code:    int32(e.Code()),
		Message: msg,
		Data:    b,
	})
	c.Set("Content-Type", "application/json")
	if e.Code() == xerr.CodeNotFound {
		return c.Status(fiber.StatusNotFound).Send(buf)
	}
	return c.Status(fiber.StatusOK).Send(buf)
}

func PrintUsefulRoutes(app *fiber.App) {
	fmt.Println("method  | path                  | name | handlers")
	fmt.Println("------  | ----                  | ---- | --------")
	for _, route := range app.GetRoutes(true) {
		if shouldSkipRoute(route) {
			continue
		}
		fmt.Printf("%-7s | %-22s | %-4s | %s\n",
			route.Method,
			route.Path,
			route.Name,
			handlerNames(route.Handlers),
		)
	}
}

func shouldSkipRoute(route fiber.Route) bool {
	if len(route.Path) <= 3 {
		return false
	}
	handlers := handlerNames(route.Handlers)
	return strings.Contains(handlers, "main.main.func1") || strings.Contains(handlers, "main.main.Recover.func3")
}

func handlerNames(handlers []fiber.Handler) string {
	names := make([]string, 0, len(handlers))
	for _, h := range handlers {
		fn := runtime.FuncForPC(reflect.ValueOf(h).Pointer())
		if fn == nil {
			continue
		}
		names = append(names, fn.Name())
	}
	return strings.Join(names, " ")
}
