package elasticsearch

import (
	"github.com/HernanMora/firehol-iplists-importer/configs"
	"github.com/HernanMora/firehol-iplists-importer/models"
	ESService "github.com/HernanMora/firehol-iplists-importer/service/elasticsearch"
)

type ESWorker struct {
	config    configs.ElasticsearchSettings
	esService ESService.ElasticsearchService
}

func NewESWorker() *ESWorker {
	return &ESWorker{}
}

var worker *ESWorker

func init() {
	worker = NewESWorker()
}

func Setup(configuration configs.ElasticsearchSettings) {
	worker.esService = ESService.NewElasticService(configuration)
	worker.esService.CheckTemplate()
}

func IndexDocuments(docs []models.Document, timestamp int64) {
	worker.esService.InsertToElasticsearch(docs, timestamp)
}
