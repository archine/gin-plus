package routing

import "github.com/gin-gonic/gin"

// annotationCache stores annotations for each API endpoint.
// Key: API full path (e.g., "/api/v1/users/:id")
// Value: map of annotation name to annotation value
var annotationCache map[string]map[string]string

// GetAnnotation retrieves a specific annotation value from the current API endpoint.
// Annotations are typically defined using struct tags or comments on controller methods
// and are used for configuration purposes like validation, documentation, or middleware behavior.
//
// Parameters:
//   - ctx: The Gin context containing the current request information
//   - Name: The name of the annotation to retrieve
//
// Returns:
//   - val: The annotation value if found
//   - has: Boolean indicating whether the annotation exists for this endpoint
//
// Example:
//
//	// For an API method with annotation @RateLimit -> 100/min
//	if limit, exists := GetAnnotation(ctx, "@RateLimit"); exists {
//	    // Use the rate limit value: "100/min"
//	}
func GetAnnotation(ctx *gin.Context, name string) (string, bool) {
	if ctx == nil || name == "" {
		return "", false
	}

	fullPath := ctx.FullPath()
	if fullPath == "" {
		return "", false
	}

	annotations, exists := annotationCache[fullPath]
	if !exists {
		return "", false
	}

	value, found := annotations[name]
	return value, found
}

// GetAnnotationWithDefault retrieves annotation value with a default fallback
func GetAnnotationWithDefault(ctx *gin.Context, name, defaultValue string) string {
	if value, exists := GetAnnotation(ctx, name); exists {
		return value
	}
	return defaultValue
}

// HasAnnotation checks if an annotation exists for the current endpoint
func HasAnnotation(ctx *gin.Context, name string) bool {
	_, exists := GetAnnotation(ctx, name)
	return exists
}
