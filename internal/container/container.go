package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/bean"
	"github.com/archine/gin-plus/v4/exception"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
)

var (
	refreshedFlag     atomic.Bool
	mutex             = sync.RWMutex{}
	beanCache         = make(map[string]*Definition)
	beanNamesForTypes = make(map[reflect.Type][]string)
	factoryBeanType   = reflect.TypeOf((*bean.AbstractBean)(nil)).Elem()
)

func SetDefinition(instance any) {
	if refreshedFlag.Load() {
		log.Fatalf("%+v", exception.NewStackErr("RegisterBeanError: container has been refreshed, cannot set new bean definition"))
	}

	ityp := reflect.TypeOf(instance)
	if ityp.Kind() != reflect.Ptr {
		log.Fatalf("%+v", exception.NewStackErr("RegisterBeanError: instance must be a pointer to a struct"))
	}
	ityp = ityp.Elem()
	if ityp.Kind() != reflect.Struct {
		log.Fatalf("%+v", exception.NewStackErr("RegisterBeanError: instance must be a pointer to a struct"))
	}

	var beanName string
	if b, ok := instance.(bean.AbstractBean); ok {
		beanName = b.BeanName()
	}
	if beanName == "" {
		beanName = ityp.Name()
		beanName = string(beanName[0]|32) + beanName[1:] // Convert first letter to lowercase
	}

	mutex.Lock()
	defer mutex.Unlock()
	if _, exists := beanCache[beanName]; exists {
		log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("RegisterBeanError: the beans '%s' definition already exists", beanName)))
	}

	definition := &Definition{ityp, instance}
	beanCache[beanName] = definition
	if _, exists := beanNamesForTypes[ityp]; !exists {
		beanNamesForTypes[ityp] = []string{beanName}
	} else {
		beanNamesForTypes[ityp] = append(beanNamesForTypes[ityp], beanName)
	}
}

func Refresh() {
	for name, def := range beanCache {

		for i := 0; i < def.Type.NumField(); i++ {
			filed := def.Type.Field(i)
			// Skip anonymous/embedded fields as they are difficult to handle properly
			if filed.Anonymous {
				continue
			}

			autowire := filed.Tag.Get("autowire")

			// Get autowire tag value, skip if not present.
			// This tag indicates that the field should be injected with a bean from the container.
			// The tag value is the name of the bean to be searched for. If it is "-",
			// it indicates lookup by filed name, if not found, it will be by type.
			//
			// Examples: `autowire:"myBean"` or `autowire:"-"`.
			if autowire == "" {
				continue
			}
		}
	}
}
