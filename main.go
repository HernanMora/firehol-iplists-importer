package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	configuration "github.com/HernanMora/firehol-iplists-importer/configs"
	logger "github.com/HernanMora/firehol-iplists-importer/logger"
	IPsetService "github.com/HernanMora/firehol-iplists-importer/service"
	CIDRUtil "github.com/HernanMora/firehol-iplists-importer/utils/net"
	CoreDNSWorker "github.com/HernanMora/firehol-iplists-importer/worker/coredns"
	CSVWorker "github.com/HernanMora/firehol-iplists-importer/worker/csv"

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

func visit(excludeSets []string, excludeCategories []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if filepath.Ext(path) != ".ipset" {
			return nil
		}
		filename := info.Name()
		if contains(excludeSets, strings.TrimSuffix(filename, filepath.Ext(filename))) {
			logger.Info("Set Excluded ", filename)
			return nil
		}
		logger.Info("Reading file: ", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("File reading error", err)
			return nil
		}

		var category string = IPsetService.GetCategory(string(data))
		if contains(excludeCategories, category) {
			logger.Info("Set Excluded ", filename, " by category ", category)
			return nil
		}

		var maintainer string = IPsetService.GetMaintainer(string(data))
		var version string = IPsetService.GetVersion(string(data))
		var ipset string = IPsetService.GetIPSet(string(data))

		ipsSets := IPsetService.ListIPs(string(data))

		if len(ipsSets) > 0 {
			var rows []map[string]string
			for _, ip := range ipsSets {
				var ipRange []string
				idx := strings.Index(ip, "/")
				if idx > -1 {
					ipRange, _ = CIDRUtil.Hosts(ip)
				} else {
					ipRange = []string{ip}
				}

				for _, ip := range ipRange {
					var values map[string]string = make(map[string]string, 4)
					values["category"] = category
					values["maintainer"] = maintainer
					values["version"] = version
					values["ipset"] = ipset
					values["ip"] = ip
					// Add rows in order from config
					rows = append(rows, values)
				}
			}
			CSVWorker.Append(rows)
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
	logger.Info("Firehol IpLists Importer")
	t := time.Now()
	csvFileName := fmt.Sprintf("%02d%02d%d_%02d%02d_firehol_ipsets.csv", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute())

	configuration, err := configuration.Load()
	if err != nil {
		panic(err)
	}

	var tempDir string = "/tmp/blocklist-ipsets"

	CSVWorker.Setup(csvFileName, configuration.CSVConfig)
	CoreDNSWorker.Setup(csvFileName, configuration.CSVConfig)

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

	err = filepath.Walk(tempDir, visit(configuration.General.ExcludeSets, configuration.General.ExcludeCategories))
	if err != nil {
		logger.Error(err.Error())
	}

	/*
		err = ClearDir(tempDir)
		if err != nil {
			logger.Error(err.Error())
		}
	*/

	CoreDNSWorker.Group()

}
