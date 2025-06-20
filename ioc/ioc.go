package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	mutex = sync.RWMutex{}
)

// Inject 注入对象的所有依赖字段
func Inject(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("must be a non-nil pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("must be a pointer to struct")
	}

	// 先缓存当前对象，避免循环依赖
	beanName := strings.ToLower(elem.Type().Name()[:1]) + elem.Type().Name()[1:]
	mutex.Lock()
	if _, exists := beanCache[beanName]; !exists {
		beanCache[beanName] = v
	}
	mutex.Unlock()

	typ := elem.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		autowire := field.Tag.Get("autowire")
		if autowire == "" {
			continue
		}

		fieldVal := elem.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		// 支持指针类型和接口类型
		if field.Type.Kind() != reflect.Ptr && field.Type.Kind() != reflect.Interface {
			continue
		}

		mutex.RLock()
		bean, exists := beanCache[autowire]
		mutex.RUnlock()

		if exists {
			beanVal := reflect.ValueOf(bean)
			if field.Type.Kind() == reflect.Interface {
				if beanVal.Type().Implements(field.Type) {
					fieldVal.Set(beanVal)
				} else {
					return fmt.Errorf("bean %s does not implement interface %s", autowire, field.Type.String())
				}
			} else if beanVal.Type().AssignableTo(field.Type) {
				fieldVal.Set(beanVal)
			} else {
				return fmt.Errorf("bean %s of type %s cannot be assigned to field %s of type %s",
					autowire, beanVal.Type().String(), field.Name, field.Type.String())
			}
		} else {
			if factory, ok := fieldVal.Interface().(FactoryBean); ok {
				created, err := createBean(autowire, factory)
				if err != nil {
					return err
				}
				if created != nil {
					fieldVal.Set(reflect.ValueOf(created))
					continue
				}
			}
			return fmt.Errorf("bean %s not found for field %s", autowire, field.Name)
		}
	}
	return nil
}

// createBean 创建并注入 bean
func createBean(beanName string, factory FactoryBean) (any, error) {
	if beanName == "" {
		// 使用结构体名称，首字母小写
		typ := reflect.TypeOf(factory)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		structName := typ.Name()
		if structName != "" {
			beanName = strings.ToLower(structName[:1]) + structName[1:]
		}
	}

	mutex.Lock()
	if bean, ok := beanCache[beanName]; ok {
		mutex.Unlock()
		return bean, nil
	}

	newBean := factory.CreateBean()
	if newBean == nil {
		mutex.Unlock()
		return nil, fmt.Errorf("factory failed to create bean: %s", beanName)
	}

	beanCache[beanName] = newBean
	mutex.Unlock()

	// 递归注入新创建的实例
	if err := Inject(newBean); err != nil {
		mutex.Lock()
		delete(beanCache, beanName)
		mutex.Unlock()
		return nil, err
	}

	return newBean, nil
}

// SetBean 手动设置 bean 到 IoC 容器中
// 主要用于在运行时动态添加 bean，非实现 FactoryBean 接口的对象
func SetBean(beanName string, bean any) error {
	if beanName == "" || bean == nil {
		return errors.New("beanName cannot be empty and bean cannot be nil")
	}
	beanTyp := reflect.TypeOf(bean)

	if beanTyp.Kind() != reflect.Ptr {
		return errors.New("bean must be a non-nil pointer")
	}

	mutex.RLock()
	if _, exists := beanCache[beanName]; exists {
		return nil
	}
	mutex.RUnlock()

	mutex.Lock()
	beanCache[beanName] = bean
	mutex.Unlock()

	return nil
}

// GetBean 根据类型获取 bean
func GetBean(beanStruct any) any {
	if beanStruct == nil {
		return nil
	}

	typ := reflect.TypeOf(beanStruct)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	mutex.RLock()
	bean := beanCache[typ.Name()]
	mutex.RUnlock()
	return bean
}

// GetBeanByName 根据名称获取 bean
func GetBeanByName(beanName string) any {
	if beanName == "" {
		return nil
	}

	mutex.RLock()
	bean := beanCache[beanName]
	mutex.RUnlock()
	return bean
}
