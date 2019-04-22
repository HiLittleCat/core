package core

import (
	"path"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

//Log core框架日志
var Log *logrus.Logger

//SetLog 设置logrus日志配置，logPath参数表示记录日志的路径，logFileName表示日志名称前缀
func SetLog(logPath string, logFileName string) {
	if Log != nil {
		return
	}
	Log = logrus.New()
	maxAge := time.Hour * 24 * 10
	rotationTime := time.Hour * 24
	//info用于记录http状态等一般日期
	infoLogPaht := path.Join(logPath, logFileName+"_info")
	infoWriter, err := rotatelogs.New(
		infoLogPaht+".%Y%m%d",
		rotatelogs.WithLinkName(infoLogPaht),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	//error记录告警和错误信息
	errorLogPaht := path.Join(logPath, logFileName+"_error")
	errorWriter, err := rotatelogs.New(
		errorLogPaht+".%Y%m%d",
		rotatelogs.WithLinkName(errorLogPaht),     // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)

	if err != nil {
		logrus.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}
	// 为不同级别设置不同的输出文件
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel:  infoWriter,
		logrus.WarnLevel:  errorWriter,
		logrus.ErrorLevel: errorWriter,
	}, &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
	Log.AddHook(lfHook)
}
