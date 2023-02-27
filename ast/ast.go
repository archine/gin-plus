package ast

// Apis Project global API cache
// Method injection into gin removes the corresponding method cache
var Apis map[string][]*MethodInfo

// MethodInfo Api method info
type MethodInfo struct {
	Method  string // API method。such as: POST、GET、DELETE、PUT、OPTIONS、PATCH、HEAD
	ApiPath string // API path
}
