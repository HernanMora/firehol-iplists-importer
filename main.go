package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	configuration "github.com/HernanMora/firehol-iplists-importer/configs"
	logger "github.com/HernanMora/firehol-iplists-importer/logger"
	"github.com/HernanMora/firehol-iplists-importer/models"
	IPsetService "github.com/HernanMora/firehol-iplists-importer/service"
	ESWorker "github.com/HernanMora/firehol-iplists-importer/worker/elasticsearch"
	"gopkg.in/src-d/go-git.v4"
)

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func visit(sets []string, skip []string, timestamp int64) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Fatal(err.Error())
		}
		if (filepath.Ext(path) != ".ipset") && (filepath.Ext(path) != ".netset") {
			return nil
		}

		filename := info.Name()
		if !contains(sets, strings.TrimSuffix(filename, filepath.Ext(filename))) {
			logger.Info("Set Excluded ", filename)
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			logger.Error(err.Error())
			fmt.Println("File reading error", err)
			return nil
		}

		var category string = IPsetService.GetCategory(string(data))
		var maintainer string = IPsetService.GetMaintainer(string(data))
		var version string = IPsetService.GetVersion(string(data))
		var ipset string = IPsetService.GetIPSet(string(data))

		list := IPsetService.ListIPs(string(data))

		if len(list) > 0 {
			var rows []models.Document
			for _, ip := range list {
				// Excluded entries
				if contains(skip, ip) {
					return nil
				}
				var values models.Document = make(models.Document, 6)
				values["category"] = category
				values["maintainer"] = maintainer
				values["version"] = version
				values["ipset"] = ipset
				values["timestamp"] = strconv.FormatInt(timestamp, 10)
				idx := strings.Index(ip, "/")
				if idx > -1 {
					values["network"] = ip
				} else {
					values["ip"] = ip
				}
				rows = append(rows, values)
			}
			ESWorker.IndexDocuments(rows, timestamp)
		}

		return nil
	}
}

func ClearDir(dir string) error {
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		names, err := ioutil.ReadDir(dir)

		if err != nil {
			return err
		}
		for _, entery := range names {
			os.RemoveAll(path.Join([]string{dir, entery.Name()}...))
		}

		os.Remove(dir)
	}
	return nil
}

func main() {
	configuration, err := configuration.Load()
	if err != nil {
		panic(err)
	}

	logger.Setup(configuration.General.Logfile)
	logger.Info("Firehol IpLists Importer")
	now := time.Now()

	ESWorker.Setup(configuration.ElasticConfig)

	var tempDir string = "/tmp/blocklist-ipsets"

	err = ClearDir(tempDir)
	if err != nil {
		logger.Error(err.Error())
	}

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      configuration.General.Repository,
		Progress: os.Stdout,
	})

	if err != nil {
		logger.Error(err.Error())
	}

	err = filepath.Walk(tempDir, visit(configuration.General.Sets, configuration.General.Skip, now.Unix()*1000))
	if err != nil {
		logger.Error(err.Error())
	}

	err = ClearDir(tempDir)
	if err != nil {
		logger.Error(err.Error())
	}

}
