package factory

import (
	"github.com/archine/gin-plus/v4/exception"
	"github.com/archine/gin-plus/v4/internal/container"
	"log"
)

func RegisterBeanDefinition(instance any) {
	if instance == nil {
		log.Fatalf("%+v", exception.NewStackErr("RegisterBeanError: bean instance is nil"))
	}
	container.SetDefinition(instance)
}
