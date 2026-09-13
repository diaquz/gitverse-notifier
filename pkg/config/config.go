package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

type Config struct {
	Root string

	LogDirPath string `mapstructure:"LOG_DIR_PATH"`
	LogMaxSize int    `mapstructure:"LOG_MAX_SIZE"`
	LogMaxAge  int    `mapstructure:"LOG_MAX_AGE"`

	ConfigsDirPath      string `mapstructure:"CONFIGS_DIR_PATH"`
	TemplatesDirPath    string `mapstructure:"TEMPLATES_DIR_PATH"`
	TemplatesPattern    string `mapstructure:"TEMPLATES_PATTERN"`
	RepositoriesDirPath string `mapstructure:"REPOSITORIES_DIR_PATH"`

	BasicAuthUser     string `mapstructure:"BASIC_AUTH_USER"`
	BasicAuthPassword string `mapstructure:"BASIC_AUTH_PASSWORD"`

	BindHost    string `mapstructure:"BIND_HOST"`
	HTTPPort    string `mapstructure:"HTTPD_PORT"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	LogFileName string `mapstructure:"LOG_FILE_NAME"`

	LanguageCode     string `mapstructure:"LANGUAGE_CODE"`

	JiraURL string `mapstructure:"JIRA_URL"`

	JiraUsername string `mapstructure:"JIRA_USERNAME"`
	JiraPassword string `mapstructure:"JIRA_PASSWORD"`
	JiraToken    string `mapstructure:"JIRA_TOKEN"`

	TelegramBotToken string `mapstructure:"TELEGRAM_BOT_TOKEN"`
	TelegramChatID   string `mapstructure:"TELEGRAM_CHAT_ID"`
	TelegramProxyURL string `mapstructure:"TELEGRAM_PROXY_URL"`

	GitverseBaseURL string `mapstructure:"GITVERSE_BASE_URL"`
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

	dataFolderPath := filepath.Join(rootPath, "data")
	logDirPath := filepath.Join(dataFolderPath, "logs")

	configPath := filepath.Join(rootPath, "configs")
	templatesPath := filepath.Join(configPath, "templates")
	repositoriesPath := filepath.Join(configPath, "repositories")

	folders := []string{dataFolderPath, logDirPath, configPath, templatesPath, repositoriesPath}
	for i := range folders {
		if err := EnsureDirExist(folders[i]); err != nil {
			log.Fatalf("Create folder failed: %s", err.Error())
		}
	}

	return Config{
		Root:                rootPath,
		LogDirPath:          logDirPath,
		LogMaxSize:          15,
		LogMaxAge:           7,
		ConfigsDirPath:      configPath,
		TemplatesDirPath:    templatesPath,
		RepositoriesDirPath: repositoriesPath,
		BindHost:            "0.0.0.0",
		HTTPPort:            "9001",
		LogLevel:            "INFO",
		LogFileName:         "gitverse-notifier.log",
		LanguageCode:        "ru",
		TemplatesPattern:    "**/*.tmpl",
	}
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
		log.Println("Successfully loaded config from env")
	}
}

func loadConfigFromFile(path string, conf *Config) {
	var err error
	if _, err1 := os.Stat(path); err1 == nil {
		fileViper := viper.New()
		fileViper.SetConfigFile(path)
		if err = fileViper.ReadInConfig(); err == nil {
			if err = fileViper.Unmarshal(conf); err == nil {
				log.Printf("Successfully loaded config from %s", path)
				return
			}
		}
	}
	if err != nil {
		log.Fatalf("failed to load config from %s: %s", path, err)
	}
}
