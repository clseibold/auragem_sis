package crawler

// TODO: Connect robots.txt to IP Addresses instead of domains

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/krayzpipes/cronticker/cronticker"
	_ "github.com/nakagami/firebirdsql"
	// "golang.org/x/text/encoding/ianaindex"
	// "golang.org/x/text/transform"
)

var ErrNotSupportedScheme = errors.New("not a supported protocol")
var ErrNotAllowed = errors.New("not allowed by robots.txt")
var ErrSlowDown = errors.New("slowing down")
var ErrAlreadyCrawled = errors.New("already crawled")
var ISO8601Layout = "2006-01-02T15:04:05Z0700"

var wg = &sync.WaitGroup{}

var threadSleepDurationMiliSeconds = 21 // 61 // 31
var threadSleepDurationString = "21ms"  // "61ms"  //"31ms"
var timeWaitDelay, _ = time.ParseDuration("4m")

/*func main() {
	dbConn := NewConn()
	globalData := NewGlobalData(dbConn, true, true, 0) // Follows all links
	wg.Add(2)
	go RegularCrawler(globalData, wg)
	go FeedCrawler(globalData, 13, wg)

	wg.Wait()
	globalData.dbConn.Close()
}*/

func RegularCrawler(globalData *GlobalData, wg *sync.WaitGroup) {
	time.Sleep(time.Second * 5)
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	ticker, _ := cronticker.NewTicker("@monthly") // Run on first day of every month
	wg2 := &sync.WaitGroup{}
	// globalData := NewGlobalData(false, true) // Follows internal links only

	for {
		_, ok := <-ticker.C
		if !ok {
			break
		}

		globalData.Reset()
		fmt.Printf("[0-5] Starting Search Engine Crawler.\n")
		seeds := GetSeeds(globalData)
		globalData.AddUrl("scroll://scrollprotocol.us.to/", UrlToCrawlData{})
		for _, seed := range seeds {
			globalData.AddUrl(seed.Url, UrlToCrawlData{})
		}

		wg2.Add(4)
		go Crawl(globalData, 0, wg2, 60)
		go Crawl(globalData, 1, wg2, 60)
		go Crawl(globalData, 2, wg2, 60)
		go Crawl(globalData, 3, wg2, 60)
		//go Crawl(globalData, 4, wg2, 60)
		//go Crawl(globalData, 5, wg2, 60)

		wg2.Wait()
		fmt.Printf("[0-4] Search Engine Crawler Finished.\n")
		globalData.Reset()

		// Execute procedures to update FTS database
		globalData.dbConn.Exec("EXECUTE PROCEDURE FTS$MANAGEMENT.FTS$REBUILD_INDEX('FTS_DOMAIN_ID');")
		globalData.dbConn.Exec("EXECUTE PROCEDURE FTS$MANAGEMENT.FTS$REBUILD_INDEX('FTS_PAGE_ID_EN');")

		time.Sleep(time.Minute * 30)
	}
}

// Crawls every feed and its internal links
func FeedCrawler(globalData *GlobalData, hourDuration int, wg *sync.WaitGroup, finished func()) {
	time.Sleep(time.Second * 5)
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	// Sleep to offset the start of the feed crawler until 2 days into the regular crawler
	//time.Sleep(time.Duration(float32(time.Hour*13) * 3.64))

	ticker, err := cronticker.NewTicker("0 0,13 * * *") // Every day at 00:00 and 13:00.
	if err != nil {
		panic(err)
	}
	//ticker := time.NewTicker(time.Hour * time.Duration(hourDuration)) // Every 13 hours
	wg2 := &sync.WaitGroup{}

	feedData := NewSubGlobalData(globalData, false, true, 1)
	for {
		_, ok := <-ticker.C
		if !ok {
			break
		}

		feedData.Reset()
		fmt.Printf("[6] Starting Feed Crawler.\n")
		seeds := GetFeedsAsSeeds(feedData)
		fmt.Printf("Getting %d feeds to crawl.\n", len(seeds))
		for _, seed := range seeds {
			/*if page, exists := feedData.urlsCrawled.Get(seed.Url); time.Now().Sub(page.(Page).LastSuccessfulVisit) >= time.Hour*time.Duration(hourDuration) && exists {
			feedData.AddUrl(seed.Url, UrlToCrawlData{PageFrom_LinkText: seed.Title})
			feedData.urlsCrawled.Remove(seed.Url)
			} else {*/
			feedData.AddUrl(seed.Url, UrlToCrawlData{PageFrom_LinkText: seed.Title})
			//}
		}

		wg2.Add(4)
		go Crawl(feedData, 6, wg2, 60)
		go Crawl(feedData, 7, wg2, 60)
		go Crawl(feedData, 8, wg2, 60)
		go Crawl(feedData, 9, wg2, 60)

		wg2.Wait()
		fmt.Printf("[6-9] Feed Crawler Finished.\n")
		feedData.Reset()
		finished()

		// Execute procedures to update FTS database
		globalData.dbConn.Exec("EXECUTE PROCEDURE FTS$MANAGEMENT.FTS$REBUILD_INDEX('FTS_DOMAIN_ID');")
		globalData.dbConn.Exec("EXECUTE PROCEDURE FTS$MANAGEMENT.FTS$REBUILD_INDEX('FTS_PAGE_ID_EN');")

		time.Sleep(time.Minute * 5)
	}
}

// Crawls a singular page
func OnDemandPageCrawl(globalData *GlobalData, url, title string) {
	pageCrawlData := NewSubGlobalData(globalData, false, false, 0) // Do not follow any links
	pageCrawlData.Reset()
	pageCrawlData.AddUrl(url, UrlToCrawlData{PageFrom_LinkText: title})
	Crawl(pageCrawlData, 1000, nil, 1)
}

// Crawls a root page and any internal links it leads to
func OnDemandCapsuleCrawl(globalData *GlobalData, rootUrl, title string) {
	capsuleCrawlData := NewSubGlobalData(globalData, false, true, 0) // Follow all internal links
	capsuleCrawlData.Reset()
	capsuleCrawlData.AddUrl(rootUrl, UrlToCrawlData{PageFrom_LinkText: title})
	Crawl(capsuleCrawlData, 2000, nil, 1)
}
