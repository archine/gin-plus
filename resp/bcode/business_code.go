package bcode

// Response business code enumeration
const (
	BadRequest      = 40000 // Bad Request
	NonLogin        = 40100 // Not logged in
	TokenExpired    = 40101 // Token expired
	Forbidden       = 40300 // Forbidden
	ParamValidation = 40001 // Parameter error
	SystemError     = 50000 // Server exception
)
