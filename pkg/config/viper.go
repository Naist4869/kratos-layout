package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	test = "./test/"
)

// Init init config
func Init(confPath string, prefix string) error {
	err := initConfig(confPath, prefix)
	if err != nil {
		return err
	}
	return nil
}

// initConfig init config from conf file
func initConfig(confPath string, prefix string) error {
	if confPath != "" {
		viper.SetConfigFile(confPath) // 如果指定了配置文件，则解析指定的配置文件
	} else {
		viper.AddConfigPath("conf") // 如果没有指定配置文件，则解析默认的配置文件
		viper.SetConfigName("application")
	}
	viper.SetConfigType("yaml") // 设置配置文件格式为YAML
	viper.AutomaticEnv()        // 读取匹配的环境变量
	viper.SetEnvPrefix(prefix)  // 读取环境变量的前缀为 lidian
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil { // viper解析配置文件
		return fmt.Errorf("err: %+v,stack: %s", err, debug.Stack())
	}
	watchConfig()
	if viper.GetString("deploy.env") == "uat" {
		os.Setenv("DEPLOY_ENV", "uat")
	}
	log.Printf("%s 启动, 部署环境: %s", prefix, os.Getenv("DEPLOY_ENV"))

	return nil
}

// 监控配置文件变化并热加载程序
func watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file changed: %s", e.Name)
	})
}
func IsDebug() bool {
	return flag.Lookup("test.v") != nil
}

func GuessTestDirPath() (string, error) {
	// 对于debug模式，其实就是测试模式，工作目录是单个的模块目录，那么需要进入具备main.go的目录
	files := make(map[string]*os.File, 10)
	defer func() {
		for _, file := range files {
			_ = file.Close()
		}
	}()
	ok := false
	if IsDebug() {
		for wd, _ := os.Getwd(); !ok; wd = filepath.Dir(wd) {
			log.Printf("判断目录[%s]是否为主目录\n", wd)
			file, err := os.Open(wd)
			if err != nil {
				return "", fmt.Errorf("定位主目录: %w", err)
			}
			files[wd] = file
			// 读取文件夹里的文件
			fileInfos, err := file.Readdir(-1)
			// 读取目录中间有错误发生
			if err != nil {
				return "", fmt.Errorf("读取目录信息: %w", err)
			}
			// 一路读到目录的末尾
			for _, info := range fileInfos {
				if strings.Contains(info.Name(), "go.mod") {
					err := os.Chdir(wd)
					if err != nil {
						return "", fmt.Errorf("设置目录: %w", err)
					}
					ok = true
					break
				}
			}
		}
	}
	if !ok {
		return "", errors.New("未找到测试环境的Config")
	}
	return test, nil
}
