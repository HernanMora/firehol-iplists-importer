package elasticsearch

var (
	TemplateName    = "firehol-ipset-tpl"
	FireholTemplate = `
{
	"order": 0,
	"index_patterns": ["@indexName@"],
	"settings": {
		"index": {
			"codec": "best_compression"
		}
	},
	"mappings": {
		"ipset_record": {
			"properties": {
				"ip": {
					"type": "ip"
				},
				"category": {
					"type": "keyword"
				},
				"maintainer": {
					"type": "keyword"
				},
				"ipset": {
					"type": "keyword"
				}
				"version": {
					"type": "keyword"
				},
				"@timestamp": {
					"format": "epoch_millis||epoch_second||date_time||MMM dd YYYY HH:mm:ss z||MMM dd yyyy HH:mm:ss",
					"type": "date"
				}
			}
		}
	}
}
`
	DocumentType = "ipset_record"
)
