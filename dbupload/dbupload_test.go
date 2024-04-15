package dbupload

import (
	"fmt"
	"log"
	"os"
	"rumbleon_inventory/scraper"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	if err := Upload(nil, nil, nil, nil); err != nil {
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

	vehMap, err := scraper.Scrape(scraper.URL, errChan, errLog)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("vehicleMap:", vehMap)

	avg, amount, err := calculateAvgPrice(vehMap, errChan)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("amount of inventory:", amount, " avg_price:", avg, " inventory value:", (avg * float64(amount)))
}

// TODO
func TestUpload(t *testing.T) {
	day := time.Now()
	testVehMap := map[scraper.Brand][]scraper.Vehicle{
		"ford": []scraper.Vehicle{
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
		"toyota": []scraper.Vehicle{
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

}
