package gormlog

import (
	"context"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm/logger"

	"github.com/aide-family/magicbox/log"
)

var _ logger.Interface = (*gormLogger)(nil)

func New(logger log.Interface) logger.Interface {
	return &gormLogger{helper: klog.NewHelper(klog.With(logger, "module", "gorm"))}
}

type gormLogger struct {
	helper *klog.Helper
}

// Error implements logger.Interface.
func (g *gormLogger) Error(_ context.Context, msg string, args ...any) {
	g.helper.Errorf(msg, args...)
}

// Info implements logger.Interface.
func (g *gormLogger) Info(_ context.Context, msg string, args ...any) {
	g.helper.Infof(msg, args...)
}

// LogMode implements logger.Interface.
func (g *gormLogger) LogMode(logger.LogLevel) logger.Interface {
	return g
}

// Trace implements logger.Interface.
func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	since := time.Since(begin)
	sql, rowsAffected := fc()
	if err != nil {
		g.helper.Errorw("sql", sql, "rowsAffected", rowsAffected, "since", since, "error", err)
	} else {
		g.helper.Infow("sql", sql, "rowsAffected", rowsAffected, "since", since)
	}
}

// Warn implements logger.Interface.
func (g *gormLogger) Warn(_ context.Context, msg string, args ...any) {
	g.helper.Warnf(msg, args...)
}
