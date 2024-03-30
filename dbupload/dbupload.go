package dbupload

import (
	"database/sql"
	"rumbleon_inventory/scraper"
	"strconv"
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

	avg := sum / float64(amount)

	return avg, nil
}
