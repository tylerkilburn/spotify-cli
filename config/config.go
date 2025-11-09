package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const CONFIG_FILE_LOCATION = "config.yaml"

type YamlConfig struct {
	Spotify struct {
		ClientId     string   `yaml:"client_id"`
		ClientSecret string   `yaml:"client_secret"`
		Scope        []string `yaml:"scope"`
	} `yaml:"spotify"`
	Http struct {
		BaseUrl          string `yaml:"base_url"`
		LoginEndpoint    string `yaml:"login_endpoint"`
		Port             int    `yaml:"port"`
		RedirectEndpoint string `yaml:"redirect_endpoint"`
		RedirectUri      string // derived convenience value
	} `yaml:"http"`
	LocalTokenDirectory string `yaml:"local_token_directory"`
}

var configCache *YamlConfig

func Get() *YamlConfig {
	if configCache == nil {
		yamlData := readConfigFile()

		err := yaml.Unmarshal([]byte(yamlData), &configCache)
		if err != nil {
			log.Fatalf("Error unmarshaling YAML: %v", err)
		}
		configCache.Http.RedirectUri = fmt.Sprint(configCache.Http.BaseUrl) + ":" + fmt.Sprint(configCache.Http.Port) + fmt.Sprint(configCache.Http.RedirectEndpoint)
	}

	return configCache
}

func readConfigFile() string {
	fileData, err := os.ReadFile(CONFIG_FILE_LOCATION)

	if err != nil {
		panic(fmt.Sprintf("Error reading file: %v\n", err))
	}

	return string(fileData)
}
