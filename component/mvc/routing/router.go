package routing

import (
	"github.com/archine/gin-plus/v4/internal/event_manager"
	"github.com/gin-gonic/gin"
)

// Apply attaches all APIs to the Gin engine.
// It uses reflection to dynamically bind methods to the Gin engine based on the API definitions
// provided by the ast_base.Result.Apis map.
// It also injects dependencies into the controllers
//
// Args:
//
//	engine: The Gin engine to which the APIs will be attached.
//	eventManager: The event manager to handle app events.
func Apply(engine *gin.Engine, eventManager *event_manager.AppEventManager) {

}
