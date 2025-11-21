package parser

import (
	"path/filepath"
	"regexp"
	"strings"
)

type MovieInfo struct {
	Title string
	Year  string
}

// ParseFilename extracts movie title and year from a filename
func ParseFilename(filePath string) MovieInfo {
	// Get base filename without extension
	filename := filepath.Base(filePath)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	// Replace dots and underscores with spaces
	cleanName := strings.ReplaceAll(filename, ".", " ")
	cleanName = strings.ReplaceAll(cleanName, "_", " ")

	// Extract year (4 digits, possibly in parentheses)
	yearRegex := regexp.MustCompile(`\(?([12][0-9]{3})\)?`)
	yearMatch := yearRegex.FindStringSubmatch(cleanName)
	year := ""
	if len(yearMatch) > 1 {
		year = yearMatch[1]
	}

	// Remove year from the name
	cleanName = yearRegex.ReplaceAllString(cleanName, "")

	// Remove common quality/source indicators
	qualityRegex := regexp.MustCompile(`(?i)(720p|1080p|2160p|4k|BRRip|BDRip|BluRay|WEBRip|HDTV|WEB-DL|x264|x265|10bit|HEVC|AAC|AC3|DTS|GalaxyRG|RARBG|YTS|YIFY|\d+MB|\d+GB).*`)
	cleanName = qualityRegex.ReplaceAllString(cleanName, "")

	// Remove extra spaces and trim
	spaceRegex := regexp.MustCompile(`\s+`)
	cleanName = spaceRegex.ReplaceAllString(cleanName, " ")
	cleanName = strings.TrimSpace(cleanName)

	return MovieInfo{
		Title: cleanName,
		Year:  year,
	}
}
