package dlog

import (
	"github.com/hyperion-hyn/common/dlog/hooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)


var LogFields map[string]interface{}

func init() {
	// 以JSON格式为输出，代替默认的ASCII格式
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	// 以Stdout为输出，代替默认的stderr
	logrus.SetOutput(os.Stdout)
	// 设置日志等级
	logrus.SetLevel(logrus.InfoLevel)

	logrus.SetReportCaller(true)

	logrus.AddHook(hooks.NewContextHook())
}

func WriteToFile(fields map[string]interface{}, logs string, level logrus.Level) {
	CurrDir, _ := filepath.Abs("./logs")
	DLog := logrus.New()

	DLog.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	// 这三个等级，记录方法调用
	if level == logrus.ErrorLevel || level == logrus.PanicLevel || level == logrus.FatalLevel {
		DLog.SetReportCaller(true)
		DLog.AddHook(hooks.NewContextHook())

		// 发email
		emailHook, err := hooks.NewMailAuthHook(viper.GetString("app.name"), viper.GetString("app.send_email.host"),
			viper.GetInt("app.send_email.port"), viper.GetString("app.send_email.from"),
			viper.GetString("app.send_email.to"), viper.GetString("app.send_email.username"),
			viper.GetString("app.send_email.auth_pwd"))
		if err == nil {
			DLog.Hooks.Add(emailHook)

			context := DLog.WithField("From", "Hyperion dmapper server")

			localTimer, _ := time.LoadLocation("Asia/Chongqing")
			context.Time = time.Now().In(localTimer)
			context.Message = "Hyperion dmapper server alarm message"
			context.Level = level

			err = emailHook.Fire(context)
			if err != nil {
				log.Printf("Send alarm message email get error: %v", err)
			}
		}
	}

	fileH, err := os.OpenFile(CurrDir + "/application.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer fileH.Close()
	if err == nil {
		// 同时输出到文件和console
		mw := io.MultiWriter(os.Stdout, fileH)
		DLog.SetOutput(mw)
	} else {
		// 以Stdout为输出，代替默认的stderr
		DLog.SetOutput(os.Stdout)
		DLog.Infof("Failed to log to file, using default stderr: $v", err)
	}

	if len(fields) > 0 {
		DLog.WithFields(fields)
	}

	switch level {
	case logrus.TraceLevel:
		DLog.Trace(logs)
	case logrus.DebugLevel:
		DLog.Debug(logs)
	case logrus.InfoLevel:
		DLog.Info(logs)
	case logrus.WarnLevel:
		DLog.Warn(logs)
	case logrus.ErrorLevel:
		DLog.Error(logs)
	case logrus.FatalLevel:
		DLog.Fatal(logs)
	case logrus.PanicLevel:
		DLog.Panic(logs)
	}
}




