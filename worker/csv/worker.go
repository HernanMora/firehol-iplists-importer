package csv

import (
	"path"

	"github.com/HernanMora/firehol-iplists-importer/configs"
	"github.com/HernanMora/firehol-iplists-importer/service/csv"
)

type CSVWorker struct {
	config     configs.CSVSettings
	csvService csv.CSVService
}

func NewCSVWorker() *CSVWorker {
	return &CSVWorker{}
}

var worker *CSVWorker

func init() {
	worker = NewCSVWorker()
}

func Setup(csvFileName string, configuration configs.CSVSettings) {
	worker.csvService = csv.NewCSVService(path.Join([]string{configuration.OutputFolder, csvFileName}...), configuration)
}

func Append(lines []map[string]string) {
	worker.csvService.Append(lines)
}
