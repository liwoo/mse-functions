package function

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/oiime/logrusbun"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Handle an HTTP Request.
func Handle(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	err := godotenv.Load()

	if err != nil {
		fmt.Println(err)
	}

	var pgconn *pgdriver.Connector = pgdriver.NewConnector(pgdriver.WithDSN(os.Getenv("DB_CONNECTION_STRING")))
	psdb := sql.OpenDB(pgconn)
	db := bun.NewDB(psdb, pgdialect.New())

	// TODO: Remove in production
	log := logrus.New()
	log.Level = logrus.DebugLevel
	db.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{
		LogSlow:    time.Second,
		Logger:     log,
		QueryLevel: logrus.DebugLevel,
		ErrorLevel: logrus.ErrorLevel,
		SlowLevel:  logrus.WarnLevel,
	}))

	data, _ := generateGraphs(db)

	jsnStr, _ := json.Marshal(data)
	fmt.Println(string(jsnStr))
	saveStats(db, data)
}

func prettyPrint(req *http.Request) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "%v %v %v %v\n", req.Method, req.URL, req.Proto, req.Host)
	for k, vv := range req.Header {
		for _, v := range vv {
			fmt.Fprintf(b, "  %v: %v\n", k, v)
		}
	}

	if req.Method == "POST" {
		req.ParseForm()
		fmt.Fprintln(b, "Body:")
		for k, v := range req.Form {
			fmt.Fprintf(b, "  %v: %v\n", k, v)
		}
	}

	return b.String()
}

func generateGraphs(db *bun.DB) ([]CompanyStat, error) {
	points := [7]RangePoint{
		{years: 0, months: 0, days: 7, exclude: 0, name: "WEEKLY"},
		{years: 0, months: 1, days: 0, exclude: 0, name: "MONTHLY"},
		{years: 0, months: 3, days: 0, exclude: 2, name: "THREEMONTHS"},
		{years: 0, months: 6, days: 0, exclude: 2, name: "SIXMONTHS"},
		{years: 1, months: 0, days: 0, exclude: 5, name: "YEARLY"},
		{years: 2, months: 0, days: 0, exclude: 10, name: "TWOYEARS"},
		{years: 3, months: 0, days: 0, exclude: 10, name: "THREEYEARS"},
	}

	outputSlice := make(map[string](map[string]Coord))

	for _, point := range points {
		data, err := getModelsInRage(db, point)

		if err != nil {
			fmt.Println(err)
		} else {
			outputSlice[point.name] = data
		}
	}
	// jsnStr, _ := json.Marshal(outputSlice)
	// fmt.Println(string(jsnStr))
	return MergeMap(outputSlice), nil
}

func getModelsInRage(db *bun.DB, point RangePoint) (map[string]Coord, error) {
	timeRange := getDateRange(point)
	var rates []DailyCompanyRateModel

	if err := db.NewSelect().Model(&rates).Where("date IN (?)", bun.In(timeRange)).OrderExpr("date").Scan(context.Background()); err != nil {
		return nil, err
	}

	// TODO: Process in Goroutine
	output := make(map[string]Coord)

	for _, rate := range rates {
		data, exists := output[rate.CODE]
		// TODO: Generate coordinate
		coordinate := []int{0, int(rate.BUY)}
		if exists {
			coordinate[0] = (len(data) - 1) + 1
			data[rate.DATE] = coordinate
			output[rate.CODE] = data
		} else {
			output[rate.CODE] = map[string][]int{
				rate.DATE: coordinate,
			}
		}
	}

	return output, nil
}

func getDateRange(point RangePoint) []time.Time {
	// TODO: Check if weekends are also traded so that those can be ignored in the calculation
	var timeRange []time.Time
	end := time.Now()
	start := end.AddDate(-point.years, -point.months, -point.days)
	var lastTime = start
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		if point.exclude > 0 {
			var days = lastTime.Sub(d) / (24 * time.Hour)
			if int(days) == -point.exclude {
				timeRange = append(timeRange, d)
				lastTime = d
			}
		} else {
			timeRange = append(timeRange, d)
		}
	}
	return timeRange
}

func MergeMap(data map[string](map[string]Coord)) []CompanyStat {
	output := make(map[string]CompanyStat)

	for name, m := range data {
		for k, v := range m {
			dd, exists := output[k]
			if exists {
				dd.MONTHLY = v
				reflect.ValueOf(&dd).Elem().FieldByName(name).Set(reflect.ValueOf(v))
				output[k] = dd
			} else {
				dddd := CompanyStat{
					STOCK: k,
					DATE:  time.Now().Format("2006/01/02"),
				}
				reflect.ValueOf(&dddd).Elem().FieldByName(name).Set(reflect.ValueOf(v))
				output[k] = dddd
			}
		}
	}
	return maps.Values(output)
}

func saveStats(db *bun.DB, data []CompanyStat) {
	res, err := db.NewInsert().Model(&data).Exec(context.Background())
	if err != nil {
		fmt.Println("Error saving codes:", err)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Finished adding codes", affected)
}
