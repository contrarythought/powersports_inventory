package dbupload

import (
	"fmt"
	"log"
	"os"
	"rumbleon_inventory/scraper"
	"testing"
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

	avg, amount, err := calculateAvgPrice(vehMap, errChan)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("amount of inventory:", amount, " avg_price:", avg, " inventory value:", (avg * float64(amount)))
}
