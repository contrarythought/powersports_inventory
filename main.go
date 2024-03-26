package main

import (
	"fmt"
	"log"
	"os"
	"rumbleon_inventory/scraper"
)

const (
	URL            = "https://www.rumbleon.com/buy?page=1"
	URL_TEST       = "https://scrapingclub.com/exercise/list_infinite_scroll/"
	USER_AGENT     = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
	ENDING_ELEMENT = `#Layer_1-2 > path:nth-child(3)`
)

func main() {
	vehicleMap, err := scraper.Scrape(URL)
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create("test_out_file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	numInventory := 0
	for k, v := range vehicleMap {
		fmt.Fprintln(outFile, k)
		for _, veh := range v {
			numInventory++
			fmt.Fprintln(outFile, "\t", veh.Brand, "-->", veh.Model, "-->", veh.Price)
		}
		fmt.Fprintln(outFile)
	}
}
