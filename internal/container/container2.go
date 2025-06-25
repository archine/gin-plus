package container

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/archine/gin-plus/v4/component/bean"
	"github.com/archine/gin-plus/v4/exception"
)

// BeanState represents the state of a bean in the container
type BeanState int

const (
	BeanStateRegistered BeanState = iota // Bean definition registered
	BeanStateCreating                    // Bean is being created (to detect circular dependencies)
	BeanStateCreated                     // Bean instance created but not fully initialized
	BeanStateCompleted                   // Bean fully initialized and ready to use
)

// BeanFactory is a function type for creating bean instances
type BeanFactory func() (any, error)

// BeanContainer manages beans with three-level cache system
type BeanContainer struct {
	// 一级缓存：完成品缓存 - 存储完全初始化完成的Bean实例
	singletonObjects sync.Map // 改用sync.Map减少锁竞争

	// 二级缓存：早期暴露缓存 - 存储已实例化但未完全初始化的Bean实例
	earlySingletonObjects sync.Map

	// 三级缓存：工厂缓存 - 存储Bean的创建工厂函数
	singletonFactories sync.Map

	// Bean定义缓存
	beanDefinitions map[string]*Definition

	// 类型到Bean名称的映射
	beanNamesForTypes map[reflect.Type][]string

	// Bean状态跟踪
	beanStates map[string]BeanState

	// 正在创建的Bean集合，用于检测循环依赖
	singletonsCurrentlyInCreation sync.Map

	// 字段信息缓存，避免重复反射
	fieldInfoCache map[reflect.Type][]*FieldInfo

	// 读写锁 - 只保护需要读写锁的数据结构
	mutex sync.RWMutex

	// 容器是否已刷新
	refreshed atomic.Bool
}

// FieldInfo 缓存字段信息以减少反射开销
type FieldInfo struct {
	Index       int
	Name        string
	Type        reflect.Type
	Autowire    string
	IsInterface bool
	IsPointer   bool
}

var (
	defaultContainer = NewBeanContainer()
)

// NewBeanContainer creates a new bean container
func NewBeanContainer() *BeanContainer {
	return &BeanContainer{
		beanDefinitions:   make(map[string]*Definition),
		beanNamesForTypes: make(map[reflect.Type][]string),
		beanStates:        make(map[string]BeanState),
		fieldInfoCache:    make(map[reflect.Type][]*FieldInfo),
	}
}

// SetDefinition registers a bean definition in the container
func SetDefinition(instance any) {
	defaultContainer.SetDefinition(instance)
}

// Refresh initializes all beans and resolves dependencies
func Refresh() {
	defaultContainer.Refresh()
}

// GetBean retrieves a bean from the container
func GetBean(name string) any {
	return defaultContainer.GetBean(name)
}

// SetDefinition registers a bean definition in the container
func (c *BeanContainer) SetDefinition(instance any) {
	if c.refreshed.Load() {
		panic(exception.NewStackErr("BeanDefinitionError: container has been refreshed, cannot set new bean definition"))
	}

	ityp := reflect.TypeOf(instance)
	if ityp.Kind() != reflect.Ptr {
		panic(exception.NewStackErr(fmt.Sprintf("BeanDefinitionError: instance '%s' must be a pointer to a struct", ityp.Name())))
	}

	elemType := ityp.Elem()
	if elemType.Kind() != reflect.Struct {
		panic(exception.NewStackErr(fmt.Sprintf("BeanDefinitionError: instance '%s' must be a pointer to a struct", elemType.Name())))
	}

	var beanName string
	if b, ok := instance.(bean.AbstractBean); ok {
		beanName = b.BeanName()
	}
	if beanName == "" {
		beanName = elemType.Name()
		if len(beanName) > 0 {
			beanName = string(beanName[0]|32) + beanName[1:] // Convert first letter to lowercase
		}
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.beanDefinitions[beanName]; exists {
		panic(exception.NewStackErr(fmt.Sprintf("BeanDefinitionError: the beans '%s' definition already exists", beanName)))
	}

	definition := &Definition{elemType, instance}
	c.beanDefinitions[beanName] = definition
	c.beanStates[beanName] = BeanStateRegistered

	// 预计算字段信息
	c.cacheFieldInfo(elemType)

	// 更新类型映射
	if names, exists := c.beanNamesForTypes[elemType]; !exists {
		c.beanNamesForTypes[elemType] = []string{beanName}
	} else {
		c.beanNamesForTypes[elemType] = append(names, beanName)
	}
}

// cacheFieldInfo 预计算并缓存字段信息
func (c *BeanContainer) cacheFieldInfo(typ reflect.Type) {
	if _, exists := c.fieldInfoCache[typ]; exists {
		return
	}

	var fields []*FieldInfo
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Anonymous {
			continue
		}

		autowire := field.Tag.Get("autowire")
		if autowire == "" {
			continue
		}

		fieldType := field.Type
		fieldKind := fieldType.Kind()
		isInterface := fieldKind == reflect.Interface
		isPointer := fieldKind == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct

		if !isInterface && !isPointer {
			continue
		}

		fieldInfo := &FieldInfo{
			Index:       i,
			Name:        field.Name,
			Type:        fieldType,
			Autowire:    autowire,
			IsInterface: isInterface,
			IsPointer:   isPointer,
		}
		fields = append(fields, fieldInfo)
	}

	c.fieldInfoCache[typ] = fields
}

// Refresh initializes all beans and resolves dependencies using three-level cache
func (c *BeanContainer) Refresh() {
	c.refreshed.Store(true)

	c.mutex.RLock()
	beanNames := make([]string, 0, len(c.beanDefinitions))
	for name := range c.beanDefinitions {
		beanNames = append(beanNames, name)
	}
	c.mutex.RUnlock()

	// 预实例化所有单例Bean
	for _, beanName := range beanNames {
		c.getSingleton(beanName)
	}
}

// getSingleton implements three-level cache mechanism
func (c *BeanContainer) getSingleton(beanName string) any {
	// 一级缓存：完成品缓存
	if singletonObject, exists := c.singletonObjects.Load(beanName); exists {
		return singletonObject
	}

	// 检查是否正在创建中（循环依赖检测）
	if _, creating := c.singletonsCurrentlyInCreation.Load(beanName); creating {
		// 二级缓存：早期暴露缓存
		if earlySingletonObject, exists := c.earlySingletonObjects.Load(beanName); exists {
			return earlySingletonObject
		}

		// 三级缓存：工厂缓存
		if factoryVal, exists := c.singletonFactories.Load(beanName); exists {
			factory := factoryVal.(BeanFactory)
			earlySingletonObject, err := factory()
			if err != nil {
				panic(exception.NewStackErr(fmt.Sprintf("BeanCreationError: failed to create bean '%s': %v", beanName, err)))
			}
			// 将早期对象放入二级缓存，并从三级缓存移除
			c.earlySingletonObjects.Store(beanName, earlySingletonObject)
			c.singletonFactories.Delete(beanName)
			return earlySingletonObject
		}
	}

	// 如果都没有，则需要创建Bean
	return c.createBean(beanName)
}

// createBean creates a new bean instance and handles dependencies
func (c *BeanContainer) createBean(beanName string) any {
	c.mutex.RLock()
	definition, exists := c.beanDefinitions[beanName]
	c.mutex.RUnlock()

	if !exists {
		panic(exception.NewStackErr(fmt.Sprintf("BeanCreationError: bean definition '%s' not found", beanName)))
	}

	// 标记正在创建
	c.singletonsCurrentlyInCreation.Store(beanName, true)

	c.mutex.Lock()
	c.beanStates[beanName] = BeanStateCreating
	c.mutex.Unlock()

	// 创建Bean实例
	beanInstance := c.instantiateBean(definition)

	// 将工厂函数放入三级缓存
	c.singletonFactories.Store(beanName, BeanFactory(func() (any, error) {
		return c.getEarlyBeanReference(beanName, beanInstance)
	}))

	// 标记为已创建
	c.mutex.Lock()
	c.beanStates[beanName] = BeanStateCreated
	c.mutex.Unlock()

	// 填充属性（依赖注入）
	c.populateBean(beanName, beanInstance, definition)

	// 初始化Bean
	c.initializeBean(beanName, beanInstance)

	// 从创建中移除
	c.singletonsCurrentlyInCreation.Delete(beanName)

	// 移除早期缓存和工厂缓存
	c.earlySingletonObjects.Delete(beanName)
	c.singletonFactories.Delete(beanName)

	// 放入一级缓存
	c.singletonObjects.Store(beanName, beanInstance)

	c.mutex.Lock()
	c.beanStates[beanName] = BeanStateCompleted
	c.mutex.Unlock()

	return beanInstance
}

// instantiateBean creates bean instance using reflection
func (c *BeanContainer) instantiateBean(definition *Definition) any {
	// 由于我们已经有了预创建的实例，直接返回
	return definition.Bean
}

// getEarlyBeanReference returns early bean reference for circular dependency resolution
func (c *BeanContainer) getEarlyBeanReference(beanName string, bean any) (any, error) {
	// 在这里可以应用AOP等处理
	// 对于当前实现，直接返回原始bean
	return bean, nil
}

// populateBean fills bean properties with dependencies
func (c *BeanContainer) populateBean(beanName string, beanInstance any, definition *Definition) {
	beanVal := reflect.ValueOf(beanInstance).Elem()

	// 使用缓存的字段信息
	c.mutex.RLock()
	fieldInfos := c.fieldInfoCache[definition.Type]
	c.mutex.RUnlock()

	for _, fieldInfo := range fieldInfos {
		fieldVal := beanVal.Field(fieldInfo.Index)
		if !fieldVal.IsNil() {
			continue
		}

		var injectBean any
		if fieldInfo.IsInterface {
			injectBean = c.resolveInterfaceDependency(fieldInfo.Autowire, fieldInfo.Type.Elem())
		} else {
			injectBean = c.resolveObjectDependency(fieldInfo.Autowire, fieldInfo.Type.Elem())
		}

		if injectBean == nil {
			panic(exception.NewStackErr(fmt.Sprintf("DependencyInjectionError: the bean '%s' not found for field '%s' in bean '%s'", fieldInfo.Autowire, fieldInfo.Name, beanName)))
		}

		if fieldVal.CanSet() {
			fieldVal.Set(reflect.ValueOf(injectBean))
		}
	}
}

// resolveObjectDependency resolves object type dependency
func (c *BeanContainer) resolveObjectDependency(autowire string, fieldType reflect.Type) any {
	if autowire == "-" {
		autowire = fieldType.Name()
		if len(autowire) > 0 {
			autowire = string(autowire[0]|32) + autowire[1:]
		}
	}

	// 尝试直接通过名称获取
	if bean := c.getSingleton(autowire); bean != nil {
		beanType := reflect.TypeOf(bean).Elem()
		if beanType != fieldType {
			panic(exception.NewStackErr(fmt.Sprintf("TypeMismatchError: the bean '%s' type '%s' does not match field type '%s'", autowire, beanType.Name(), fieldType.Name())))
		}
		return bean
	}

	// 按类型查找（优先使用缓存的映射）
	c.mutex.RLock()
	if beanNames, ok := c.beanNamesForTypes[fieldType]; ok && len(beanNames) > 0 {
		c.mutex.RUnlock()
		return c.getSingleton(beanNames[0])
	}

	// 遍历查找匹配类型（只在缓存未命中时执行）
	var foundName string
	for name, def := range c.beanDefinitions {
		if def.Type == fieldType {
			foundName = name
			break
		}
	}
	c.mutex.RUnlock()

	if foundName != "" {
		c.mutex.Lock()
		c.beanNamesForTypes[fieldType] = append(c.beanNamesForTypes[fieldType], foundName)
		c.mutex.Unlock()
		return c.getSingleton(foundName)
	}

	return nil
}

// resolveInterfaceDependency resolves interface type dependency
func (c *BeanContainer) resolveInterfaceDependency(autowire string, fieldType reflect.Type) any {
	if autowire == "-" {
		autowire = fieldType.Name()
		if len(autowire) > 0 {
			autowire = string(autowire[0]|32) + autowire[1:]
		}
	}

	// 尝试直接通过名称获取
	if bean := c.getSingleton(autowire); bean != nil {
		beanType := reflect.TypeOf(bean).Elem()
		if !beanType.Implements(fieldType) {
			panic(exception.NewStackErr(fmt.Sprintf("InterfaceImplementationError: the bean '%s' type '%s' does not implement interface '%s'", autowire, beanType.Name(), fieldType.Name())))
		}
		return bean
	}

	// 按类型查找（优先使用缓存的映射）
	c.mutex.RLock()
	if beanNames, ok := c.beanNamesForTypes[fieldType]; ok && len(beanNames) > 0 {
		c.mutex.RUnlock()
		return c.getSingleton(beanNames[0])
	}

	// 遍历查找实现该接口的类型（只在缓存未命中时执行）
	var foundName string
	for name, def := range c.beanDefinitions {
		if def.Type.Implements(fieldType) {
			foundName = name
			break
		}
	}
	c.mutex.RUnlock()

	if foundName != "" {
		c.mutex.Lock()
		c.beanNamesForTypes[fieldType] = append(c.beanNamesForTypes[fieldType], foundName)
		c.mutex.Unlock()
		return c.getSingleton(foundName)
	}

	return nil
}

// initializeBean performs bean initialization
func (c *BeanContainer) initializeBean(beanName string, beanInstance any) {
	// 调用PostConstruct方法
	if abstractBean, ok := beanInstance.(bean.AbstractBean); ok {
		abstractBean.BeanPostConstruct()
	}
}

// GetBean retrieves a bean by name
func (c *BeanContainer) GetBean(name string) any {
	return c.getSingleton(name)
}
