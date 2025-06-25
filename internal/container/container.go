package container

import (
	"fmt"
	"reflect"

	"github.com/archine/gin-plus/v4/component/bean"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/exception"
	"github.com/archine/gin-plus/v4/internal/container/topo"
	"github.com/archine/gin-plus/v4/internal/util"
)

var (
	beanInterfaceType = reflect.TypeOf((*bean.Marker)(nil)).Elem()
	lazeInterfaceType = reflect.TypeOf((*bean.Lazy)(nil)).Elem()
)

type BeanContainer struct {
	refreshed      bool
	beans          map[string]any
	earlyInitBeans map[string]reflect.Type
	typeForNames   map[reflect.Type][]string
}

var container = &BeanContainer{
	refreshed:      false,
	beans:          make(map[string]any),
	earlyInitBeans: make(map[string]reflect.Type),
	typeForNames:   make(map[reflect.Type][]string),
}

func GetBeanContainer() *BeanContainer {
	if !container.refreshed {
		return nil
	}
	return container
}

func RegisterBeanDefinition(bType reflect.Type) {
	if container.refreshed {
		gplog.Fatal(fmt.Sprintf("%+v", exception.NewStackErr("BeanDefinitionErr: register failed, container has been refreshed")))
	}

	if bType.Implements(lazeInterfaceType) || !bType.Implements(beanInterfaceType) {
		// 如果是懒加载bean或不是bean类型，直接返回
		return
	}

	beanName := util.FirstToLower(bType.Name())

	if _, exists := container.earlyInitBeans[beanName]; exists {
		gplog.Fatal(fmt.Sprintf("%+v", exception.NewStackErr("BeanDefinitionErr: bean name already exists: "+beanName)))
	}

	container.earlyInitBeans[beanName] = bType
	container.typeForNames[bType] = append(container.typeForNames[bType], beanName)
}

func Refresh() {
	container.refreshed = true // Mark the container as refreshed

	dependEdges := buildBeanDependencies()
	creationOrder, err := topo.Sort(dependEdges)
	if err != nil {
		gplog.Fatal(fmt.Sprintf("%+v", exception.NewStackErr("BeanCreationErr: Failed to resovle bean dependencies: "+err.Error())))
	}

	for _, beanName := range creationOrder {
		createBean(beanName)
	}
}

// buildBeanDependencies 构建bean依赖关系
func buildBeanDependencies() []*topo.DependencyEdge {
	var edges []*topo.DependencyEdge

	for beanName, def := range container.earlyInitBeans {
		dependencies := getStructDependencies(def)

		// 为每个依赖创建边：beanName依赖dependency
		for _, dependency := range dependencies {
			edge := &topo.DependencyEdge{
				From: beanName,   // 依赖者
				To:   dependency, // 被依赖者
			}
			edges = append(edges, edge)
		}
	}

	return edges
}

// getStructDependencies 获取结构体字段的依赖
func getStructDependencies(structType reflect.Type) []string {
	var dependencies []string

	// 跳过receiver参数，从第1个参数开始
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Anonymous {
			continue
		}

		fieldKind := field.Type.Kind()
		fieldIsInterface := fieldKind == reflect.Interface

		autowire := field.Tag.Get("autowire")

		if !fieldIsInterface && !(fieldKind == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) && autowire != "" {
			gplog.Fatal(fmt.Sprintf("%+v", exception.NewStackErr("BeanCreationErr: Invalid field type for autowire: "+field.Name+" in "+structType.Name())))
		}

		fieldType := field.Type.Elem()

		if names, exists := container.typeForNames[fieldType]; exists {
			dependencies = append(dependencies, names...)
		} else {
			if fieldIsInterface {
				implementingBeans := findBeansImplementingInterface(fieldType)
				if len(implementingBeans) == 0 {
					gplog.Fatal(fmt.Sprintf("%+v", exception.NewStackErr("BeanCreationErr: No bean found for type "+fieldType.String()+" in "+structType.Name())))
				}
				dependencies = append(dependencies, implementingBeans...)
				container.typeForNames[fieldType] = dependencies
			} else {
				gplog.Fatal(fmt.Sprintf("%+v", exception.NewStackErr("BeanCreationErr: No bean found for type "+fieldType.String()+" in "+structType.Name())))
			}
		}
	}

	return dependencies
}

// findBeansImplementingInterface 查找实现了指定接口的bean
func findBeansImplementingInterface(interfaceType reflect.Type) []string {
	var implementingBeans []string

	for beanName, typ := range container.earlyInitBeans {
		if typ.Implements(interfaceType) {
			implementingBeans = append(implementingBeans, beanName)
		}
	}

	return implementingBeans
}

// createBean 创建bean实例
func createBean(beanName string) {
	if _, exists := container.beans[beanName]; exists {
		return // 已创建
	}

	beanTyp := container.earlyInitBeans[beanName]
	structName := beanTyp.Name()

	var instance any
	// 查找构造方法
	constructMethod, exists := beanTyp.MethodByName("New" + structName)
	if exists {
		instance = createBeanWithConstructor(beanName, constructMethod)
	} else {
		instance = reflect.New(beanTyp).Interface()
	}

	container.beans[beanName] = instance

	if postConstruct, ok := instance.(bean.PostConstruct); ok {
		postConstruct.BeanPostConstruct()
	}
}

// createBeanWithConstructor 使用构造器创建bean
func createBeanWithConstructor(beanName string, method reflect.Method) any {
	// todo
	return nil
}
