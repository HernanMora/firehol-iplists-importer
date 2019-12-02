package csv

import (
	"os"

	encodingCSV "encoding/csv"

	"github.com/HernanMora/firehol-iplists-importer/configs"
	logger "github.com/HernanMora/firehol-iplists-importer/logger"
)

type CSVService struct {
	filename string
	settings configs.CSVSettings
}

func NewCSVService(csvFileName string, settings configs.CSVSettings) CSVService {
	return CSVService{
		filename: csvFileName,
		settings: settings,
	}
}

func (csv *CSVService) Append(lines []map[string]string) {

	file, err := os.OpenFile(csv.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		logger.Error("File ", csv.filename, " could not be opened")
		os.Exit(1)
	}
	var rows [][]string

	for _, line := range lines {
		row := []string{}
		for _, field := range csv.settings.Fields {
			row = append(row, line[field])
		}
		rows = append(rows, row)
	}

	logger.Info("Total records: ", len(rows))
	csvWriter := encodingCSV.NewWriter(file)
	csvWriter.Comma = []rune(csv.settings.Comma)[0]
	err = csvWriter.WriteAll(rows)
	if err != nil {
		logger.Error(err.Error())
	}
	csvWriter.Flush()

}
