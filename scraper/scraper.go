package scraper

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"rumbleon_inventory/errorhandling"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Brand string

type Vehicle struct {
	Brand string
	Model string
	Price string
}

const (
	// USER_AGENT       = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
	WAIT_ELEMENT     = `#Layer_2`
	MAX_PAGE_ELE_SEL = `#root > div > section > main > div.css-15g0dol-Base.e1n4b2jv0 > div:nth-child(6) > div.css-l0mhay-emotion--Pagination--SearchPagination > a:nth-child(4)`
	TOTAL_INVENTORY  = `#root > div > section > main > div.css-15g0dol-Base.e1n4b2jv0 > div.ant-row.css-1qi96nr-emotion--BuyPage--BuyPage > div.ant-col.ant-col-xs-24.ant-col-md-12 > div > div:nth-child(2) > h4`
	NUM_WORKERS      = 1 // TODO: change this
)

// sets up the process of scraping vehicles (grabs max page to loop through)
func Scrape(url string, errChan chan error, errLog *log.Logger) (map[Brand][]Vehicle, int, error) {
	ret := make(map[Brand][]Vehicle)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(GrabUserAgent()),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// start thread to handle concurrent errors
	go errorhandling.ErrorResolver(errChan, errLog, 3)

	var maxpages, totalInventory string
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(WAIT_ELEMENT),
		chromedp.Text(MAX_PAGE_ELE_SEL, &maxpages, chromedp.ByQuery),
		chromedp.Text(TOTAL_INVENTORY, &totalInventory, chromedp.ByQuery),
	); err != nil {
		return nil, -1, err
	}

	_, err := strconv.Atoi(maxpages)
	if err != nil {
		return nil, -1, err
	}

	inventory, err := convertInventoryEle(totalInventory)
	if err != nil {
		return nil, -1, err
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	// figure out how to buffer channel of urls to create less work for the server

	scrapeCtx, scrapeCancel := context.WithCancel(context.Background())

	vehicles := []Vehicle{}
	urlChan := make(chan string, NUM_WORKERS)
	defer close(urlChan)

	// start workers to take in url and scrape it
	for i := 0; i < NUM_WORKERS; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case url := <-urlChan:
					fmt.Println("url received:", url)
					vehs, err := ScrapeInventory(url, opts, &taskCtx)
					if err != nil {
						errChan <- err
					}
					fmt.Println("vehicle list from scrapeInventory:", vehs)

					mu.Lock()
					vehicles = append(vehicles, vehs...)
					mu.Unlock()

					time.Sleep(2 * time.Second)
				case <-scrapeCtx.Done():
					if len(urlChan) == 0 {
						return
					}
				}
			}
		}()
	}

	origLen := len(url)
	// TODO: change for loop back to max
	for i := 0; i < 5; i++ {
		url = url[:origLen-1] + strconv.Itoa(i+1)
		fmt.Println("sending url:", url)
		time.Sleep(time.Second * 1)

		urlChan <- url
	}

	scrapeCancel()
	wg.Wait()

	for _, v := range vehicles {
		ret[Brand(v.Brand)] = append(ret[Brand(v.Brand)], v)
	}

	return ret, inventory, nil
}

func convertInventoryEle(inventoryStr string) (int, error) {
	arr := strings.Split(inventoryStr, " ")
	invStrTrim := strings.TrimSpace(arr[0])
	noComma := strings.ReplaceAll(invStrTrim, ",", "")
	invInt, err := strconv.Atoi(noComma)
	if err != nil {
		return -1, err
	}

	return invInt, nil
}

func GrabUserAgent() string {
	userAgents := []string{
		`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`,
	}

	idx := rand.Intn(len(userAgents))

	return userAgents[idx]
}

const (
	URL      = "https://www.rumbleon.com/buy?page=1"
	URL_TEST = "https://scrapingclub.com/exercise/list_infinite_scroll/"
	// USER_AGENT     = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
	ENDING_ELEMENT = `#Layer_1-2 > path:nth-child(3)`
)

// TODO
// grabs vehicles from the url
func ScrapeInventory(url string, opts []chromedp.ExecAllocatorOption, taskCtx *context.Context) ([]Vehicle, error) {
	fmt.Println("scraping:", url)
	time.Sleep(time.Second * 2)

	ret := []Vehicle{}

	/*
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()
	*/

	/*
		taskCtx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()
	*/

	var nodes []*cdp.Node

	// <div class="ant-card css-1gl0vip-emotion--Result--cardCss ant-card-bordered ant-card-hoverable">
	err := chromedp.Run(*taskCtx,
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
		err = chromedp.Run(*taskCtx,
			chromedp.Text(`div.ant-card-body > div > div:nth-child(2) > span`, &brand, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(`div.ant-card-body > div > div:nth-child(3) > span`, &model, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(`span.ant-typography > strong`, &price, chromedp.ByQuery, chromedp.FromNode(node)),
		)
		if err != nil {
			log.Fatal(err)
		}

		price = strings.ReplaceAll(price[1:], ",", "")

		ret = append(ret, Vehicle{
			Brand: brand,
			Model: model,
			Price: price,
		})

		fmt.Println(brand)
		fmt.Println(model)
		fmt.Println(price)
		fmt.Println()
	}

	return ret, nil
}
