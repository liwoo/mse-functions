package function

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"
	"github.com/joho/godotenv"
	"github.com/pdftables/go-pdftables-api/pkg/client"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
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

	var pgconn *pgdriver.Connector = pgdriver.NewConnector(pgdriver.WithDSN(scraper.DBConnectionString))
	psdb := sql.OpenDB(pgconn)
	db := bun.NewDB(psdb, pgdialect.New())

	// TODO: Test the DB connection here before going forward

	date := time.Now()
	var current = fmt.Sprint(date.Day(), date.Month(), date.Year())
	// TODO: Scrap current file from saver
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
	var cleaner = MSECsvCleaner{
		FileUrl:      downloader.FileNameCSV,
		ErrorPath:    scraper.ErrorPath,
		CleanCsvPath: scraper.CleanedCSVPath,
	}

	data, err := cleaner.Clean()

	if err != nil {
		log.Fatal(err)
	}

	// Save
	var saver = MSESaver{
		FileUrl:   data.file,
		ErrorPath: scraper.ErrorPath,
		Db:        db,
	}

	saver.Save()
}
