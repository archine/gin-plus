package ioc

import (
	"errors"
	"fmt"
	"reflect"
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

		// 如果字段不是接口类型 且 不是指针型的结构体则跳过
		if field.Type.Kind() != reflect.Interface && !(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
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
				created, err := createBean(field, factory)
				if err != nil {
					return err
				}
				if created != nil {
					fieldVal.Set(reflect.ValueOf(created))
					continue
				}
			}
			return fmt.Errorf("bean '%s' not found for field %s.%s", autowire, elem.Type().Name(), field.Name)
		}
	}
	return nil
}

// createBean 创建并注入 bean
func createBean(filed reflect.StructField, factory FactoryBean) (any, error) {
	beanName := factory.GetBeanName()
	if beanName == "" {
		beanName = filed.Type.Name()
	}

	mutex.Lock()
	if bean, ok := beanCache[beanName]; ok {
		mutex.Unlock()
		return bean, nil
	}

	newBean := factory.CreateBean()
	if newBean == nil {
		mutex.Unlock()
		return fmt.Errorf("对象 %s 尝试创建Bean失败，返回值为空", filed.Type.String()), nil
	}

	beanCache[beanName] = newBean
	mutex.Unlock()

	// 递归注入新创建的实例
	if err := Inject(newBean); err != nil {
		mutex.Lock()
		delete(beanCache, beanName)
		mutex.Unlock()
		return nil, nil
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
