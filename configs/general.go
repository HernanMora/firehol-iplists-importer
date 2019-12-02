package configs

import (
	"io/ioutil"

	"github.com/HernanMora/firehol-iplists-importer/logger"
	"gopkg.in/yaml.v2"
)

type CSVSettings struct {
	Comma        string   `yaml:"comma"`
	Fields       []string `yaml:"fields,flow"`
	OutputFolder string   `yaml:"outputFolder"`
}

type ElasticsearchSettings struct {
	IndexName string `yaml:"index"`
	Url       string `yaml:"url"`
}

type GeneralSettings struct {
	Repository        string   `yaml:"repository"`
	ExcludeSets       []string `yaml:"ignore_sets,flow"`
	ExcludeCategories []string `yaml:"ignore_categories,flow"`
}

type Settings struct {
	General       GeneralSettings       `yaml:"general"`
	CSVConfig     CSVSettings           `yaml:"csv"`
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

	if len(settings.General.ExcludeSets) == 0 {
		logger.Info("General Exclude sets: Setting default value \"[]\"")
		settings.General.ExcludeSets = []string{}
	}

	if len(settings.General.ExcludeCategories) == 0 {
		logger.Info("General Exclude categories: Setting default value \"[]\"")
		settings.General.ExcludeCategories = []string{}
	}

	if len(settings.CSVConfig.Comma) == 0 {
		logger.Info("CSV Comma: Setting default value \",\"")
		settings.CSVConfig.Comma = ","
	}

	if len(settings.CSVConfig.Fields) == 0 {
		logger.Info("CSV Fields: Setting default value \"[\"ip\", \"category\", \"ipset\", \"maintainer\", \"version\"]\"")
		settings.CSVConfig.Fields = []string{"ip", "category", "ipset", "maintainer", "version"}
	}

	if len(settings.CSVConfig.OutputFolder) == 0 {
		logger.Info("CSV Output Folder: Setting default value \"/tmp\"")
		settings.CSVConfig.OutputFolder = "/tmp"
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
