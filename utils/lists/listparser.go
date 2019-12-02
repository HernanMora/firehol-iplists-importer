package utils

import (
	"regexp"
	"strings"
)

type ListParserUtilStruct struct{}

func NewListParserUtil() *ListParserUtilStruct {
	return &ListParserUtilStruct{}
}

var ListParserUtil *ListParserUtilStruct

func init() {
	ListParserUtil = NewListParserUtil()
}

func GetCategoryFromComment(content string) string {
	r := regexp.MustCompile(`(?m)[\r\n]^# Category\s*\:\s*(.*)`)
	res := r.FindAllStringSubmatch(content, -1)[0]
	// Remove empty lines and spaces
	content = strings.TrimSpace(res[1])
	return content
}

func GetVersionFromComment(content string) string {
	r := regexp.MustCompile(`(?m)[\r\n]^# Version\s*\:\s*(.*)`)
	res := r.FindAllStringSubmatch(content, -1)[0]
	// Remove empty lines and spaces
	content = strings.TrimSpace(res[1])
	return content
}

func GetMaintainerFromComment(content string) string {
	r := regexp.MustCompile(`(?m)[\r\n]^# Maintainer\s*\:\s*(.*)`)
	res := r.FindAllStringSubmatch(content, -1)[0]
	// Remove empty lines and spaces
	content = strings.TrimSpace(res[1])
	return content
}

func GetIPSetFromComment(content string) string {
	r := regexp.MustCompile(`(?m)[\r\n]^#  http\:\/\/iplists\.firehol\.org\/\?ipset\=(.*)`)
	res := r.FindAllStringSubmatch(content, -1)[0]
	// Remove empty lines and spaces
	content = strings.TrimSpace(res[1])
	return content
}

func RemoveComments(content string) string {
	r := regexp.MustCompile(`(?m)[\r\n]*^#.*`)
	// Remove empty lines and spaces
	content = strings.TrimSpace(r.ReplaceAllString(content, ""))
	return content
}

func ParseToList(content string) []string {
	var ipList []string
	// Fill list if has lines
	if len(content) > 0 {
		ipList = strings.Split(content, "\n")
	}

	return ipList
}
