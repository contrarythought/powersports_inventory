package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Product struct {
	name, price string
}

func TestScrape(t *testing.T) {
	testLog, err := os.Create("errLogTest.txt")
	if err != nil {
		t.Error(err)
	}
	defer testLog.Close()

	errChan := make(chan error)
	defer close(errChan)

	errLog := log.New(testLog, "err:", log.Lshortfile|log.LstdFlags)

	vehicleMap, _, err := Scrape(URL, errChan, errLog)
	if err != nil {
		t.Error(err)
	}

	output, err := os.Create("testoutput.txt")
	if err != nil {
		t.Error(err)
	}
	defer output.Close()

	for br, vehs := range vehicleMap {
		fmt.Fprintln(output, br)
		for _, v := range vehs {
			fmt.Fprintln(output, "\t", v.Model, "-->", v.Price)
		}
	}

}

func TestScraper(t *testing.T) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`),
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	var nodes []*cdp.Node
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(URL),
		chromedp.WaitVisible(ENDING_ELEMENT),
		chromedp.Nodes(".ant-card-body", &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		t.Error(err)
	}

	var brand, model, price string
	for _, node := range nodes {
		err = chromedp.Run(taskCtx,
			chromedp.Text(`div.ant-card-body > div > div:nth-child(2) > span`, &brand, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(`div.ant-card-body > div > div:nth-child(3) > span`, &model, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(`span.ant-typography > strong`, &price, chromedp.ByQuery, chromedp.FromNode(node)),
		)
		if err != nil {
			t.Error(err)
		}

		fmt.Println(brand)
		fmt.Println(model)
		fmt.Println(price)
		fmt.Println()
	}
}

func TestURL(t *testing.T) {
	url := URL
	origLen := len(url)

	for i := 0; i < 200; i++ {
		url = url[:origLen-1] + strconv.Itoa(i+1)
		fmt.Println(url)
	}
}

func TestConcur(t *testing.T) {
	maxNum := 50
	maxWork := 10
	numChan := make(chan int, maxWork)
	defer close(numChan)

	validate := [50]int{}

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < maxWork; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case num := <-numChan:
					fmt.Println(num)

					mu.Lock()
					validate[num-1] = num
					mu.Unlock()

				case <-ctx.Done():
					if len(numChan) == 0 {
						return
					}
				}
			}
		}()
	}

	for i := 1; i <= maxNum; i++ {
		numChan <- i
	}

	cancel()
	wg.Wait()

	fmt.Println("finished!")
	fmt.Println("validating:")

	for _, v := range validate {
		fmt.Println(v)
	}

}

func TestMaxPage(t *testing.T) {
	url := `https://www.rumbleon.com/buy?page=1`

	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(GrabUserAgent()),
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	var maxpages, totalInventory string
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(WAIT_ELEMENT),
		chromedp.Text(MAX_PAGE_ELE_SEL, &maxpages, chromedp.ByQuery),
		chromedp.Text(TOTAL_INVENTORY, &totalInventory, chromedp.ByQuery),
	); err != nil {
		t.Error(err)
	}

	max, err := strconv.Atoi(maxpages)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("max pages:", max)
	fmt.Println("inventory:", totalInventory)
	arr := strings.Split(totalInventory, " ")
	totalInventoryTrimmed := strings.TrimSpace(string(arr[0]))
	fmt.Println(totalInventoryTrimmed)

}

func TestExample(t *testing.T) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(GrabUserAgent()),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var nodes []*cdp.Node

	// <div class="ant-card css-1gl0vip-emotion--Result--cardCss ant-card-bordered ant-card-hoverable">
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(URL),
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
}

func TestUA(t *testing.T) {
	for i := 0; i < 20; i++ {
		ua := GrabUserAgent()
		fmt.Println(ua)
	}
}
