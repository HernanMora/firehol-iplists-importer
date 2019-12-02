package coredns

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/HernanMora/firehol-iplists-importer/configs"
	logger "github.com/HernanMora/firehol-iplists-importer/logger"
)

type CoreDNSWorker struct {
	filename      string
	configuration configs.CSVSettings
}

func NewCoreDNSWorker() *CoreDNSWorker {
	return &CoreDNSWorker{}
}

var worker *CoreDNSWorker

func init() {
	worker = NewCoreDNSWorker()
}

func Setup(csvFileName string, configuration configs.CSVSettings) {
	worker.filename = path.Join([]string{configuration.OutputFolder, csvFileName}...)
	worker.configuration = configuration

}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func Group() {
	csvFile, _ := os.Open(worker.filename)
	reader := csv.NewReader(bufio.NewReader(csvFile))
	reader.Comma = []rune(worker.configuration.Comma)[0]
	var records map[string][]string = make(map[string][]string)

	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		var key string = line[0]

		if values, ok := records[key]; ok {
			if !contains(values, line[1]) {
				records[key] = append(records[key], line[1])
			}
		} else {
			records[key] = append(records[key], line[1])
		}
	}

	corednsFile := path.Join([]string{worker.configuration.OutputFolder, "hosts.blacklist"}...)
	file, err := os.OpenFile(corednsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		logger.Error("File ", corednsFile, " could not be opened")
		os.Exit(1)
	}

	csvWriter := csv.NewWriter(file)
	csvWriter.Comma = []rune("\t")[0]
	for key, values := range records {
		var ip string = key
		var categories string = strings.Join(values[:], ".")
		err = csvWriter.WriteAll([][]string{{ip, categories}})
		if err != nil {
			logger.Error(err.Error())
		}
	}
	csvWriter.Flush()
}
