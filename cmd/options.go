package cmd

import "strings"

const (
	FormatHtml  = "html"
	FormatJson  = "json"
	FormatXml   = "xml"
	FormatBadge = "badge"
	FormatValue = "value"
)

const (
	LevelFull    = "full"
	LevelSummary = "summary"
)

var allFormats = []string{FormatHtml, FormatJson, FormatXml, FormatBadge, FormatValue}

func AllFormats() []string {
	return allFormats
}

func AllFormatsString() string {
	return strings.Join(allFormats, ", ")
}

func IsValidFormat(value string) bool {
	for _, f := range allFormats {
		if f == value {
			return true
		}
	}

	return false
}

var allLevels = []string{LevelFull, LevelSummary}

func AllLevels() []string {
	return allLevels
}

func AllLevelsString() string {
	return strings.Join(allLevels, ", ")
}

func IsValidLevel(value string) bool {
	for _, l := range allLevels {
		if l == value {
			return true
		}
	}

	return false
}
