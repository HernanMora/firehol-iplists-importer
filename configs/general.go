package configs

import (
	"io/ioutil"

	"github.com/HernanMora/firehol-iplists-importer/logger"
	"gopkg.in/yaml.v2"
)

type ElasticsearchSettings struct {
	IndexName string `yaml:"index"`
	Url       string `yaml:"url"`
}

type GeneralSettings struct {
	Repository string   `yaml:"repository"`
	Sets       []string `yaml:"sets,flow"`
	Skip       []string `yaml:"skip,flow"`
	Logfile    string   `yaml:"logfile"`
}

type Settings struct {
	General       GeneralSettings       `yaml:"general"`
	ElasticConfig ElasticsearchSettings `yaml:"elasticsearch"`
}

func Load() (*Settings, error) {
	logger.Info("Loading configuration file...")
	source, err := ioutil.ReadFile("config.yml")
	if err != nil {
		logger.Fatalf("Error loading configuration file: %v", err)
	}
	logger.Info("Configuration file loaded")
	cf := &Settings{}
	err = yaml.Unmarshal(source, &cf)
	if err != nil {
		logger.Fatalf("Error parsing configuration file: %v", err)
	}
	cf = setDefaults(cf)
	logger.Info("Configuration file parsed")

	return cf, err
}

func setDefaults(settings *Settings) *Settings {
	if len(settings.General.Repository) == 0 {
		logger.Info("General Repository: Setting default value \"https://github.com/firehol/blocklist-ipsets.git\"")
		settings.General.Repository = "https://github.com/firehol/blocklist-ipsets.git"
	}

	if len(settings.General.Sets) == 0 {
		logger.Info("General sets: Setting default value \"[]\"")
		settings.General.Sets = []string{}
	}

	if len(settings.General.Skip) == 0 {
		logger.Info("General Skip: Setting default value \"[]\"")
		settings.General.Skip = []string{}
	}

	if len(settings.General.Logfile) == 0 {
		logger.Info("General logfile: Setting default value \"\"")
		settings.General.Logfile = ""
	}

	if len(settings.ElasticConfig.Url) == 0 {
		logger.Info("Elasticsearch URL: Setting default value \"http://localhost:9200\"")
		settings.ElasticConfig.Url = "http://localhost:9200"
	}

	if len(settings.ElasticConfig.IndexName) == 0 {
		logger.Info("Elasticsearch index name: Setting default value \"firehol-ipsets\"")
		settings.ElasticConfig.IndexName = "firehol-ipsets"
	}

	return settings
}
