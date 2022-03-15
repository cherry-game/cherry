package main

import (
	"fmt"
	"github.com/cherry-game/cherry/logger/rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"time"
)

var sugarLogger *zap.SugaredLogger

func main() {
	InitLogger()
	defer sugarLogger.Sync()

	httpServer()
}

func httpServer() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		sugarLogger.Debugf("GET request for %s", request.RequestURI)
		writer.Write([]byte("ok " + time.Now().String()))
	})

	fmt.Println("http://127.0.0.1:9090")
	http.ListenAndServe(":9090", nil)
}

func InitLogger() {
	encoder := getEncoder()

	//两个interface,判断日志等级
	//warnlevel以下归到info日志
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel
	})
	//warnlevel及以上归到warn日志
	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})

	infoWriter := getLogWriter("./_examples/test_zap/logs/info")
	warnWriter := getLogWriter("./_examples/test_zap/logs/warn")

	//创建zap.Core,for logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, infoWriter, infoLevel),
		zapcore.NewCore(encoder, warnWriter, warnLevel),
	)
	//生成Logger
	logger := zap.New(core, zap.AddCaller())
	sugarLogger = logger.Sugar()
}

//getEncoder
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

//得到LogWriter
func getLogWriter(filePath string) zapcore.WriteSyncer {
	warnIoWriter := getWriter(filePath)
	return zapcore.AddSync(warnIoWriter)
}

//日志文件切割
func getWriter(filename string) io.Writer {
	// 保存日志30天，每24小时分割一次日志
	/*
		hook, err := rotatelogs.New(
			filename+"_%Y%m%d.log",
			rotatelogs.WithLinkName(filename),
			rotatelogs.WithMaxAge(time.Hour*24*30),
			rotatelogs.WithRotationTime(time.Hour*24),
		)
	*/
	//保存日志30天，每1分钟分割一次日志
	hook, err := rotatelogs.New(
		//filename+"_%Y%m%d%H%M.log",
		filename,
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*30),
		rotatelogs.WithRotationTime(time.Minute*1),
	)
	if err != nil {
		panic(err)
	}
	return hook
}
