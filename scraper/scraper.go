package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/chromedp/cdproto/cdp"
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

	// start thread to handle concurrent errors
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

	var mu sync.Mutex
	var wg sync.WaitGroup

	// figure out how to buffer channel of urls to create less work for the server

	vehicles := []Vehicle{}
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func(i int) {
			wg.Done()
			url = url[:len(url)-1] + strconv.Itoa(i+1)
			veh, err := scrapeInventory(url, opts)
			if err != nil {
				errChan <- err
			}

			mu.Lock()
			vehicles = append(vehicles, veh...)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	for _, v := range vehicles {
		ret[brand(v.Brand)] = append(ret[brand(v.Brand)], v)
	}

	return ret, nil
}

const (
	URL      = "https://www.rumbleon.com/buy?page=1"
	URL_TEST = "https://scrapingclub.com/exercise/list_infinite_scroll/"
	// USER_AGENT     = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
	ENDING_ELEMENT = `#Layer_1-2 > path:nth-child(3)`
)

// TODO
// grabs vehicles from the url
func scrapeInventory(url string, opts []chromedp.ExecAllocatorOption) ([]Vehicle, error) {
	ret := []Vehicle{}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var nodes []*cdp.Node

	// <div class="ant-card css-1gl0vip-emotion--Result--cardCss ant-card-bordered ant-card-hoverable">
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(ENDING_ELEMENT),
		chromedp.Nodes(".ant-card-body", &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		log.Fatal(err)
	}

	// brand:
	// div.ant-card-body > div > div:nth-child(2) > span

	// model:
	// div.ant-card-body > div > div:nth-child(3) > span

	// price:
	// div:nth-child(1) > span > strong

	var brand, model, price string
	for _, node := range nodes {
		err = chromedp.Run(taskCtx,
			chromedp.Text(`div.ant-card-body > div > div:nth-child(2) > span`, &brand, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(`div.ant-card-body > div > div:nth-child(3) > span`, &model, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(`span.ant-typography > strong`, &price, chromedp.ByQuery, chromedp.FromNode(node)),
		)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(brand)
		fmt.Println(model)
		fmt.Println(price)
		fmt.Println()
	}

	return ret, nil
}
