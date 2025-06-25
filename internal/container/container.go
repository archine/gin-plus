package container

// import (
// 	"fmt"
// 	"github.com/archine/gin-plus/v4/component/bean"
// 	"github.com/archine/gin-plus/v4/exception"
// 	"log"
// 	"reflect"
// 	"sync"
// 	"sync/atomic"
// )

// var (
// 	refreshedFlag     atomic.Bool
// 	mutex             = sync.RWMutex{}
// 	beanCache         = make(map[string]*Definition)
// 	beanNamesForTypes = make(map[reflect.Type][]string)
// )

// func SetDefinition(instance any) {
// 	if refreshedFlag.Load() {
// 		log.Fatalf("%+v", exception.NewStackErr("BeanDefinitionError: container has been refreshed, cannot set new bean definition"))
// 	}

// 	ityp := reflect.TypeOf(instance)
// 	if ityp.Kind() != reflect.Ptr {
// 		log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("BeanDefinitionError: instance '%s' must be a pointer to a struct", ityp.Name())))
// 	}
// 	ityp = ityp.Elem()
// 	if ityp.Kind() != reflect.Struct {
// 		log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("BeanDefinitionError: instance '%s' must be a pointer to a struct", ityp.Name())))
// 	}

// 	var beanName string
// 	if b, ok := instance.(bean.AbstractBean); ok {
// 		beanName = b.BeanName()
// 	}
// 	if beanName == "" {
// 		beanName = ityp.Name()
// 		beanName = string(beanName[0]|32) + beanName[1:] // Convert first letter to lowercase
// 	}

// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	if _, exists := beanCache[beanName]; exists {
// 		log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("BeanDefinitionError: the beans '%s' definition already exists", beanName)))
// 	}

// 	definition := &Definition{ityp, instance}
// 	beanCache[beanName] = definition
// 	if _, exists := beanNamesForTypes[ityp]; !exists {
// 		beanNamesForTypes[ityp] = []string{beanName}
// 	} else {
// 		beanNamesForTypes[ityp] = append(beanNamesForTypes[ityp], beanName)
// 	}
// }

// func Refresh() {
// 	refreshedFlag.Store(true) // Set the flag to forbid further bean definitions

// 	var fieldKind reflect.Kind
// 	var fieldType reflect.Type
// 	var fieldIsInterface bool

// 	for name, def := range beanCache {
// 		beanVal := reflect.ValueOf(def.Bean).Elem()

// 		for i := 0; i < def.Type.NumField(); i++ {
// 			field := def.Type.Field(i)
// 			// Skip anonymous/embedded fields as they are difficult to handle properly
// 			if field.Anonymous {
// 				continue
// 			}

// 			autowire := field.Tag.Get("autowire")

// 			// Get autowire tag value, skip if not present.
// 			// This tag indicates that the field should be injected with a bean from the container.
// 			// The tag value is the name of the bean to be searched for. If it is "-",
// 			// it indicates lookup by filed name, if not found, it will be by type.
// 			//
// 			// Examples: `autowire:"myBean"` or `autowire:"-"`.
// 			if autowire == "" {
// 				continue
// 			}

// 			fieldKind = field.Type.Kind()
// 			fieldIsInterface = fieldKind == reflect.Interface

// 			if !fieldIsInterface && !(fieldKind == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
// 				// Skip fields that are not interfaces or pointers to structs
// 				continue
// 			}

// 			fieldType = field.Type.Elem()

// 			fieldVal := beanVal.Field(i)
// 			if !fieldVal.IsNil() {
// 				// If the field is already set, skip it
// 				continue
// 			}

// 			var injectBean any
// 			if fieldIsInterface {
// 				injectBean = processInterface(autowire, fieldType)
// 			} else {
// 				injectBean = processObj(autowire, fieldType)
// 			}

// 			if injectBean == nil {
// 				log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("RefreshContextError: the bean '%s' not found for field '%s' in bean '%s'", autowire, field.Name, name)))
// 			}

// 			// Set the field value to the injected bean
// 			if fieldVal.CanSet() {
// 				fieldVal.Set(reflect.ValueOf(injectBean))
// 			}
// 		}
// 	}
// }

// func processObj(autowire string, fieldType reflect.Type) any {
// 	if autowire == "-" {
// 		autowire = fieldType.Name()
// 		autowire = string(autowire[0]|32) + autowire[1:]

// 		if targetBeanDef, exist := beanCache[autowire]; exist {
// 			if targetBeanDef.Type != fieldType {
// 				log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("RefreshContextError: the bean '%s' does not match the type '%s'", targetBeanDef.Type.Name(), fieldType.Name())))
// 			}
// 			return targetBeanDef.Bean
// 		}

// 		if beanNames, ok := beanNamesForTypes[fieldType]; ok && len(beanNames) > 0 {
// 			return beanCache[beanNames[0]].Bean
// 		} else {
// 			// 缓存中没有找到对应的bean，遍历缓存，直到找到第一个符合的
// 			for n, d := range beanCache {
// 				if d.Type == fieldType {
// 					beanNamesForTypes[fieldType] = append(beanNamesForTypes[fieldType], n)
// 					return d.Bean
// 				}
// 			}
// 		}
// 	} else {
// 		if targetBeanDef, exist := beanCache[autowire]; exist {
// 			if targetBeanDef.Type != fieldType {
// 				log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("RefreshContextError: the bean '%s' does not match the type '%s'", targetBeanDef.Type.Name(), fieldType.Name())))
// 			}
// 			return targetBeanDef.Bean
// 		}
// 	}
// 	return nil
// }

// func processInterface(autowire string, fieldType reflect.Type) any {
// 	if autowire == "-" {
// 		autowire = fieldType.Name()
// 		autowire = string(autowire[0]|32) + autowire[1:]

// 		if targetBeanDef, exist := beanCache[autowire]; exist {
// 			if !targetBeanDef.Type.Implements(fieldType) {
// 				log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("RefreshContextError: the bean '%s' does not implement the interface '%s'", targetBeanDef.Type.Name(), fieldType.Name())))
// 			}
// 			return targetBeanDef.Bean
// 		}
// 		// If not found by name, try to find by type
// 		if beanNames, ok := beanNamesForTypes[fieldType]; ok && len(beanNames) > 0 {
// 			return beanCache[beanNames[0]].Bean
// 		}
// 		// 缓存中没有找到对应的bean，遍历缓存，直到找到第一个符合的
// 		for n, d := range beanCache {
// 			if d.Type.Implements(fieldType) {
// 				beanNamesForTypes[fieldType] = append(beanNamesForTypes[fieldType], n)
// 				return d.Bean
// 			}
// 		}
// 	} else {
// 		if targetBeanDef, exist := beanCache[autowire]; exist {
// 			if !targetBeanDef.Type.Implements(fieldType) {
// 				log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("RefreshContextError: the bean '%s' does not implement the interface '%s'", targetBeanDef.Type.Name(), fieldType.Name())))
// 			}
// 			return targetBeanDef.Bean
// 		}
// 	}

// 	return nil
// }
