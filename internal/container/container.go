package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/exception"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
)

var (
	refreshedFlag     atomic.Bool
	mutex             = sync.RWMutex{}
	beanCache         = make(map[string]*Definition)    // Complete IoC container for managing beans
	beanNamesForTypes = make(map[reflect.Type][]string) // Map of bean names by their type
	//factoryBeanType   = reflect.TypeOf((*FactoryBean)(nil)).Elem()
)

func SetDefinition(beanName string, beanType reflect.Type) {
	if refreshedFlag.Load() {
		log.Fatalf("%+v", exception.NewStackErr("container has been refreshed, cannot set new bean definition"))
	}
	mutex.Lock()
	defer mutex.Unlock()
	if _, exists := beanCache[beanName]; exists {
		log.Fatalf("%+v", exception.NewStackErr(fmt.Sprintf("bean %s already exists", beanName)))
	}

	beanCache[beanName] = &Definition{Type: beanType}
	beanNamesForTypes[beanType] = append(beanNamesForTypes[beanType], beanName)
}

func Refresh() {

}
