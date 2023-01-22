package platform

import (
	"fmt"
	"github.com/chenxuan520/goweb-platform/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"path"
)

const (
	extensionJson = ".json"
	extensionYaml = ".yaml"
	extensionInI  = ".ini"

	nameSpace = "conf"
)

var (
	//Automatic loading sequence of local Config
	autoLoadLocalConfigs = []string{
		extensionJson,
		extensionYaml,
		extensionInI,
	}
)

type (
	Mysql struct {
		Path     string `mapstructure:"path" json:"path" yaml:"path" ini:"path"`                 // 服务器地址
		Port     string `mapstructure:"port" json:"port" yaml:"port" ini:"port"`                 // 端口
		Config   string `mapstructure:"config" json:"config" yaml:"config" ini:"config"`         // 高级配置
		Dbname   string `mapstructure:"db-name" json:"dbname" yaml:"db-name" ini:"db-name"`      // 数据库名
		Username string `mapstructure:"username" json:"username" yaml:"username" ini:"username"` // 数据库用户名
		Password string `mapstructure:"password" json:"password" yaml:"password" ini:"password"` // 数据库密码
	}
	Redis struct {
		DB       int    `mapstructure:"db" json:"db" yaml:"db" ini:"db"`                         // redis的哪个数据库
		Addr     string `mapstructure:"addr" json:"addr" yaml:"addr" ini:"addr"`                 // 服务器地址:端口
		Password string `mapstructure:"password" json:"password" yaml:"password" ini:"password"` // 密码
	}
	Mongo struct {
		Host     string `mapstructure:"host" json:"host" yaml:"host" ini:"host"`
		Port     string `mapstructure:"port" json:"port" yaml:"port" ini:"port"`
		User     string `mapstructure:"user" json:"user" yaml:"user" ini:"user"`
		Password string `mapstructure:"password" json:"password" yaml:"password" ini:"password"`
		DBname   string `mapstructure:"db" json:"db" yaml:"db" ini:"db"`
	}
	System struct {
		Env        string `mapstructure:"env" json:"env" yaml:"env" ini:"env"`
		Addr       int    `mapstructure:"addr" json:"addr" yaml:"addr" ini:"addr"`
		UploadType string `mapstructure:"upload-type" json:"upload-type" yaml:"upload-type" ini:"upload-type"` // Oss类型
		Version    string `mapstructure:"version" json:"version" yaml:"version" ini:"version"`
	}
	Log struct {
		Level         string `mapstructure:"level" json:"level" yaml:"level" ini:"level"`                                    // 级别
		Format        string `mapstructure:"format" json:"format" yaml:"format" ini:"level"`                                 // 输出
		Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix" ini:"level"`                                 // 日志前缀
		Director      string `mapstructure:"director" json:"director"  yaml:"director" ini:"level"`                          // 日志文件夹
		ShowLine      bool   `mapstructure:"show-line" json:"showLine" yaml:"showLine" ini:"showLine"`                       // 显示行
		EncodeLevel   string `mapstructure:"encode-level" json:"encodeLevel" yaml:"encode-level" ini:"encode-level"`         // 编码级
		StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktraceKey" yaml:"stacktrace-key" ini:"stacktrace-key"` // 栈名
		LogInConsole  bool   `mapstructure:"log-in-console" json:"logInConsole" yaml:"log-in-console" ini:"log-in-console"`  // 输出控制台
	}
)

type Config struct {
	Log    Log    `mapstructure:"log" json:"log" yaml:"log" ini:"log"`
	System System `mapstructure:"system" json:"system" yaml:"system" ini:"system"`
	Mysql  Mysql  `mapstructure:"mysql" json:"mysql" yaml:"mysql" ini:"mysql"`
	Redis  Redis  `mapstructure:"redis" json:"redis" yaml:"redis" ini:"redis"`
	Mongo  Mongo  `mapstructure:"mongo" json:"mongo" yaml:"mongo" ini:"mongo"`
}

func (m *Mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}

func (m *Mysql) EmptyDsn() string {
	if m.Path == "" {
		m.Path = "127.0.0.1"
	}
	if m.Port == "" {
		m.Port = "3306"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/", m.Username, m.Password, m.Path, m.Port)
}

var _defaultConfig *Config

func LoadConfig(env, configFileName string) (*Config, error) {
	var c Config
	var confPath string
	dir := fmt.Sprintf("%s/%s", nameSpace, env)
	for _, registerExt := range autoLoadLocalConfigs {
		confPath = path.Join(dir, configFileName+registerExt)
		if utils.Exists(confPath) {
			break
		}
	}
	fmt.Println("the path to the configuration file you are using is :", confPath)
	v := viper.New()
	v.SetConfigFile(confPath)
	ext := utils.Ext(confPath)
	v.SetConfigType(ext)
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err := v.Unmarshal(&c); err != nil {
			fmt.Println(err)
		}
	})
	if err := v.Unmarshal(&c); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("load config is :%#v\n", c)
	_defaultConfig = &c
	return &c, nil
}

func GetConfigModels() *Config {
	return _defaultConfig
}
