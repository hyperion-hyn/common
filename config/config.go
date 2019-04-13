package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/hyperion-hyn/common/dlog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)
type Cfg struct {
	Name string
}

func SetupConfig(env string, configUrl string) {
	c := Cfg{
		Name: configUrl,
	}
	// 初始化配置文件
	if err := c.initConfig(env); err != nil {
		dlog.WriteToFile(dlog.LogFields, fmt.Sprintf("Loac config file get error: %v", err), logrus.FatalLevel)
		return
	}
	// 监控配置文件变化并热加载程序
	c.watchConfig()
}

func (c *Cfg) initConfig(env string) error {
	if c.Name != "" {
		viper.SetConfigFile(c.Name) // 如果指定了配置文件，则解析指定的配置文件
	} else {
		fmt.Println("init"+ env)
		viper.AddConfigPath("conf/" + env) // 如果没有指定配置文件，则解析默认的配置文件
		viper.SetConfigName("config")
	}

	viper.SetConfigType("json") // 设置配置文件格式为json
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil { // viper解析配置文件
		return err
	}
	return nil
}

// 监控配置文件变化并热加载程序
func (c *Cfg) watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		dlog.WriteToFile(dlog.LogFields, fmt.Sprintf("Config file changed: %s", e.Name), logrus.InfoLevel)
	})
}