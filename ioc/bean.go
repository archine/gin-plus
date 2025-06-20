package ioc

var (
	beanCache = make(map[string]any)
)

// FactoryBean 定义可被 IoC 容器管理的 bean 接口
type FactoryBean interface {
	// CreateBean 创建并返回一个 bean 实例
	CreateBean() FactoryBean

	// GetBeanName 返回 bean 的名称
	GetBeanName() string
}

type Bean struct{}

// CreateBean 创建并返回一个 bean 实例
func (b *Bean) CreateBean() FactoryBean {
	return b
}

// GetBeanName 返回 bean 的名称
func (b *Bean) GetBeanName() string {
	return ""
}
