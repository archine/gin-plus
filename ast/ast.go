package ast

import (
	"fmt"
	"github.com/archine/gin-plus/v2/exception"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
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

func Parse(dir string) {
	p, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	info := parseApiInfo(p)
	tmpl, err := template.New("api").Parse(AstTempStr)
	exception.OrThrow(err)
	f, _ := os.Create("base/template.go")
	exception.OrThrow(tmpl.Execute(f, info))
}

func parseApiInfo(dir string) map[string][]*MethodInfo {
	pkgs, err := parser.ParseDir(token.NewFileSet(), dir, func(info fs.FileInfo) bool {
		return !info.IsDir()
	}, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	data := make(map[string][]*MethodInfo)
	prefix := ""
	for _, v := range pkgs {
		for _, file := range v.Files {
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
							break
						}
						continue
					}
				case *ast.FuncDecl:
					if t.Doc == nil || t.Name.Name == "CallBefore" || t.Name.Name == "PostConstruct" {
						continue
					}
					if _, ok := data[t.Name.Name]; ok {
						panic("Duplicate method name: " + t.Name.Name)
					}
					var methods []*MethodInfo
					for _, comment := range t.Doc.List {
						comment.Text = removePrefix(comment.Text)
						if strings.HasPrefix(comment.Text, "@") {
							m := &MethodInfo{
								Method:     comment.Text[1:strings.Index(comment.Text, "(")],
								GlobalFunc: true,
							}
							comment.Text = comment.Text[strings.Index(comment.Text, "(")+1 : strings.Index(comment.Text, ")")]
							arr := strings.Split(comment.Text, ",")
							for _, s := range arr {
								if strings.HasPrefix(s, "path=") {
									m.ApiPath = path.Join(prefix, strings.ReplaceAll(s[5:], "\"", ""))
									continue
								}
								if strings.HasPrefix(s, "globalFunc=") {
									globalFunc := s[11:]
									if globalFunc == "true" {
										m.GlobalFunc = true
									} else if globalFunc == "false" {
										m.GlobalFunc = false
									} else {
										panic(fmt.Sprintf("globalFunc accepts values of true or false: %s", t.Name.Name))
									}
								}
							}
							methods = append(methods, m)
						}
					}
					data[t.Name.Name] = methods
				}
			}
		}
	}
	return data
}

func removePrefix(text string) string {
	text = strings.ReplaceAll(text, " ", "")
	if strings.HasPrefix(text, "//") {
		return text[2:]
	}
	return text
}
