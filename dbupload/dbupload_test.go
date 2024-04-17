package dbupload

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"rumbleon_inventory/scraper"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestTime(t *testing.T) {
	if err := Upload(nil, nil, 0, nil, nil); err != nil {
		t.Error(err)
	}
}

func TestAverage(t *testing.T) {
	errChan := make(chan error)
	defer close(errChan)

	testErrFile, err := os.Create("testErrFile.txt")
	if err != nil {
		t.Error(err)
	}
	defer testErrFile.Close()

	errLog := log.New(testErrFile, "err:", log.Lshortfile|log.LstdFlags)

	vehMap, totalInventory, err := scraper.Scrape(scraper.URL, errChan, errLog)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("vehicleMap:", vehMap)

	avg, _, err := calculateAvgPrice(vehMap, errChan)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("amount of inventory:", totalInventory, " avg_price:", avg, " inventory value:", (avg * float64(totalInventory)))
}

// it works
func TestUpload(t *testing.T) {
	credFile, err := os.Open(``)
	if err != nil {
		t.Error(err)
	}
	defer credFile.Close()

	data, err := io.ReadAll(credFile)
	if err != nil {
		t.Error(err)
	}

	creds := struct {
		User string `json:"user"`
		Pass string `json:"pass"`
	}{}

	if err := json.Unmarshal(data, &creds); err != nil {
		t.Error(err)
	}

	connStr := "postgres://" + creds.User + ":" + creds.Pass + "@localhost/rumbleon_inventory?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Error(err)
	}

	day := time.Now()
	testVehMap := map[scraper.Brand][]scraper.Vehicle{
		"ford": {
			{
				Brand: "ford",
				Model: "f150",
				Price: "40000",
			},
			{
				Brand: "ford",
				Model: "fiesta",
				Price: "20000",
			},
		},
		"toyota": {
			{
				Brand: "toyota",
				Model: "corolla",
				Price: "30000",
			},
			{
				Brand: "toyota",
				Model: "tundra",
				Price: "40000",
			},
		},
	}

	errChan := make(chan error)

	avgPrice, amount, err := calculateAvgPrice(testVehMap, errChan)
	if err != nil {
		t.Error(err)
	}

	queryStr := `INSERT INTO timeseries.inventory_tracker (timestamp, inventory_count, avg_price, inventory_value_estimate) VALUES ($1, $2, $3, $4)`

	res, err := db.Exec(queryStr, day, amount, avgPrice, (avgPrice * float64(amount)))
	if err != nil {
		t.Error(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		t.Error(err)
	}

	if rows <= 0 {
		t.Error(errors.New("err: failed to insert data"))
	}
}
