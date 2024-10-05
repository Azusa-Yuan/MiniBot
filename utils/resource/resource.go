package resource

import "context"

// ResourceManager 管理注册的资源释放函数
type resourceManager struct {
	cleanupFuncs []func(context.Context) error
}

// Register 注册释放资源的函数
func (rm *resourceManager) Register(cleanupFunc func(context.Context) error) {
	rm.cleanupFuncs = append(rm.cleanupFuncs, cleanupFunc)
}

// Cleanup 执行所有注册的释放函数
func (rm *resourceManager) Cleanup(ctx context.Context) error {
	for _, cleanup := range rm.cleanupFuncs {
		if err := cleanup(ctx); err != nil {
			return err
		}
	}
	return nil
}

var ResourceManager = &resourceManager{}
