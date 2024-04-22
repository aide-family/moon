package controller

import (
	"fmt"
	"k8s.io/klog/v2"
	"runtime"
	"strings"
	"time"
)

// print stack trace for debug
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(2, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() HandlerFunc {
	return func(c *Context) (*time.Duration, error) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				klog.Errorf("%s\n\n", trace(message))
			}
		}()
		return c.Next()
	}
}
