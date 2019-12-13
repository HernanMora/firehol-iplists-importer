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

func NewElasticService(configuration configs.ElasticsearchSettings) ElasticsearchService {
	logger.Info("Setting up Elasticsearch Service...")

	cfg := elasticsearch.Config{
		Addresses: []string{
			configuration.Url,
		},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 60 * time.Second,
		},
	}

	esClient, _ := elasticsearch.NewClient(cfg)

	return ElasticsearchService{
		esClient:  esClient,
		indexName: configuration.IndexName,
	}
}

func (es *ElasticsearchService) CheckTemplate() {
	logger.Infof("Checking template %s from elasticsearch", elasticsConf.TemplateName)
	response, _ := es.esClient.Indices.ExistsTemplate([]string{elasticsConf.TemplateName})
	if response.IsError() && response.StatusCode != 404 {
		logger.Errorf("Error on check index template %s from elasticsearch", elasticsConf.TemplateName)
	} else if response.IsError() && response.StatusCode == 404 {
		logger.Infof("Index template %s not found... Creating now", elasticsConf.TemplateName)
		template := strings.Replace(elasticsConf.FireholTemplate, "@indexName@", es.indexName, -1)
		var buffer bytes.Buffer
		templateBody := []byte(fmt.Sprintf("%s\n", template))
		buffer.Grow(len(templateBody))
		buffer.Write(templateBody)
		response, err := es.esClient.Indices.PutTemplate(elasticsConf.TemplateName, bytes.NewReader(buffer.Bytes()))
		if err != nil {
			logger.Errorf("Cannot create index template %s: %s", elasticsConf.TemplateName, err)
		}
		if response.IsError() {
			logger.Errorf("Cannot create index template %s: %s", elasticsConf.TemplateName, response)
		}
		time.Sleep(1 * time.Second)
	}
}

func (es *ElasticsearchService) removeOldDocs(index string, timestamp int64) {

	var buffer bytes.Buffer
	var raw map[string]interface{}
	deleteBody := []byte(fmt.Sprintf(`{"query": { "range": { "timestamp": { "lt" : "%d" } } } }\n`, timestamp))
	buffer.Grow(len(deleteBody))
	buffer.Write(deleteBody)
	response, err := es.esClient.DeleteByQuery([]string{index}, bytes.NewReader(buffer.Bytes()))
	if err != nil {
		logger.Errorf("Cannot delete old rules from index %s: %s", index, err)
	}
	if response.IsError() {
		if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
			logger.Fatalf("Failure to to parse response body: %s", err)
		} else {
			logger.Errorf("Error: %v", raw)
		}
	}

}

func (es *ElasticsearchService) InsertToElasticsearch(docs []models.Document, timestamp int64) {
	response, _ := es.esClient.Indices.Exists([]string{es.indexName})
	if response.IsError() && response.StatusCode != 404 {
		logger.Errorf("Error on check index %s from elasticsearch", es.indexName)
	} else if response.IsError() && response.StatusCode == 404 {
		logger.Infof("Index %s not found... Creating now", es.indexName)
		response, err := es.esClient.Indices.Create(es.indexName)
		if err != nil {
			logger.Errorf("Cannot create index %s: %s", es.indexName, err)
		}
		if response.IsError() {
			logger.Errorf("Cannot create index %s: %s", es.indexName, response)
		}
		time.Sleep(2 * time.Second)
	}

	//ok := es.insertRulesBulk(es.indexName, elasticsConf.DocumentType, docs)
	es.insertRulesBulk(es.indexName, elasticsConf.DocumentType, docs)
	/*
		if ok {
			// Needed to avoid version conflicts of docs
			time.Sleep(1 * time.Second)
			es.removeOldDocs(es.indexName, timestamp)
		}
	*/
}

func (es *ElasticsearchService) insertRulesBulk(index string, indexType string, docs []models.Document) bool {

	var maxSteps = 2500
	var step int
	for len(docs) > 0 {
		if len(docs) < maxSteps {
			step = len(docs)
		} else {
			step = maxSteps
		}
		bulkDocs := docs[:step]

		var (
			buf             bytes.Buffer
			elasticResponse *esapi.Response
			raw             map[string]interface{}
			blk             *bulkResponse
		)

		for _, doc := range bulkDocs {
			var ip string
			if value, ok := doc["network"]; ok {
				ip = value[:strings.Index(value, "/")]
			} else {
				ip = doc["ip"]
			}
			id := fmt.Sprintf("%s_%s", doc["ipset"], ip)
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

		elasticResponse, err := es.esClient.Bulk(bytes.NewReader(buf.Bytes()), es.esClient.Bulk.WithIndex(index))
		if err != nil {
			logger.Fatalf("Failure indexing batch: %s", err)
		}
		if elasticResponse.IsError() {
			if err := json.NewDecoder(elasticResponse.Body).Decode(&raw); err != nil {
				logger.Fatalf("Failure to to parse response body: %s", err)
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
				logger.Fatalf("Failure to to parse response body: %s", err)
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

		docs = docs[step:]
	}

	return true
}
