package dbupload

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"rumbleon_inventory/errorhandling"
	"rumbleon_inventory/scraper"
	"strconv"
	"sync"
	"time"
)

func Upload(db *sql.DB, vehicleMap map[scraper.Brand][]scraper.Vehicle, totalInventory int, errChan chan error, errLog *log.Logger) error {
	fmt.Println("uploading data into DB...")
	time.Sleep(time.Second * 2)

	go errorhandling.ErrorResolver(errChan, errLog, 3)

	day := time.Now()
	avg_price, amount, err := calculateAvgPrice(vehicleMap, errChan)
	if err != nil {
		return err
	}

	fmt.Println("number of vehicles scraped: ", amount)
	time.Sleep(time.Second * 1)

	inventoryValue := avg_price * float64(amount)

	queryStr := `INSERT INTO timeseries.inventory_tracker (timestamp, inventory_count, avg_price, inventory_value_estimate) VALUES ($1, $2, $3, $4)`

	res, err := db.Exec(queryStr, day, totalInventory, avg_price, inventoryValue)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows <= 0 {
		return errors.New("err: failed to insert data")
	}

	return nil
}

func calculateAvgPrice(vehicleMap map[scraper.Brand][]scraper.Vehicle, errChan chan error) (float64, int, error) {
	amount := 0
	sum := 0.0

	vehSliceChan := make(chan []scraper.Vehicle)
	defer close(vehSliceChan)

	var wg sync.WaitGroup
	var mu sync.Mutex
	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < len(vehicleMap); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case vehicles := <-vehSliceChan:
					mu.Lock()
					for _, v := range vehicles {
						amount++
						price, err := strconv.ParseFloat(v.Price, 64)
						if err != nil {
							errChan <- err
						}
						sum += price
					}
					mu.Unlock()
				case <-ctx.Done():
					if len(vehSliceChan) == 0 {
						return
					}
				}
			}
		}()
	}

	for _, vehicles := range vehicleMap {
		vehSliceChan <- vehicles
	}

	cancel()
	wg.Wait()

	/*
		for _, vehicles := range vehicleMap {
			for _, v := range vehicles {
				amount++
				price, err := strconv.ParseFloat(v.Price, 64)
				if err != nil {
					return -1, err
				}
				sum += price
			}
		}
	*/

	avg := sum / float64(amount)

	return avg, amount, nil
}
