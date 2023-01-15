package ast

import (
	"fmt"
	"github.com/archine/gin-plus/v2/exception"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type MethodInfo struct {
	Method     string // API method。such as: POST、GET、DELETE、PUT、OPTIONS、PATCH、HEAD
	ApiPath    string // API path
	GlobalFunc bool   // Whether load global func
}

// Parse parse apis in multiple paths
func Parse(dirs ...string) {
	if dirs == nil {
		dirs = append(dirs, "controller")
	}
	var info map[string][]*MethodInfo
	for i, dir := range dirs {
		absPath, err := filepath.Abs(dir)
		exception.OrThrow(err)
		info = parseApiInfo(absPath)
		if i == 0 {
			process("base/template.go", AstTempStr, info)
		} else {
			AstTempStr = strings.Replace(AstTempStr, "Ast", fmt.Sprintf("Ast%d", i), 1)
			process(fmt.Sprintf("base/template%d.go", i), AstTempStr, info)
		}
	}
}

func parseApiInfo(dir string) map[string][]*MethodInfo {
	pkgs, err := parser.ParseDir(token.NewFileSet(), dir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	data := make(map[string][]*MethodInfo)
	prefix := ""
	for _, v := range pkgs {
		for fileName, file := range v.Files {
			prefix = ""
			for _, decl := range file.Decls {
				switch t := decl.(type) {
				case *ast.GenDecl:
					if t.Doc == nil || t.Tok.String() != "type" {
						continue
					}
					for _, comment := range t.Doc.List {
						comment.Text = removePrefix(comment.Text)
						if strings.HasPrefix(comment.Text, "@BasePath") {
							prefix = comment.Text[10:strings.LastIndex(comment.Text, ")")]
							prefix = strings.ReplaceAll(prefix, "\"", "")
							if !strings.HasPrefix(prefix, "/") {
								prefix = "/" + prefix
							}
							break
						}
						continue
					}
				case *ast.FuncDecl:
					if t.Doc == nil || len(t.Recv.List) == 0 || t.Name.Name == "PostConstruct" || t.Name.Name == "CallBefore" {
						continue
					}
					var parent string // Which controller does it belong to
					for _, f := range t.Recv.List {
						switch ft := f.Type.(type) {
						case *ast.StarExpr:
							parent = ft.X.(*ast.Ident).Name
						}
					}
					var methods []*MethodInfo
					for _, comment := range t.Doc.List {
						comment.Text = removePrefix(comment.Text)
						if strings.HasPrefix(comment.Text, "@POST") ||
							strings.HasPrefix(comment.Text, "@GET") ||
							strings.HasPrefix(comment.Text, "@DELETE") ||
							strings.HasPrefix(comment.Text, "@PUT") ||
							strings.HasPrefix(comment.Text, "@PATCH") ||
							strings.HasPrefix(comment.Text, "@OPTIONS") ||
							strings.HasPrefix(comment.Text, "@HEAD") {

							startIndex := strings.Index(comment.Text, "(")
							endIndex := strings.Index(comment.Text, ")")
							if startIndex == -1 || endIndex == -1 {
								log.Fatalf("[%s] %s: invalid api definition, example: @GET(path=\"/test\", globalFunc=true)", fileName, t.Name.Name)
							}
							m := &MethodInfo{
								Method:     comment.Text[1:strings.Index(comment.Text, "(")],
								GlobalFunc: true,
							}

							comment.Text = comment.Text[startIndex+1 : endIndex]
							arr := strings.Split(comment.Text, ",")
							if strings.HasPrefix(arr[0], "path=") {
								m.ApiPath = path.Join(prefix, strings.ReplaceAll(arr[0][5:], "\"", ""))
							} else {
								log.Fatalf("[%s] %s invalid path parameter. Must start with path=", fileName, t.Name.Name)
							}
							if len(arr) > 1 {
								if strings.HasPrefix(arr[1], "globalFunc=") {
									globalFunc := arr[1][11:]
									if globalFunc == "true" {
										m.GlobalFunc = true
									} else if globalFunc == "false" {
										m.GlobalFunc = false
									} else {
										log.Fatalf("[%s] %s invalid global func value, accept only true or false", fileName, t.Name.Name)
									}
								}
							}
							methods = append(methods, m)
						}
					}
					data[parent+"/"+t.Name.Name] = methods
				}
			}
		}
	}
	return data
}

func process(fileName, astTemplate string, apiInfo map[string][]*MethodInfo) {
	tmpl, err := template.New("api").Parse(astTemplate)
	f, err := os.Create(fileName)
	defer f.Close()
	exception.OrThrow(err)
	exception.OrThrow(tmpl.Execute(f, apiInfo))
}

func removePrefix(text string) string {
	text = strings.ReplaceAll(text, " ", "")
	if strings.HasPrefix(text, "//") {
		return text[2:]
	}
	return text
}
