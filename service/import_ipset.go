package service

import (
	ListParserUtil "github.com/HernanMora/firehol-iplists-importer/utils/lists"
)

type CSVService struct{}

func NewIPsetService() *CSVService {
	return &CSVService{}
}

var IPsetService *CSVService

func init() {
	IPsetService = NewIPsetService()
}

func ListIPs(content string) []string {
	var list []string
	list = ListParserUtil.ParseToList(ListParserUtil.RemoveComments(content))
	return list
}

func GetCategory(content string) string {
	var category string
	category = ListParserUtil.GetCategoryFromComment(content)
	return category
}

func GetVersion(content string) string {
	var version string
	version = ListParserUtil.GetVersionFromComment(content)
	return version
}

func GetMaintainer(content string) string {
	var maintainer string
	maintainer = ListParserUtil.GetMaintainerFromComment(content)
	return maintainer
}

func GetIPSet(content string) string {
	var ipset string
	ipset = ListParserUtil.GetIPSetFromComment(content)
	return ipset
}
