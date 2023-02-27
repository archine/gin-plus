package main

import (
	"fmt"
	"github.com/archine/gin-plus/v2/ast"
	"github.com/archine/gin-plus/v2/exception"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/jennifer/jen"
	log "github.com/sirupsen/logrus"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Parse project controllers and API methods, Run with go generate or mvc command
// by default, the controller directory in the current path is parsed.
// you can set the parsed directory by following the preceding parameter.
// the controller_init.go file is generated in the specified directory. Do not edit it
func main() {
	controllerDir := "controller"
	if len(os.Args) > 1 {
		controllerDir = os.Args[1]
	}
	controllerAbs, err := filepath.Abs(controllerDir)
	if err != nil {
		log.Fatalf("[%s] get controller directory abstract path error, %s", controllerDir, err.Error())
	}
	pkgs, err := decorator.ParseDir(token.NewFileSet(), controllerAbs, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("[%s] parse controller directory error, %s", controllerDir, err.Error())
	}
	var controllerNames []string                      // controller struct name
	apiCache := make(map[string][]*ast.MethodInfo)    // api cache, key is controller name + method name, such as: UserController/AddUser
	basePathEachController := make(map[string]string) // base path for each controller, key indicates the owning controller
	bracketRegex := regexp.MustCompile("[(](.*?)[)]")
	var pkg string // package name
	var p *dst.Package
	for pkg, p = range pkgs {
		for filePath, file := range p.Files {
			dst.Inspect(file, func(node dst.Node) bool {
				switch t := node.(type) {
				case *dst.GenDecl:
					var match bool
					var structType *dst.StructType
					var spec *dst.TypeSpec
					if spec, match = t.Specs[0].(*dst.TypeSpec); !match {
						return false
					}
					if structType, match = spec.Type.(*dst.StructType); !match {
						return false
					}
					if isController(structType.Fields.List) {
						controllerNames = append(controllerNames, spec.Name.Name)
						var prefix string
						for _, comment := range t.Decs.Start {
							comment = removePrefix(comment)
							if strings.HasPrefix(comment, "@BasePath") {
								basePathSub := bracketRegex.FindStringSubmatch(comment)
								if len(basePathSub) == 0 {
									prefix = "/"
									break
								}
								prefix = strings.ReplaceAll(basePathSub[1], "\"", "")
								break
							}
						}
						basePathEachController[spec.Name.Name] = prefix
					}
				case *dst.FuncDecl:
					if t.Decs.Start == nil || t.Recv == nil || t.Name.Name == "PostConstruct" {
						return false
					}
					onwer := searchFather(t.Recv.List) // Which controller does it belong to
					var methods []*ast.MethodInfo
					for _, comment := range t.Decs.Start {
						comment = removePrefix(comment)
						if strings.HasPrefix(comment, "@POST") || strings.HasPrefix(comment, "@GET") ||
							strings.HasPrefix(comment, "@DELETE") || strings.HasPrefix(comment, "@PUT") ||
							strings.HasPrefix(comment, "@PATCH") || strings.HasPrefix(comment, "@OPTIONS") ||
							strings.HasPrefix(comment, "@HEAD") {

							if unicode.IsLower(rune(t.Name.Name[0])) {
								log.Fatalf("[%s] %s: invalid method name. name first word must be uppercase", filePath, t.Name.Name)
							}
							submatch := bracketRegex.FindStringSubmatch(comment)
							if len(submatch) == 0 {
								log.Fatalf("[%s] %s: invalid api definition, example: @GET(path=\"/test\")", filePath, t.Name.Name)
							}
							m := ast.MethodInfo{
								Method: comment[1:strings.Index(comment, "(")],
							}
							apiDefine := strings.Split(submatch[1], ",")
							if strings.HasPrefix(apiDefine[0], "path=") {
								m.ApiPath = path.Join(basePathEachController[onwer], strings.ReplaceAll(apiDefine[0][5:], "\"", ""))
							} else {
								log.Fatalf("[%s] %s invalid path parameter. Must start with path=", filePath, t.Name.Name)
							}
							methods = append(methods, &m)
						}
					}
					if len(methods) > 0 {
						apiCache[onwer+"/"+t.Name.Name] = methods
					}
				}
				return true
			})
		}
	}
	recordProjectControllerAndApi(controllerNames, controllerAbs, pkg, apiCache)
}

// Determines whether the current structure is a controller
func isController(fields []*dst.Field) bool {
	var ok bool
	var selectorExpr *dst.SelectorExpr
	for _, field := range fields {
		selectorExpr, ok = field.Type.(*dst.SelectorExpr)
		if !ok {
			continue
		}
		x := selectorExpr.X.(*dst.Ident)
		sel := selectorExpr.Sel
		if x.Name == "mvc" && sel.Name == "Controller" {
			ok = true
			break
		}
	}
	return ok
}

// Query the controller to which the method belongs
func searchFather(fields []*dst.Field) string {
	for _, field := range fields {
		if f, ok := field.Type.(*dst.StarExpr); ok {
			return f.X.(*dst.Ident).Name
		}
	}
	return ""
}

// Remove comment prefixes such as // @BasePath("/") -> @BasePath("/")
func removePrefix(text string) string {
	text = strings.ReplaceAll(text, " ", "")
	if strings.HasPrefix(text, "//") {
		return text[2:]
	}
	return text
}

// All controller information and Api information for the current project is recorded here
func recordProjectControllerAndApi(controllerNames []string, controllerAbs, pkg string, apiCache map[string][]*ast.MethodInfo) {
	if len(controllerNames) == 0 {
		return
	}
	newFile := jen.NewFile(pkg)
	newFile.HeaderComment("// ⚠️⛔ Auto generate code by hj-gin, Do not edit!!!")
	newFile.HeaderComment("// All controller information and Api information for the current project is recorded here\n")
	newFile.ImportName("github.com/archine/gin-plus//hj-gin/v2/mvc", "mvc")
	newFile.ImportName("github.com/archine/gin-plus//hj-gin/v2/ast", "ast")
	var registerCode []jen.Code
	for _, controllerName := range controllerNames {
		registerCode = append(registerCode, jen.Id(fmt.Sprintf("&%s{}", controllerName)))
	}
	newFile.Func().Id("init").Params().Block(
		jen.Qual("github.com/archine/gin-plus//hj-gin/v2/mvc", "Register").Call(registerCode...),
		jen.Qual("github.com/archine/gin-plus//hj-gin/v2/ast", "Apis").Op("=").Map(jen.String()).Index().Op("*").
			Qual("github.com/archine/gin-plus//hj-gin/v2/ast", "MethodInfo").
			Values(jen.DictFunc(func(dict jen.Dict) {
				for k, methodInfos := range apiCache {
					dict[jen.Lit(k)] = jen.BlockFunc(func(group *jen.Group) {
						for _, info := range methodInfos {
							group.Add(jen.Id(fmt.Sprintf("{Method:%s, ApiPath: %s},", strconv.Quote(info.Method), strconv.Quote(info.ApiPath))))
						}
					})
				}
			})),
	)
	exception.OrThrow(newFile.Save(filepath.Join(controllerAbs, "controller_init.go")))
}
