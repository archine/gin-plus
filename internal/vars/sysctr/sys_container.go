package sysctr

import (
	"github.com/archine/gin-plus/v4/internal/container"
)

var (
	BeanRegistry = container.NewBeanDefinitionRegistry() // BeanRegistry stores candidate bean definitions before refresh
	Container    = container.NewContainer()              // Container is the global IOC container
)
