// Package stdio is a log driver for stdio logger.
package stdio

import (
	"io"
	"os"

	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/log"
)

var _ log.Driver = (*initializer)(nil)

// LoggerDriver is a log driver for stdio logger.
func LoggerDriver(ws ...io.Writer) log.Driver {
	var w io.Writer = os.Stdout
	if len(ws) > 0 {
		w = ws[0]
	}
	return &initializer{w: w}
}

type initializer struct {
	w io.Writer
}

func (i *initializer) New() (log.Interface, error) {
	return klog.NewStdLogger(i.w), nil
}
