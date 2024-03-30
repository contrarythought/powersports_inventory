package dbupload

import (
	"database/sql"
	"fmt"
	"rumbleon_inventory/scraper"
	"time"
)

func Upload(db *sql.DB, vehicleMap map[scraper.Brand][]scraper.Vehicle) error {
	day := time.Now()
	fmt.Println(day)
	return nil
}
