package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var GlobalConfig *Config

type Config struct {
	Root          string
	LogDirPath    string
	ActionDirPath string

	BindHost    string `mapstructure:"BIND_HOST"`
	HTTPPort    string `mapstructure:"HTTPD_PORT"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	LogFileName string `mapstructure:"LOG_FILE_NAME"`

	SecretEncryptKey string `mapstructure:"SECRET_ENCRYPT_KEY"`
	LanguageCode     string `mapstructure:"LANGUAGE_CODE"`

	JiraURL string `mapstructure:"JIRA_URL"`

	JiraUsername string `mapstructure:"JIRA_USERNAME"`
	JiraPassword string `mapstructure:"JIRA_PASSWORD"`
	JiraToken    string `mapstructure:"JIRA_TOKEN"`
}

func Setup(configPath string) {
	var conf = getDefaultConfig()
	loadConfigFromEnv(&conf)
	loadConfigFromFile(configPath, &conf)
	GlobalConfig = &conf
	log.Printf("%+v\n", GlobalConfig)

}

func getDefaultConfig() Config {
	rootPath := getPwdDirPath()
	actionsPath := filepath.Join(rootPath, "actions")
	dataFolderPath := filepath.Join(rootPath, "data")
	LogDirPath := filepath.Join(dataFolderPath, "logs")

	folders := []string{dataFolderPath, LogDirPath, actionsPath}
	for i := range folders {
		if err := EnsureDirExist(folders[i]); err != nil {
			log.Fatalf("Create folder failed: %s", err.Error())
		}
	}

	return Config{
		Root:          rootPath,
		ActionDirPath: actionsPath,
		BindHost:      "0.0.0.0",
		HTTPPort:      "9001",
		LogLevel:      "INFO",
		LogFileName:   "gitverse-notifier.log",
		LanguageCode:  "ru",
	}
}

func ExistFile(path string) string {
	if FileExists(path) {
		return path
	}
	return ""
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func EnsureDirExist(path string) error {
	if !haveDir(path) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func haveDir(file string) bool {
	fi, err := os.Stat(file)
	return err == nil && fi.IsDir()
}

func getPwdDirPath() string {
	if rootPath, err := os.Getwd(); err == nil {
		return rootPath
	}
	return ""
}

func loadConfigFromEnv(conf *Config) {
	viper.AutomaticEnv()
	envViper := viper.New()
	for _, item := range os.Environ() {
		envItem := strings.SplitN(item, "=", 2)
		if len(envItem) == 2 {
			envViper.Set(envItem[0], viper.Get(envItem[0]))
		}
	}
	if err := envViper.Unmarshal(conf); err == nil {
		log.Println("Load config from env")
	}

}

func loadConfigFromFile(path string, conf *Config) {
	var err error
	if _, err1 := os.Stat(path); err1 == nil {
		fileViper := viper.New()
		fileViper.SetConfigFile(path)
		if err = fileViper.ReadInConfig(); err == nil {
			if err = fileViper.Unmarshal(conf); err == nil {
				log.Printf("Load config from %s success\n", path)
				return
			}
		}
	}
	if err != nil {
		log.Fatalf("Load config from %s failed: %s\n", path, err)
	}
}
