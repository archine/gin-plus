package ctrl

import "github.com/archine/gin-plus/v4/component/bean"

// Controller is a base struct that declares an entity as a controller.
// API methods can be added to this struct.
type Controller struct {
	bean.Component
}
