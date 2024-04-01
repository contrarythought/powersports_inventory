package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"os"
	"rumbleon_inventory/dbupload"
	"rumbleon_inventory/scraper"
)

const (
	URL            = "https://www.rumbleon.com/buy?page=1"
	URL_TEST       = "https://scrapingclub.com/exercise/list_infinite_scroll/"
	USER_AGENT     = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
	ENDING_ELEMENT = `#Layer_1-2 > path:nth-child(3)`
)

type Creds struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

func getCreds() (Creds, error) {
	file, err := os.Open("creds.json")
	if err != nil {
		return Creds{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Creds{}, err
	}

	creds := Creds{}

	if err := json.Unmarshal(data, &creds); err != nil {
		return Creds{}, err
	}

	return creds, nil
}

func DBConnStr() (string, error) {
	creds, err := getCreds()
	if err != nil {
		return "", err
	}

	connStr := "postgres://" + creds.User + ":" + creds.Pass + "@localhost/rumbleon_inventory?sslmode=disable"

	return connStr, nil
}

func connectDB() (*sql.DB, error) {
	connStr, err := DBConnStr()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	errFile, err := os.Create("errlog.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer errFile.Close()

	errLog := log.New(errFile, "err:", log.Lshortfile|log.LstdFlags)
	errChan := make(chan error)
	defer close(errChan)

	vehicleMap, err := scraper.Scrape(URL, errChan, errLog)
	if err != nil {
		log.Fatal(err)
	}

	if err := dbupload.Upload(db, vehicleMap); err != nil {
		log.Fatal(err)
	}

	/*
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
	*/
}
