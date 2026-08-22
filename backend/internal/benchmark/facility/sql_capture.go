package facilitybenchmark

import (
	"context"
	"time"

	"gorm.io/gorm/logger"
)

type sqlCapture struct{ statements []string }

func (c *sqlCapture) LogMode(logger.LogLevel) logger.Interface { return c }
func (c *sqlCapture) Info(context.Context, string, ...any)     {}
func (c *sqlCapture) Warn(context.Context, string, ...any)     {}
func (c *sqlCapture) Error(context.Context, string, ...any)    {}

func (c *sqlCapture) Trace(_ context.Context, _ time.Time, sql func() (string, int64), _ error) {
	statement, _ := sql()
	c.statements = append(c.statements, statement)
}
