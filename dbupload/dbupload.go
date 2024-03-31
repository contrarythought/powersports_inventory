package dbupload

import (
	"context"
	"database/sql"
	"rumbleon_inventory/scraper"
	"strconv"
	"sync"
	"time"
)

func Upload(db *sql.DB, vehicleMap map[scraper.Brand][]scraper.Vehicle) error {
	day := time.Now()
	avg_price, err := calculateAvgPrice(vehicleMap)
	if err != nil {
		return err
	}

	return nil
}

// TODO: make multi-threaded
func calculateAvgPrice(vehicleMap map[scraper.Brand][]scraper.Vehicle) (float64, error) {
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
							// create general error handling package
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

	return avg, nil
}
