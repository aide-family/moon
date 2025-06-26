package safety

import "github.com/go-kratos/kratos/v2/log"

func Go(name string, fn func(), logger log.Logger) {
	helper := log.NewHelper(log.With(logger, "module", "safety.go", "name", name))
	go func() {
		defer func() {
			if err := recover(); err != nil {
				helper.Errorf("panic: %v", err)
			}
		}()
		fn()
	}()
}
