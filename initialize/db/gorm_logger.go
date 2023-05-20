package db

import (
	"context"
	"fmt"
	"github.com/JaanaiShi/flint/common"
	"github.com/JaanaiShi/flint/initialize/logger"
	"go.uber.org/zap"
	gormLog "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type writer interface {
	Printf(string, ...interface{})
}

type customConfig struct {
	SlowThreshold time.Duration
	Colorful      bool
	LogLevel      gormLog.LogLevel
}

var (
	Discard = New(log.New(ioutil.Discard, "", log.LstdFlags), customConfig{})
	Default = New(log.New(os.Stdout, "\r\n", log.LstdFlags), customConfig{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      gormLog.Warn,
		Colorful:      true,
	})
	Recorder = traceRecorder{Interface: Default, BeginAt: time.Now()}
)

func New(writer writer, config customConfig) gormLog.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s;  #%s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s;  #%s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s;  #%s"
	)

	if config.Colorful {
		infoStr = gormLog.Green + "%s\n" + gormLog.Reset + gormLog.Green + "[info] " + gormLog.Reset
		warnStr = gormLog.BlueBold + "%s\n" + gormLog.Reset + gormLog.Magenta + "[warn] " + gormLog.Reset
		errStr = gormLog.Magenta + "%s\n" + gormLog.Reset + gormLog.Red + "[error] " + gormLog.Reset
		traceStr = gormLog.Green + "%s\n" + gormLog.Reset + gormLog.Yellow + "[%.3fms] " + gormLog.BlueBold + "[rows:%v]" + gormLog.Reset + " %s" + ";  #%s"
		traceWarnStr = gormLog.Green + "%s " + gormLog.Yellow + "%s\n" + gormLog.Reset + gormLog.RedBold + "[%.3fms] " + gormLog.Yellow + "[rows:%v]" + gormLog.Magenta + " %s" + gormLog.Reset + ";  #%s"
		traceErrStr = gormLog.RedBold + "%s " + gormLog.MagentaBold + "%s\n" + gormLog.Reset + gormLog.Yellow + "[%.3fms] " + gormLog.BlueBold + "[rows:%v]" + gormLog.Reset + " %s" + ";  #%s"
	}

	return &customLogger{
		writer:       writer,
		customConfig: config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type customLogger struct {
	writer
	customConfig
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func (c *customLogger) LogMode(level gormLog.LogLevel) gormLog.Interface {
	newLogger := *c
	newLogger.LogLevel = level
	return &newLogger
}

func (c *customLogger) Info(ctx context.Context, message string, data ...interface{}) {
	if c.LogLevel >= gormLog.Info {
		c.Printf(c.infoStr+message, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (c *customLogger) Warn(ctx context.Context, message string, data ...interface{}) {
	if c.LogLevel >= gormLog.Warn {
		c.Printf(c.warnStr+message, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (c *customLogger) Error(ctx context.Context, message string, data ...interface{}) {
	if c.LogLevel >= gormLog.Error {
		c.Printf(c.errStr+message, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (c *customLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if c.LogLevel > 0 {
		fileWithLineNum := strings.Replace(utils.FileWithLineNum(), common.ROOT_PATH+"/", "", 1)

		elapsed := time.Since(begin)
		switch {
		case err != nil && c.LogLevel >= gormLog.Error:
			sql, rows := fc()
			if rows == -1 {
				c.Printf(c.traceErrStr, fileWithLineNum, err, float64(elapsed.Nanoseconds())/1e6, "-", sql, ctx)
			} else {
				c.Printf(c.traceErrStr, fileWithLineNum, err, float64(elapsed.Nanoseconds())/1e6, rows, sql, ctx)
			}
		case elapsed > c.SlowThreshold && c.SlowThreshold != 0 && c.LogLevel >= gormLog.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", c.SlowThreshold)
			if rows == -1 {
				c.Printf(c.traceWarnStr, fileWithLineNum, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql, ctx)
			} else {
				c.Printf(c.traceWarnStr, fileWithLineNum, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql, ctx)
			}
		case c.LogLevel >= gormLog.Info:
			sql, rows := fc()
			if rows == -1 {
				c.Printf(c.traceStr, fileWithLineNum, float64(elapsed.Nanoseconds())/1e6, "-", sql, ctx)
			} else {
				c.Printf(c.traceStr, fileWithLineNum, float64(elapsed.Nanoseconds())/1e6, rows, sql, ctx)
			}
		default:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", c.SlowThreshold)
			if rows == -1 {
				c.Printf(c.traceWarnStr, fileWithLineNum, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql, ctx)
			} else {
				c.Printf(c.traceWarnStr, fileWithLineNum, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql, ctx)
			}
		}
	}
}

func (c *customLogger) Printf(message string, data ...interface{}) {
	if common.Conf.DB.LogZap != "" {
		msg := "gorm"
		ctx := logger.SetContext(context.Background(), "")
		switch len(data) {
		case 0:
			logger.Info(ctx, message)
		case 1:
			logger.Info(ctx, "gorm", zap.Any("src", data[0]))
		case 2:
			logger.Info(ctx, "gorm", zap.Any("src", data[0]), zap.Any("duration", data[1]))
		case 3:
			logger.Info(ctx, "gorm", zap.Any("src", data[0]), zap.Any("duration", data[1]), zap.Any("rows", data[2]))
		case 4:
			logger.Info(ctx, "gorm", zap.Any("src", data[0]), zap.Any("duration", data[1]), zap.Any("rows", data[2]), zap.Any("sql", data[3]))
		case 5:
			logger.Info(ctx, "gorm", zap.Any("src", data[0]), zap.Any("duration", data[1]), zap.Any("rows", data[2]), zap.Any("sql", data[3]))
		case 6:
			if _, ok := data[1].(string); ok {
				msg = data[1].(string)
			}
			if _, ok := data[5].(context.Context); ok {
				ctx = data[5].(context.Context)
			}
			logger.Info(ctx, msg, zap.Any("src", data[0]), zap.Any("duration", data[2]), zap.Any("rows", data[3]), zap.Any("sql", data[4]))
		default:
			logger.Info(ctx, "gorm", zap.Any("src", data[0]), zap.Any("duration", data[1]), zap.Any("rows", data[2]), zap.Any("sql", data[3]))
		}
		return
	}

	switch len(data) {
	case 0:
		c.writer.Printf(message, "")
	case 1:
		c.writer.Printf(message, data[0])
	case 2:
		c.writer.Printf(message, data[0], data[1])
	case 3:
		c.writer.Printf(message, data[0], data[1], data[2])
	case 4:
		c.writer.Printf(message, data[0], data[1], data[2], data[3])
	case 5:
		c.writer.Printf(message, data[0], data[1], data[2], data[3], data[4].(context.Context).Value("link_id").(string))
	case 6:
		c.writer.Printf(message, data[0], data[1], data[2], data[3], data[5].(context.Context).Value("link_id").(string))
	}
}

type traceRecorder struct {
	gormLog.Interface
	BeginAt      time.Time
	SQL          string
	RowsAffected int64
	Err          error
}

func (t traceRecorder) New() *traceRecorder {
	return &traceRecorder{Interface: t.Interface, BeginAt: time.Now()}
}

func (t *traceRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	t.BeginAt = begin
	t.SQL, t.RowsAffected = fc()
	t.Err = err
}
