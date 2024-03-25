package scraper

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Product struct {
	name, price string
}

func TestScraper(t *testing.T) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`),
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var nodes []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://scrapingclub.com/exercise/list_infinite_scroll/`),
		chromedp.WaitVisible("body"),
		chromedp.Nodes(".post", &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		t.Error(err)
	}

	var name, price string
	for _, node := range nodes {
		err = chromedp.Run(ctx,
			chromedp.Text("h4", &name, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text("h5", &price, chromedp.ByQuery, chromedp.FromNode(node)),
		)
		if err != nil {
			fmt.Println("err: ", err)
		}
		fmt.Println(name, " ", price)

	}
}

// TODO
func TestMaxPage(t *testing.T) {

}

const (
	URL      = "https://www.rumbleon.com/buy?page=1"
	URL_TEST = "https://scrapingclub.com/exercise/list_infinite_scroll/"
	// USER_AGENT     = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
	ENDING_ELEMENT = `#Layer_1-2 > path:nth-child(3)`
)

func TestExample(t *testing.T) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(USER_AGENT),
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var nodes []*cdp.Node

	// <div class="ant-card css-1gl0vip-emotion--Result--cardCss ant-card-bordered ant-card-hoverable">
	err := chromedp.Run(ctx,
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
		err = chromedp.Run(ctx,
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
