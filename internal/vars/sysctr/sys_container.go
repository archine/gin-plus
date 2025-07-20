package sysctr

import (
	"github.com/archine/gin-plus/v4/internal/container"
)

var (
	Container = container.NewContainer() // Container is the global IOC container
)
