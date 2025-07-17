package syscontainer

import (
	"github.com/archine/gin-plus/v4/internal/container"
)

var (
	Container *container.Container // Container is the global IOC container, initialized in main.go
)
