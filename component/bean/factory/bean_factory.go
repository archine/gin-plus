package factory

import (
	"fmt"
	"log"
	"reflect"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/exception"
	"github.com/archine/gin-plus/v4/internal/container"
)

func RegisterBeanDefinition(bType reflect.Type) {
	if bType == nil {
		gplog.Fatal(fmt.Sprintf("%+v", exception.NewStackErr("BeanDefinitionErr: register failed, bType is nil")))
	}
	if bType.Kind() == reflect.Ptr {
		bType = bType.Elem()
	}
	if bType.Kind() != reflect.Struct {
		gplog.Fatal(fmt.Sprintf("%+v", exception.NewStackErr("BeanDefinitionErr: register failed, bType must be a struct type")))
	}

	container.RegisterBeanDefinition(bType)
}
