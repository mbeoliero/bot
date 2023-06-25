package conf

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	filename = "conf.yaml"
)

type ServerConf struct {
	Account struct {
		Uid      int64  `yaml:"uid"`
		Password string `yaml:"password"`
	} `yaml:"account"`

	Gpt struct {
		ApiKey string `yaml:"api_key"`
	} `yaml:"gpt"`

	WhiteList struct {
		GroupList []int64 `yaml:"group_list"`
		UserList  []int64 `yaml:"user_list"`
	} `yaml:"white_list"`
}

var config ServerConf
var once sync.Once

func loadConf(config interface{}, filename string) {
	cur, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := cur + "/conf/"
	path += filename

	if err := loadConfFromYaml(path, config); err != nil {
		panic(err)
	}
}

func loadConfFromYaml(filePath string, v interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(v); err != nil {
		return err
	}
	return nil
}

func InitConfig() {
	once.Do(func() {
		loadConf(&config, filename)
	})
}

func Get() ServerConf {
	return config
}
