package factory

import (
	"github.com/archine/gin-plus/v4/exception"
	"github.com/archine/gin-plus/v4/internal/container"
	"log"
	"reflect"
)

func RegisterBeanDefinition(beanName string, beanType reflect.Type) {
	if beanType == nil {
		log.Fatalf("%+v", exception.NewStackErr("[RegisterBeanDefinition] type is nil"))
	}
	if beanType.Kind() == reflect.Ptr {
		beanType = beanType.Elem()
	}
	if beanName == "" {
		beanName = beanType.Name()
	}

	container.SetDefinition(beanName, beanType)
}
