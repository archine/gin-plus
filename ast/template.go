package ast

var AstTempStr = `package base

// 自动生成,请不要编辑

import "github.com/archine/gin-plus/v2/ast"

var Ast = map[string][]*ast.MethodInfo  {{ with . }} { {{range $key, $element := .}}
    "{{$key}}":{ {{range $element}}
		{Method:"{{.Method}}", ApiPath:"{{.ApiPath}}", GlobalFunc:{{.GlobalFunc}}}, {{end}}
	}, {{end}}
}
{{else}}{}
{{end}}`
