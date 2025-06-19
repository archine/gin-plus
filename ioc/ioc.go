package ioc

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	beanCache = make(map[string]any)
	mutex     = sync.RWMutex{}
)

// Bean 定义可被 IoC 容器管理的 bean 接口
type Bean interface {
	CreateBean() Bean
}

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
	beanName := elem.Type().String()
	mutex.Lock()
	if _, exists := beanCache[beanName]; !exists {
		beanCache[beanName] = v
	}
	mutex.Unlock()

	// 遍历字段进行注入
	typ := elem.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldVal := elem.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		if field.Type.Kind() == reflect.Interface {
			injectInterface(fieldVal, field)
		} else if field.Type.Kind() == reflect.Ptr {
			injectPointer(fieldVal, field)
		}
	}
	return nil
}

// injectInterface 注入接口类型字段
func injectInterface(fieldVal reflect.Value, field reflect.StructField) {
	primary := field.Tag.Get("autowire")
	mutex.RLock()
	defer mutex.RUnlock()

	for _, bean := range beanCache {
		if reflect.TypeOf(bean).Implements(field.Type) {
			if primary == "" || reflect.TypeOf(bean).Elem().String() == primary {
				fieldVal.Set(reflect.ValueOf(bean))
				break
			}
		}
	}
}

// injectPointer 注入指针类型字段
func injectPointer(fieldVal reflect.Value, field reflect.StructField) {
	fieldBeanName := field.Type.Elem().String()

	mutex.RLock()
	bean, exists := beanCache[fieldBeanName]
	mutex.RUnlock()

	if exists {
		fieldVal.Set(reflect.ValueOf(bean))
		return
	}

	// 尝试通过 Bean 接口创建实例
	if field.Type.Implements(reflect.TypeOf((*Bean)(nil)).Elem()) && !fieldVal.IsNil() {
		if factory, ok := fieldVal.Interface().(Bean); ok {
			if created := createBean(fieldBeanName, factory); created != nil {
				fieldVal.Set(reflect.ValueOf(created))
			}
		}
	}
}

// createBean 创建并注入 bean
func createBean(beanName string, factory Bean) any {
	mutex.Lock()
	if bean, ok := beanCache[beanName]; ok {
		mutex.Unlock()
		return bean
	}

	instance := factory.CreateBean()
	if instance == nil {
		mutex.Unlock()
		return nil
	}

	beanCache[beanName] = instance
	mutex.Unlock()

	// 递归注入新创建的实例
	if err := Inject(instance); err != nil {
		mutex.Lock()
		delete(beanCache, beanName)
		mutex.Unlock()
		return nil
	}

	return instance
}

// SetBeans 批量设置 bean 到容器
func SetBeans(beans ...any) error {
	mutex.Lock()
	defer mutex.Unlock()

	for i, bean := range beans {
		if bean == nil || reflect.TypeOf(bean).Kind() != reflect.Ptr {
			return fmt.Errorf("bean at index %d must be a non-nil pointer", i)
		}
		beanName := reflect.TypeOf(bean).Elem().String()
		beanCache[beanName] = bean
	}
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
	bean := beanCache[typ.String()]
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

// GetAllBeans 获取所有 bean
func GetAllBeans() map[string]any {
	mutex.RLock()
	defer mutex.RUnlock()

	result := make(map[string]any, len(beanCache))
	for k, v := range beanCache {
		result[k] = v
	}
	return result
}

// ContainsBean 检查是否存在指定 bean
func ContainsBean(beanName string) bool {
	mutex.RLock()
	_, exists := beanCache[beanName]
	mutex.RUnlock()
	return exists
}

// RemoveBean 移除指定 bean
func RemoveBean(beanName string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := beanCache[beanName]; exists {
		delete(beanCache, beanName)
		return true
	}
	return false
}

// ClearBeans 清空所有 bean
func ClearBeans() {
	mutex.Lock()
	defer mutex.Unlock()
	beanCache = make(map[string]any)
}
