package function

import (
	"fmt"
	"os"
)

func CreateScraper(
	downloadUrl string,
	pdfPath string,
	csvPath string,
	errorPath string,
	apiKey string,
	cleanedCSVPath string,
	cleanedJsonPath string,
	dbConnectionString string) *Scraper {

	ensureDirsExist([]string{pdfPath, csvPath, errorPath, cleanedCSVPath, cleanedJsonPath})

	return &Scraper{
		DownloadUrlTemplate: downloadUrl,
		PdfPath:             pdfPath,
		CsvPath:             csvPath,
		ErrorPath:           errorPath,
		ApiKey:              apiKey,
		CleanedCSVPath:      cleanedCSVPath,
		CleanedJsonPath:     cleanedJsonPath,
		DBConnectionString:  dbConnectionString,
	}
}

type Scraper struct {
	DownloadUrlTemplate string
	PdfPath             string
	CsvPath             string
	ErrorPath           string
	ApiKey              string
	CleanedCSVPath      string
	CleanedJsonPath     string
	DBConnectionString  string
}

func ensureDirsExist(dirs []string) {
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}
