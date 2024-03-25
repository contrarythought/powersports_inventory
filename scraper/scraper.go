package scraper

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/chromedp/chromedp"
)

type brand string

type Vehicle struct {
	Brand string
	Model string
	Price string
}

const (
	USER_AGENT       = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
	WAIT_ELEMENT     = `#Layer_2`
	MAX_PAGE_ELE_SEL = `#root > div > section > main > div.css-15g0dol-Base.e1n4b2jv0 > div:nth-child(6) > div.css-l0mhay-emotion--Pagination--SearchPagination > a:nth-child(4)`
)

func errorResolver(errChan <-chan error, errLog *log.Logger, errLmt int) {
	numErr := 0
	for err := range errChan {
		numErr++
		if numErr >= errLmt {
			log.Fatal("something really wrong...check logs")
		}

		errLog.Println(err)
	}
}

// sets up the process of scraping vehicles (grabs max page to loop through)
func Scrape(url string) (map[brand][]Vehicle, error) {
	ret := make(map[brand][]Vehicle)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(USER_AGENT),
		chromedp.Headless,
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	errLogFile, err := os.Create("errLog")
	if err != nil {
		return nil, err
	}
	defer errLogFile.Close()

	// create logger
	errLog := log.New(errLogFile, "err:", log.Lshortfile|log.LstdFlags)

	// error channel
	errChan := make(chan error)

	go errorResolver(errChan, errLog, 3)

	var maxpages string
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(WAIT_ELEMENT),
		chromedp.Text(MAX_PAGE_ELE_SEL, &maxpages, chromedp.ByQuery),
	); err != nil {
		return nil, err
	}

	max, err := strconv.Atoi(maxpages)
	if err != nil {
		return nil, err
	}

	vehicles := []Vehicle{}
	for i := 0; i < max; i++ {
		go func(i int) {
			url = url[:len(url)-1] + strconv.Itoa(i+1)
			veh, err := scrapeInventory(url, opts)
			if err != nil {
				errChan <- err
			}
		}(i)
		url = url[:len(url)-1] + strconv.Itoa(i+1)
		veh, err := scrapeInventory(url, opts)
		if err != nil {
			return nil, err
		}
		vehicles = append(vehicles, veh...)
	}

	for _, v := range vehicles {
		ret[brand(v.Brand)] = append(ret[brand(v.Brand)], v)
	}

	return ret, nil
}

// TODO
// grabs vehicles from the url
func scrapeInventory(url string, opts []chromedp.ExecAllocatorOption) ([]Vehicle, error) {
	ret := []Vehicle{}

	return ret, nil
}
