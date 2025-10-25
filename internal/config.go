package internal

import (
	"math/rand"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port         int `yaml:"port"`
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
}

type TTFTConfig struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

type ModelConfig struct {
	Name             string     `yaml:"name"`
	MaxContextTokens int        `yaml:"max_context_tokens"`
	MaxOutputTokens  int        `yaml:"max_output_tokens"`
	TTFT             TTFTConfig `yaml:"ttft"`
	OTPS             int        `yaml:"otps"`
}

type CommonConfig struct {
	TokensFile string `yaml:"tokens_file"`
}

type Config struct {
	Server   ServerConfig  `yaml:"server"`
	Common   CommonConfig  `yaml:"common"`
	Models   []ModelConfig `yaml:"models"`
	ModelMap map[string]ModelConfig
}

func (c *TTFTConfig) GetTTFT() int {
	return rand.Intn(c.Max-c.Min+1) + c.Min
}

var AppConfig = new(Config)

func loadConfigFromYamlFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(content, AppConfig)
	if err != nil {
		return err
	}

	AppConfig.ModelMap = make(map[string]ModelConfig)
	for _, model := range AppConfig.Models {
		AppConfig.ModelMap[model.Name] = model
	}
	return nil
}

func InitConfig(path string) {
	err := loadConfigFromYamlFile(path)
	if err != nil {
		panic(err)
	}
}
