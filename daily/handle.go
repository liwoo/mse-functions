package function

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/pdftables/go-pdftables-api/pkg/client"
)

func Handle(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	err := godotenv.Load()

	if err != nil {
		fmt.Println(err)
	}

	scraper := CreateScraper(
		os.Getenv("MSE_URL"),
		os.Getenv("RAW_PDF_PATH"),
		os.Getenv("RAW_CSV_PATH"),
		os.Getenv("ERROR_FILE_PATH"),
		os.Getenv("PDFTABLES_API_KEY"),
		os.Getenv("CLEANED_CSV_PATH"),
		os.Getenv("CLEANED_JSON_PATH"),
		os.Getenv("DB_CONNECTION_STRING"))

	var Client = http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	var clientCSV = client.Client{
		APIKey:     os.Getenv("API_KEY"),
		HTTPClient: http.DefaultClient,
	}
	date := time.Now()
	var current = fmt.Sprint(date.Day(), date.Month(), date.Year())
	var downloader = MSEPdfDownloader{
		FileUrl:     fmt.Sprint(scraper.DownloadUrlTemplate, current),
		FileName:    fmt.Sprint(scraper.PdfPath, current, ".pdf"),
		FileNameCSV: fmt.Sprint(scraper.CsvPath, current, ".csv"),
		Client:      Client,
		CsvClient:   &clientCSV,
	}

	// Download
	size, err := downloader.downloadPdf()
	if err != nil {
		log.Fatal(err)
	}

	err = downloader.ConvertToCSV(size)

	if err != nil {
		log.Fatal(err)
	}
	// Clean
	// Save
}
