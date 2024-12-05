package helper

import (
	"net/http"
	// 注册pprof
	_ "net/http/pprof"

	"github.com/go-kratos/kratos/v2/log"
)

// Pprof 开启pprof
func Pprof(address string) {
	if address == "" {
		return
	}
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Errorf("pprof listen and serve error: %v", err)
		}
	}()
}
