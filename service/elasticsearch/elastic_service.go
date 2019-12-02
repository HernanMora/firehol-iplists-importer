package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"time"

	"github.com/HernanMora/firehol-iplists-importer/configs"
	elasticsConf "github.com/HernanMora/firehol-iplists-importer/configs/elasticsearch"
	logger "github.com/HernanMora/firehol-iplists-importer/logger"
	"github.com/HernanMora/firehol-iplists-importer/models"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
)

type ElasticsearchService struct {
	esClient  *elasticsearch.Client
	indexName string
}

type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

var ElasticService *ElasticsearchService

func init() {
	ElasticService = NewElasticService()
}

func NewElasticService() *ElasticsearchService {
	return &ElasticsearchService{}
}

func (es *ElasticsearchService) SetClient(c *elasticsearch.Client) {
	es.esClient = c
}

func (es *ElasticsearchService) SetIndexName(index string) {
	es.indexName = index
}

func Setup(config configs.ElasticsearchSettings) {
	logger.Info("Setting up Elasticsearch Service...")

	cfg := elasticsearch.Config{
		Addresses: []string{
			config.Url,
		},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 15 * time.Second,
		},
	}

	esClient, _ := elasticsearch.NewClient(cfg)

	ElasticService.esClient = esClient
	ElasticService.indexName = config.IndexName
}

func CheckTemplate() {
	logger.Infof("Checking template %s from elasticsearch", elasticsConf.TemplateName)
	response, _ := ElasticService.esClient.Indices.ExistsTemplate([]string{elasticsConf.TemplateName})
	if response.IsError() && response.StatusCode != 404 {
		logger.Errorf("Error on check index template %s from elasticsearch", elasticsConf.TemplateName)
	} else if response.IsError() && response.StatusCode == 404 {
		logger.Infof("Index template %s not found... Creating now", elasticsConf.TemplateName)
		template := strings.Replace(elasticsConf.FireholTemplate, "@indexName@", ElasticService.indexName, -1)
		var buffer bytes.Buffer
		templateBody := []byte(fmt.Sprintf("%s\n", template))
		buffer.Grow(len(templateBody))
		buffer.Write(templateBody)
		response, err := ElasticService.esClient.Indices.PutTemplate(elasticsConf.TemplateName, bytes.NewReader(buffer.Bytes()))
		if err != nil {
			logger.Errorf("Cannot create index template %s: %s", elasticsConf.TemplateName, err)
		}
		if response.IsError() {
			logger.Errorf("Cannot create index template %s: %s", elasticsConf.TemplateName, response)
		}
		time.Sleep(1 * time.Second)
	}
}

func removeOldDocs(index string, exportedAt time.Time) {
	/*
		var buffer bytes.Buffer
		var raw map[string]interface{}
		deleteBody := []byte(fmt.Sprintf(`{"query": { "range": { "exported_at": { "lt" : "%d" } } } }\n`, exportedAt.UnixNano()/1000000))
		buffer.Grow(len(deleteBody))
		buffer.Write(deleteBody)
		response, err := ElasticService.EsClient.DeleteByQuery([]string{index}, bytes.NewReader(buffer.Bytes()))
		if err != nil {
			logger.Errorf("Cannot delete old rules from index %s: %s", index, err)
		}
		if response.IsError() {
			if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
				log.Fatalf("Failure to to parse response body: %s", err)
			} else {
				logger.Errorf("Error: %v", raw)
			}
		}
	*/
}

func InsertToElasticsearch(docs []models.Document, timestamp time.Time) {
	response, _ := ElasticService.esClient.Indices.Exists([]string{ElasticService.indexName})
	if response.IsError() && response.StatusCode != 404 {
		logger.Errorf("Error on check index %s from elasticsearch", ElasticService.indexName)
	} else if response.IsError() && response.StatusCode == 404 {
		logger.Infof("Index %s not found... Creating now", ElasticService.indexName)
		response, err := ElasticService.esClient.Indices.Create(ElasticService.indexName)
		if err != nil {
			logger.Errorf("Cannot create index %s: %s", ElasticService.indexName, err)
		}
		if response.IsError() {
			logger.Errorf("Cannot create index %s: %s", ElasticService.indexName, response)
		}
		time.Sleep(2 * time.Second)
	}

	ok := insertRulesBulk(ElasticService.indexName, elasticsConf.DocumentType, docs, timestamp)
	if ok {
		// Needed to avoid version conflicts of docs
		time.Sleep(1 * time.Second)
		removeOldDocs(ElasticService.indexName, timestamp)
	}
}

func insertRulesBulk(index string, indexType string, docs []models.Document, exportedAt time.Time) bool {
	var (
		buf             bytes.Buffer
		elasticResponse *esapi.Response
		raw             map[string]interface{}
		blk             *bulkResponse
	)
	for _, doc := range docs {
		doc["@timestamp"] = exportedAt
		id := fmt.Sprintf("%s-%s", doc["ipset"], doc["ip"])
		meta := []byte(fmt.Sprintf(`{ "index" : { "_type": "%s", "_id" : "%s"} }%s`, indexType, id, "\n"))
		bulkData, err := json.Marshal(doc)
		if err != nil {
			log.Fatalf("Cannot encode document %s: %s", id, err)
		}
		bulkData = append(bulkData, "\n"...)
		buf.Grow(len(meta) + len(bulkData))
		buf.Write(meta)
		buf.Write(bulkData)
	}

	elasticResponse, err := ElasticService.esClient.Bulk(bytes.NewReader(buf.Bytes()), ElasticService.esClient.Bulk.WithIndex(index))
	if err != nil {
		log.Fatalf("Failure indexing batch: %s", err)
	}
	if elasticResponse.IsError() {
		if err := json.NewDecoder(elasticResponse.Body).Decode(&raw); err != nil {
			log.Fatalf("Failure to to parse response body: %s", err)
		} else {
			logger.Errorf("Error: [%d] %s: %s",
				elasticResponse.StatusCode,
				raw["error"].(map[string]interface{})["type"],
				raw["error"].(map[string]interface{})["reason"],
			)
		}
		return false
	} else {
		if err := json.NewDecoder(elasticResponse.Body).Decode(&blk); err != nil {
			log.Fatalf("Failure to to parse response body: %s", err)
		} else {
			for _, d := range blk.Items {
				if d.Index.Status > 201 {
					logger.Errorf(" Error: [%d]: %s: %s: %s: %s",
						d.Index.Status,
						d.Index.Error.Type,
						d.Index.Error.Reason,
						d.Index.Error.Cause.Type,
						d.Index.Error.Cause.Reason,
					)
				}
			}
		}
	}
	buf.Reset()

	return true
}
