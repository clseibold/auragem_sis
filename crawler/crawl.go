package crawler

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/barasher/go-exiftool"
	gemini "github.com/clseibold/go-gemini"
	"github.com/dhowden/tag"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-enry/go-enry/v2"
	"github.com/pemistahl/lingua-go"
)

var skip = map[string]bool{
	"gemini://gemini.bortzmeyer.org/": true,
	"gemini://techrights.org/":        true,
	"gemini://gemini.techrights.org/": true,
	"gemini://selve.xyz/":             true, // application/octet-stream weirdness
	"gemini://kvazar.duckdns.org/":    true,

	//"gemini://diesenbacher.net/": true, // robots.txt weirdness (no User-Agent groups specified)

	"gemini://localhost/":                        true,
	"gemini://192.168.4.26/":                     true,
	"gemini://fumble-around.mediocregopher.com/": true,
	"gemini://akewebdump.ddns.net/":              true, // Error on homepage
	"gemini://illegaldrugs.net/":                 true,
	"gemini://source.community/":                 true, // Error on invite link
	"gemini://singletona082.flounder.online/":    true, // Malformed strings
	"gemini://godocs.io/":                        true,
	"gemini://taz.de/":                           true,
}

var skipUrls = map[string]bool{
	"gemini://eph.smol.pub/Alegreya.fontpack":                  true,
	"gemini://tskaalgard.midnight.pub:1965/Autumn.jpg":         true,
	"gemini://gemini.conman.org/test/torture/":                 true,
	"gemini://gemini.conman.org/test/torture":                  true,
	"gemini://gemi.dev/cgi-bin/witw.cgi/play":                  true,
	"gemini://gemi.dev/cgi-bin":                                true, // TODO
	"gemini://kennedy.gemi.dev/image-search":                   true,
	"gemini://kennedy.gemi.dev/hashtags":                       true,
	"gemini://kennedy.gemi.dev/mentions":                       true,
	"gemini://hashnix.club/cgi/radio.cgi":                      true,
	"gemini://gemini.circumlunar.space/users/fgaz/calculator/": true,

	"gemini://topotun.hldns.ru/music/%D0%AE%D1%80%D0%B8%D0%B9_%D0%A8%D0%B8%D0%BC%D0%B0%D0%BD%D0%BE%D0%B2%D1%81%D0%BA%D0%B8%D0%B9-%D0%9C%D0%B0%D0%B9%D0%B4%D0%B0%D0%BD.mp3": true, // malformed string (possibly in mp3 metadata)
	"gemini://topotun.dynu.com/music/%D0%AE%D1%80%D0%B8%D0%B9_%D0%A8%D0%B8%D0%BC%D0%B0%D0%BD%D0%BE%D0%B2%D1%81%D0%BA%D0%B8%D0%B9-%D0%9C%D0%B0%D0%B9%D0%B4%D0%B0%D0%BD.mp3": true, // malformed string (possibly in mp3 metadata), seems to be a mirror or alt. url of the above link
	"gemini://asdfghasdfgh.de/media/1860-scott-au-clair-de-la-lune-05-09.ogg":                                                                                              true, // malformed string in audio metadata
	"gemini://gemini.ctrl-c.club/~singletona082/fiction/blue_shadows/nightwatch.gmi":                                                                                       true, // malformed string in headings
	"gemini://singletona082.flounder.online/fiction/blue_shadows/nightwatch.gmi":                                                                                           true, // malformed string

	"gemini://source.community/invite/":       true, // Some error with the db, possibly prompt is too long
	"spartan://gmi.noulin.net/stackoverflow/": true, // Mirror
}

func Crawl(globalData *GlobalData, crawlThread int, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	ctx := newCrawlContext(globalData)

	breakCounter := 0
	//for i := 0; i < 900000; i++ {
	for {
		//fmt.Printf("\n")
		nextUrl, crawlData := ctx.getNextUrl()                                           // Note: Removes the url from urlsToCrawl
		if nextUrl == "" && breakCounter >= ((1000/threadSleepDurationMiliSeconds)*60) { // Break only after 60 seconds (since 10ms delay with each)
			fmt.Printf("Nothing next. Breaking.\n")
			break
		} else if nextUrl == "" {
			breakCounter++
			sleepDuration, _ := time.ParseDuration(threadSleepDurationString)
			time.Sleep(sleepDuration)
			continue
		} else {
			breakCounter = 0
		}

		hostname, hostnameErr := GetHostname(nextUrl)
		if hostnameErr != nil || skip[hostname] || skipUrls[nextUrl] || strings.HasPrefix(nextUrl, "gemini://kennedy.gemi.dev/hashtags/") || strings.HasPrefix(nextUrl, "gemini://kennedy.gemi.dev/hashtags") || strings.HasPrefix(nextUrl, "gemini://kennedy.gemi.dev/mentions/") || strings.HasPrefix(nextUrl, "gemini://kennedy.gemi.dev/mentions") || strings.HasPrefix(nextUrl, "gemini://gemi.dev/cgi-bin/witw.cgi/play") || strings.HasPrefix(nextUrl, "gemini://gemini.thegonz.net/gemsokoban") {
			//sleepDuration, _ := time.ParseDuration(threadSleepDurationString)
			//time.Sleep(sleepDuration)
			continue
		}
		// Go through each skip url to check them as a prefix
		for prefix := range skipUrls {
			if strings.HasPrefix(nextUrl, prefix) {
				continue
			}
		}

		//fmt.Printf("Current[%d]: %s\n", crawlThread, nextUrl)
		fmt.Printf("[%d] %d out of %d left to crawl\n", crawlThread, globalData.urlsToCrawl.Count(), globalData.urlsCrawled.Count()+globalData.urlsToCrawl.Count())

		resp, err := ctx.Get(nextUrl, crawlThread, crawlData)
		if err != nil && strings.HasSuffix(err.Error(), "bind: An operation on a socket could not be performed because the system lacked sufficient buffer space or because a queue was full.") {
			//logError("Waiting for a socket's TIME_WAIT to end")
			time.Sleep(timeWaitDelay)
			// Add url back to crawl list and remove from urlsCrawled
			ctx.globalData.urlsCrawled.Remove(nextUrl)
			ctx.addUrl(nextUrl, crawlData)
			continue
		} else if err != nil || (resp == Response{}) || resp.Body == nil {
			if !errors.Is(err, ErrSlowDown) && !strings.HasSuffix(err.Error(), "not allowed by robots.txt; not allowed by robots.txt") && !strings.HasSuffix(err.Error(), "connectex: No connection could be made because the target machine actively refused it.") {
				logError("Gemini Get Error for '%s': %s; %v", nextUrl, err.Error(), err)
				sleepDuration, _ := time.ParseDuration(threadSleepDurationString)
				time.Sleep(sleepDuration)
				continue
			}
			sleepDuration, _ := time.ParseDuration(threadSleepDurationString)
			time.Sleep(sleepDuration)
			// Add url back to crawl list and remove from urlsCrawled
			ctx.globalData.urlsCrawled.Remove(nextUrl)
			ctx.addUrl(nextUrl, crawlData)
			continue
		}

		//defer cancel()
		var status int = resp.Status
		var meta string = resp.Description

		if meta == "" {
			domainIncrementEmptyMeta(ctx, ctx.GetDomain())
		}

		//fmt.Printf("Status: %d\n", status)
		//defer resp.Body.Close()
		switch status {
		case gemini.StatusInput:
			handleInput(ctx, crawlData)
		case gemini.StatusSensitiveInput:
		case gemini.StatusSuccess, 21, 22, 23, 24, 25, 26, 27, 28, 29:
			handleSuccess(ctx, crawlThread, crawlData)
		case gemini.StatusRedirect:
			handleRedirect(ctx, false, crawlData)
		//case gemini.StatusPermanentRedirect:
		case gemini.StatusRedirectPermanent:
			handleRedirect(ctx, true, crawlData)
			// TODO: Add this to a permanent redirect list, and remove the original url from the index if it exists
		case gemini.StatusTemporaryFailure:
		case gemini.StatusUnavailable: // StatusServerUnavailable
		case gemini.StatusCGIError: // TODO
		case gemini.StatusProxyError: // TODO
		case gemini.StatusSlowDown:
			handleSlowDown(ctx, crawlThread, nextUrl, crawlData)
		case gemini.StatusPermanentFailure:
			handleFailure(ctx)
		case gemini.StatusNotFound:
			handleFailure(ctx)
		case gemini.StatusGone:
			handleFailure(ctx)
		case gemini.StatusProxyRequestRefused:
		case gemini.StatusBadRequest:
			handleFailure(ctx)
		case gemini.StatusClientCertificateRequired: //StatusCertificateRequired:
		case gemini.StatusCertificateNotAuthorised: //StatusCertificateNotAuthorized:
		case gemini.StatusCertificateNotValid:
		}

		resp.Body.Close()

		sleepDuration, _ := time.ParseDuration(threadSleepDurationString)
		time.Sleep(sleepDuration)
	}

	//ctx.flush()

	//fmt.Printf("\n%v", ctx.urlsToCrawl)
	fmt.Printf("Thread %d exited.\n", crawlThread)
}

func handleInput(ctx CrawlContext, crawlData UrlToCrawlData) {
	// Add Domain
	domain := ctx.GetDomain()
	var success bool = false
	domain, success = addDomainToDb(ctx, domain, false)
	if !success {
		return // TODO
	}

	urlString := ctx.GetCurrentURL()

	prompt := ctx.resp.Description
	linecount := strings.Count(prompt, "\n")
	hasher := sha256.New()
	hasher.Write([]byte(prompt))
	hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	UDCClass := "4" // Unclassed
	page := Page{0, urlString, ctx.currentURL.Scheme, domain.Id, "text/plain", "UTF-8", "", linecount, UDCClass, "", prompt, "", len(prompt), hashStr, false, time.Time{}, time.Now().UTC(), "", "", "", "", 0, 0, "", CrawlIndex, time.Now().UTC(), time.Now().UTC(), false, false}
	success = false
	page, success = addPageToDb(ctx, page)
	if !success {
		return
	}
	ctx.setUrlCrawledPageData(urlString, page)

	// If this page was linked to from another page, add the link to the db here
	if crawlData.PageFromId != 0 {
		link, link_success := addLinkToDb(ctx, Link{0, crawlData.PageFromId, page.Id, crawlData.PageFrom_LinkText, !crawlData.PageFrom_InternalLink, CrawlIndex, time.Now().UTC()})
		if !link_success {
			// TODO: Log error and Ignore for now
			logError("Couldn't Add Link to Db: %v; Page: %v", link, page)
		}
	}
}

// Hides URL from DB: used for Permanent Failure, Not Found, Gone, and Bad Request statuses.
// TODO: Add a retry mechanism (with a max number of retries)
func handleFailure(ctx CrawlContext) {
	setPageToHidden(ctx, ctx.GetCurrentURL())
}

func handleRedirect(ctx CrawlContext, permanent bool, crawlData UrlToCrawlData) {
	meta := ctx.resp.Description
	url, _ := ctx.currentURL.Parse(meta)
	if _, ok := ctx.globalData.urlsCrawled.Get(url.String()); /*ctx.urlsCrawled[url.String()];*/ ok {
		return
	}

	/*if permanent {
		// TODO: Add to a permanent redirects table.
		// TODO: Fetch this table into a list of urls before the crawler starts crawling, that way all links can be checked and changed before having to fetch the url that will redirect
	}*/

	ctx.addUrl(url.String(), crawlData) // NOTE: The crawlData passes over into the redirect url. Do I want this?
}

func handleSlowDown(ctx CrawlContext, crawlThread int, url string, crawlData UrlToCrawlData) {
	hostname := ctx.GetCurrentHostname()

	// Increment SlowDownCount in Db
	domain := ctx.GetDomain()
	domainIncrementSlowDownCount(ctx, domain)

	//meta := ctx.resp.Meta
	// Parse meta into int and add to SlowDown // No longer parse META field as int.
	/*i, err := strconv.Atoi(meta)
	if err != nil {
	}*/

	r, exists := ctx.globalData.domainsCrawled.Get(hostname)
	if exists {
		domainInfo := r.(DomainInfo)
		domainInfo.slowDown = defaultSlowDown * 2
		ctx.globalData.domainsCrawled.Set(hostname, domainInfo)
		fmt.Printf("[%d] Slow Down: %v (%.0fs); %v", crawlThread, hostname, defaultSlowDown*2, domainInfo)
		//logError("Slow Down: %v (%ds)", hostname, i)
	} else {
		domainInfo := DomainInfo{defaultSlowDown, time.Now().UTC()}
		ctx.globalData.domainsCrawled.Set(hostname, domainInfo)
		fmt.Printf("[%d] Slow Down: %v (%.0fs); %v", crawlThread, hostname, defaultSlowDown, domainInfo)
		//logError("Slow Down: %v (%ds)", hostname, i)
	}

	// Give time for other threads to get other links before adding this one back onto the map
	time.Sleep(time.Second * 1)

	// Add url back to crawl list and remove from urlsCrawled
	ctx.globalData.urlsCrawled.Remove(url)
	ctx.addUrl(url, crawlData)
}

var langDetector = lingua.NewLanguageDetectorBuilder().FromAllLanguages().Build()

func handleSuccess(ctx CrawlContext, crawlThread int, crawlData UrlToCrawlData) {
	meta := ctx.resp.Description
	/*if meta == "" && ctx.currentURL.Scheme == "gemini" {
		meta = "text/gemini; charset=utf-8"
	} else if meta == "" && ctx.currentURL.Scheme == "scroll" {
		meta = "text/scroll; charset=utf-8"
	} else if meta == "" && ctx.currentURL.Scheme == "spartan" {
		meta = "text/gemini; charset=utf-8"
	}*/

	// Check if Body is nil for some reason. TODO: Sometimes I get a crash when I call io.ReadAll. This should solve it temporarily?
	if ctx.resp.Body == nil {
		//ctx.globalData.urlsCrawled.Remove(ctx.GetCurrentURL())
		//ctx.addUrl(ctx.GetCurrentURL(), crawlData)
		return
	}

	mediatype := ""
	var charset string = ""
	var language string = ""
	data, err := io.ReadAll(io.LimitReader(ctx.resp.Body, 1024*1024*200)) // 200 MiB Max
	if err != nil {
		// Add url back to crawl list and remove from urlsCrawled
		//ctx.globalData.urlsCrawled.Remove(ctx.GetCurrentURL())
		//ctx.addUrl(ctx.GetCurrentURL(), crawlData)
		return
	}
	if meta != "" && !strings.HasPrefix(meta, "application/octet-stream") && !strings.HasPrefix(meta, "octet-stream") {
		var params map[string]string
		mediatype, params, _ = mime.ParseMediaType(meta)
		if _, ok := params["charset"]; ok {
			charset = params["charset"]
		}
		if _, ok := params["lang"]; ok {
			language = params["lang"]
		}
	} else if ctx.isRootPage || strings.HasSuffix(ctx.currentURL.Path, "/") {
		if strings.HasPrefix(ctx.GetCurrentURL(), "gemini://") {
			mediatype = "text/gemini"
		} else if strings.HasPrefix(ctx.GetCurrentURL(), "scroll://") {
			mediatype = "text/scroll"
		} else if strings.HasPrefix(ctx.GetCurrentURL(), "spartan://") {
			mediatype = "text/gemini"
		} else if strings.HasPrefix(ctx.GetCurrentURL(), "nex://") {
			mediatype = "text/nex"
		}
	} else {
		if strings.HasSuffix(ctx.currentURL.Path, ".gmi") || strings.HasSuffix(ctx.currentURL.Path, ".gemini") {
			mediatype = "text/gemini"
		} else if strings.HasSuffix(ctx.currentURL.Path, ".scroll") || strings.HasSuffix(ctx.currentURL.Path, ".abstract") {
			mediatype = "text/scroll"
		} else {
			mediatype = mimetype.Detect(data).String()
		}
	}

	// Try to detect the language on textual mimetypes
	if MediatypeIsTextual(mediatype) {
		lang, reliable := langDetector.DetectLanguageOf(string(data))
		if reliable || language == "" {
			language = lang.IsoCode639_1().String()
		}
	}

	UDCClass := "4" // Unclassed
	if strings.HasPrefix(ctx.GetCurrentURL(), "scroll://") {
		UDCClass = strconv.Itoa(ctx.resp.Status - 20)
	}

	// Get/add domain (but don't provide all details unless the current page is the root of the domain)
	domain := ctx.GetDomain()
	if !ctx.isRootPage {
		var success bool = false
		domain, success = addDomainToDb(ctx, domain, false)
		if !success {
			return // TODO
		}
		//fmt.Printf("Not Root, Domain: %s, %d\n", domain.Domain, domain.Id)
	}

	if mediatype == "text/gemini" || mediatype == "text/spartan" || mediatype == "text/scroll" {
		var strippedTextBuilder strings.Builder
		tagsMap := make(map[string]float64)
		mentionsMap := make(map[string]bool)
		links := make([]GeminiLink, 0)
		update := true
		// TODO: Some articles have "📅" on a line prefixed before a publication date
		geminiTitle, linecount, headings, _, size, isFeed := ctx.GetGeminiPageInfo2(bytes.NewReader(data), &tagsMap, &mentionsMap, &links, &strippedTextBuilder, update)
		// Exclude tag pages from being considered feeds
		if strings.Contains(ctx.GetCurrentURL(), "/tag/") || strings.Contains(ctx.GetCurrentURL(), "/tags/") {
			isFeed = false
		}
		// Manually handle title for gemini://station.martinrue.com
		if ctx.GetCurrentURL() == "gemini://station.martinrue.com/" {
			geminiTitle = "Station"
		} else if ctx.GetCurrentURL() == "gemini://hashnix.club/" {
			geminiTitle = "Hashnix Club"
		}

		if geminiTitle == "" || !ContainsLetterRunes(geminiTitle) {
			if crawlData.PageFrom_InternalLink {
				geminiTitle = crawlData.PageFrom_LinkText
			}
		}

		// Publication Date Handling: Get from internal link, overwrite from title or filename if available // TODO: Check for dates in the path directories just above the file
		timeCutoff := time.Now().Add(time.Hour * 24).UTC()
		publicationDate := ctx.resp.PublishDate
		if crawlData.PageFrom_InternalLink {
			date := getTimeDate(crawlData.PageFrom_LinkText, false)
			if (date != time.Time{} && !date.After(timeCutoff)) {
				publicationDate = date
			}
		}
		if !strings.Contains(ctx.GetCurrentURL(), "~Cosmos/thread") {
			date := getTimeDate(geminiTitle, false)
			if (date != time.Time{} && !date.After(timeCutoff)) {
				publicationDate = date
			}
		}
		// TODO: Hacky - don't get publishdate from filename if "commit" or "~Cosmos/thread" is in the URL, so that there's no false positives with hashes, and Cosmos threads aren't included
		if !strings.Contains(ctx.GetCurrentURL(), "commit/") && !strings.Contains(ctx.GetCurrentURL(), "commits/") && !strings.Contains(ctx.GetCurrentURL(), "~Cosmos/thread") {
			_, filename := path.Split(ctx.GetCurrentURL())
			publicationDate2 := getTimeDate(filename, true)
			if (publicationDate2 != time.Time{} && !publicationDate2.After(timeCutoff)) {
				publicationDate = publicationDate2
			}
		}

		// If publication date is in the future, then reset publicationDate to time.Time{}
		if publicationDate.After(timeCutoff) {
			publicationDate = time.Time{}
		}

		textStr := string(data)
		hasher := sha256.New()
		hasher.Write([]byte(textStr))
		hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		// If root page of domain, update the db domain information to include title
		if ctx.isRootPage {
			//fmt.Printf("Getting Domain for Root %s\n", ctx.currentURL)
			domain.Title = geminiTitle
			var success bool = false
			domain, success = addDomainToDb(ctx, domain, true)
			if !success {
				return // TODO
			}
			//fmt.Printf("DomainId of %s: %d\n", ctx.currentURL, domain.Id)
		}

		// Update the entry in the db if needed.
		//if update {
		urlString := ctx.GetCurrentURL()
		scheme := strings.ToLower(strings.TrimSuffix(ctx.currentURL.Scheme, "://"))
		hidden := false

		// If there's non-hidden duplicates from same scheme, hide this page
		if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, scheme)) > 0 {
			hidden = true
		}

		hasDuplicateOnGemini := false
		if scheme == "gemini" {
			// If there's pages on other protocols with the hash, and current scheme is gemini, then set all of those others as having gemini duplicate.
			if len(getPagesWithHashAndNotScheme(ctx, urlString, hashStr, scheme)) > 0 {
				setPageHashHasGeminiDuplicate(ctx, urlString, hashStr, true)
			}
		} else {
			// If there's a gemini page with the hash that is not hidden, then set hasDuplicateOnGemini
			if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, "gemini")) > 0 {
				hasDuplicateOnGemini = true
			}
		}

		page := Page{0, urlString, ctx.currentURL.Scheme, domain.Id, mediatype, charset, language, linecount, UDCClass, geminiTitle, "", headings, size, hashStr, isFeed, publicationDate, time.Now().UTC(), "", "", "", "", 0, 0, "", CrawlIndex, time.Now().UTC(), time.Now().UTC(), hidden, hasDuplicateOnGemini}
		var success bool = false
		page, success = addPageToDb(ctx, page)
		if !success {
			return
		}
		ctx.setUrlCrawledPageData(urlString, page)

		// If this page was linked to from another page, add the link to the db here
		if crawlData.PageFromId != 0 {
			link, link_success := addLinkToDb(ctx, Link{0, crawlData.PageFromId, page.Id, crawlData.PageFrom_LinkText, !crawlData.PageFrom_InternalLink, CrawlIndex, time.Now().UTC()})
			if !link_success {
				// TODO: Log error and Ignore for now
				logError("Couldn't Add Link to Db: %v; Page: %v", link, page)
			}
		}

		/*
			for tag, rank := range tagsMap {
				graphemeCount := uniseg.GraphemeClusterCount(tag)
				if len(tag) <= 2 || graphemeCount > 250 {
					continue
				}
				addTagToDb(ctx, page.Id, tag, rank)
			}

			for mention := range mentionsMap {
				graphemeCount := uniseg.GraphemeClusterCount(mention)
				if graphemeCount > 250 {
					continue
				}
				addMentionToDb(ctx, page.Id, mention)
			}
		*/
		//}

		for _, link := range links {
			if link.spartanInput {
				// Skip spartan input links for now
				continue
			}
			url, _ := ctx.currentURL.Parse(link.url) // NOTE: This call will translate all relative and absolute links in the context of the current page's URL.
			if url == nil {
				continue
			}
			url.Fragment = "" // Strip the fragment
			internalLink := ctx.currentURL.Hostname() == url.Hostname() && ctx.currentURL.Port() == url.Port() && ctx.currentURL.Scheme == url.Scheme
			if crawledPage, ok := ctx.globalData.urlsCrawled.Get(url.String()); /*ctx.urlsCrawled[url.String()]*/ ok {
				// Link is already crawled. TODO: What if the crawledPage's info hasn't been set yet?
				if crawledPage.(Page).Id != 0 {
					dbLink, db_success := addLinkToDb(ctx, Link{0, page.Id, crawledPage.(Page).Id, link.name, !internalLink, CrawlIndex, time.Now().UTC()})
					if !db_success {
						logError("Couldn't Add Link to Db: %v; From Page: %v", dbLink, page)
					}
				}
				continue
			}
			if internalLink && ctx.globalData.followInternalLinks {
				allow := ctx.currentRobots.indexerGroup.Test(url.Path)
				// If not in robots.txt, or if depth is greater than max depth, then skip link
				if !allow || (ctx.globalData.maxDepth != 0 && crawlData.currentDepth+1 > ctx.globalData.maxDepth) {
					continue
				}
				ctx.addUrl(url.String(), UrlToCrawlData{page.Id, true, link.name, crawlData.currentDepth + 1})
			} else if (url.Scheme == "gemini" || url.Scheme == "nex" || url.Scheme == "scroll" || url.Scheme == "spartan") && ctx.globalData.followExternalLinks {
				ctx.addUrl(url.String(), UrlToCrawlData{page.Id, false, link.name, 0})
			}
		}
		// TODO: text/markdown and text/html
	} else if mediatype == "text/nex" { // Nex Listing file
		var strippedTextBuilder strings.Builder
		links := make([]NexLink, 0)
		title, linecount, headings, _, size, isFeed := ctx.GetNexPageInfo(bytes.NewReader(data), nil, nil, &links, &strippedTextBuilder, true)
		// Exclude tag pages from being considered feeds
		if strings.Contains(ctx.GetCurrentURL(), "/tag/") || strings.Contains(ctx.GetCurrentURL(), "/tags/") {
			isFeed = false
		} else if ctx.GetCurrentURL() == "nex://station.martinrue.com/" {
			title = "Station"
		} else if ctx.GetCurrentURL() == "nex://hashnix.club/" {
			title = "Hashnix Club"
		}

		if title == "" || !ContainsLetterRunes(title) {
			if crawlData.PageFrom_InternalLink {
				title = crawlData.PageFrom_LinkText
			}
		}

		// Publication Date Handling: Get from title or filename // TODO: Check for dates in the path directories just above the file
		publicationDate := time.Time{}
		if !strings.Contains(ctx.GetCurrentURL(), "~Cosmos/thread") {
			publicationDate = getTimeDate(title, false)
		}
		// TODO: Hacky - don't get publishdate from filename if "commit" or "~Cosmos/thread" is in the URL, so that there's no false positives with hashes, and Cosmos threads aren't included
		if !strings.Contains(ctx.GetCurrentURL(), "commit/") && !strings.Contains(ctx.GetCurrentURL(), "commits/") && !strings.Contains(ctx.GetCurrentURL(), "~Cosmos/thread") {
			_, filename := path.Split(ctx.GetCurrentURL())
			publicationDate2 := getTimeDate(filename, true)
			if (publicationDate2 != time.Time{}) {
				publicationDate = publicationDate2
			}
		}

		// If publication date is in the future, then reset publicationDate to time.Time{}
		if publicationDate.After(time.Now().Add(time.Hour * 24).UTC()) {
			publicationDate = time.Time{}
		}

		hasher := sha256.New()
		hasher.Write(data)
		hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		// If root page of domain, add it if it's not a thing yet
		if ctx.isRootPage {
			//fmt.Printf("Getting Domain for Root %s\n", ctx.currentURL)
			//domain.Title = title
			var success bool = false
			domain, success = addDomainToDb(ctx, domain, false)
			if !success {
				return // TODO
			}
			//fmt.Printf("DomainId of %s: %d\n", ctx.currentURL, domain.Id)
		}

		urlString := ctx.GetCurrentURL()
		scheme := strings.ToLower(strings.TrimSuffix(ctx.currentURL.Scheme, "://"))
		hidden := false

		// If there's non-hidden duplicates from same scheme, hide this page
		if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, scheme)) > 0 {
			hidden = true
		}

		hasDuplicateOnGemini := false
		if scheme == "gemini" {
			// If there's pages on other protocols with the hash, and current scheme is gemini, then set all of those others as having gemini duplicate.
			if len(getPagesWithHashAndNotScheme(ctx, urlString, hashStr, scheme)) > 0 {
				setPageHashHasGeminiDuplicate(ctx, urlString, hashStr, true)
			}
		} else {
			// If there's a gemini page with the hash that is not hidden, then set hasDuplicateOnGemini
			if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, "gemini")) > 0 {
				hasDuplicateOnGemini = true
			}
		}

		page := Page{0, urlString, ctx.currentURL.Scheme, domain.Id, mediatype, charset, language, linecount, UDCClass, title, "", headings, size, hashStr, isFeed, publicationDate, time.Now().UTC(), "", "", "", "", 0, 0, "", CrawlIndex, time.Now().UTC(), time.Now().UTC(), hidden, hasDuplicateOnGemini}
		var success bool = false
		page, success = addPageToDb(ctx, page)
		if !success {
			return
		}
		ctx.setUrlCrawledPageData(urlString, page)

		// If this page was linked to from another page, add the link to the db here
		if crawlData.PageFromId != 0 {
			link, link_success := addLinkToDb(ctx, Link{0, crawlData.PageFromId, page.Id, crawlData.PageFrom_LinkText, !crawlData.PageFrom_InternalLink, CrawlIndex, time.Now().UTC()})
			if !link_success {
				// TODO: Log error and Ignore for now
				logError("Couldn't Add Link to Db: %v; Page: %v", link, page)
			}
		}

		for _, link := range links {
			url, _ := ctx.currentURL.Parse(link.url) // NOTE: This call will translate all relative and absolute links in the context of the current page's URL.
			if url == nil {
				continue
			}
			url.Fragment = "" // Strip the fragment
			internalLink := ctx.currentURL.Hostname() == url.Hostname() && ctx.currentURL.Port() == url.Port() && ctx.currentURL.Scheme == url.Scheme
			if crawledPage, ok := ctx.globalData.urlsCrawled.Get(url.String()); /*ctx.urlsCrawled[url.String()]*/ ok {
				// Link is already crawled. TODO: What if the crawledPage's info hasn't been set yet?
				if crawledPage.(Page).Id != 0 {
					dbLink, db_success := addLinkToDb(ctx, Link{0, page.Id, crawledPage.(Page).Id, link.name, !internalLink, CrawlIndex, time.Now().UTC()})
					if !db_success {
						logError("Couldn't Add Link to Db: %v; From Page: %v", dbLink, page)
					}
				}
				continue
			}
			if internalLink && ctx.globalData.followInternalLinks {
				allow := ctx.currentRobots.indexerGroup.Test(url.Path)
				// If not in robots.txt, or if depth is greater than max depth, then skip link
				if !allow || (ctx.globalData.maxDepth != 0 && crawlData.currentDepth+1 > ctx.globalData.maxDepth) {
					continue
				}
				ctx.addUrl(url.String(), UrlToCrawlData{page.Id, true, link.name, crawlData.currentDepth + 1})
			} else if (url.Scheme == "gemini" || url.Scheme == "nex" || url.Scheme == "scroll" || url.Scheme == "spartan") && ctx.globalData.followExternalLinks {
				ctx.addUrl(url.String(), UrlToCrawlData{page.Id, false, link.name, 0})
			}
		}
	} else if strings.HasPrefix(mediatype, "text/markdown") {
		var strippedTextBuilder strings.Builder
		links := make([]MarkdownLink, 0)
		title, linecount, headings, _, size, isFeed := ctx.GetMarkdownPageInfo(bytes.NewReader(data), nil, nil, &links, &strippedTextBuilder, true)
		// Exclude tag pages from being considered feeds
		if strings.Contains(ctx.GetCurrentURL(), "/tag/") || strings.Contains(ctx.GetCurrentURL(), "/tags/") {
			isFeed = false
		}
		if title == "" || !ContainsLetterRunes(title) {
			if crawlData.PageFrom_InternalLink {
				title = crawlData.PageFrom_LinkText
			}
		}

		// Publication Date Handling: Get from title or filename // TODO: Check for dates in the path directories just above the file
		publicationDate := time.Time{}
		if !strings.Contains(ctx.GetCurrentURL(), "~Cosmos/thread") {
			publicationDate = getTimeDate(title, false)
		}
		// TODO: Hacky - don't get publishdate from filename if "commit" or "~Cosmos/thread" is in the URL, so that there's no false positives with hashes, and Cosmos threads aren't included
		if !strings.Contains(ctx.GetCurrentURL(), "commit/") && !strings.Contains(ctx.GetCurrentURL(), "commits/") && !strings.Contains(ctx.GetCurrentURL(), "~Cosmos/thread") {
			_, filename := path.Split(ctx.GetCurrentURL())
			publicationDate2 := getTimeDate(filename, true)
			if (publicationDate2 != time.Time{}) {
				publicationDate = publicationDate2
			}
		}

		// If publication date is in the future, then reset publicationDate to time.Time{}
		if publicationDate.After(time.Now().Add(time.Hour * 24).UTC()) {
			publicationDate = time.Time{}
		}

		hasher := sha256.New()
		hasher.Write(data)
		hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		// If root page of domain, add it if it's not a thing yet
		if ctx.isRootPage {
			//fmt.Printf("Getting Domain for Root %s\n", ctx.currentURL)
			domain.Title = title
			var success bool = false
			domain, success = addDomainToDb(ctx, domain, true)
			if !success {
				return // TODO
			}
			//fmt.Printf("DomainId of %s: %d\n", ctx.currentURL, domain.Id)
		}

		urlString := ctx.GetCurrentURL()
		scheme := strings.ToLower(strings.TrimSuffix(ctx.currentURL.Scheme, "://"))
		hidden := false

		// If there's non-hidden duplicates from same scheme, hide this page
		if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, scheme)) > 0 {
			hidden = true
		}

		hasDuplicateOnGemini := false
		if scheme == "gemini" {
			// If there's pages on other protocols with the hash, and current scheme is gemini, then set all of those others as having gemini duplicate.
			if len(getPagesWithHashAndNotScheme(ctx, urlString, hashStr, scheme)) > 0 {
				setPageHashHasGeminiDuplicate(ctx, urlString, hashStr, true)
			}
		} else {
			// If there's a gemini page with the hash that is not hidden, then set hasDuplicateOnGemini
			if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, "gemini")) > 0 {
				hasDuplicateOnGemini = true
			}
		}

		page := Page{0, urlString, ctx.currentURL.Scheme, domain.Id, mediatype, charset, language, linecount, UDCClass, title, "", headings, size, hashStr, isFeed, publicationDate, time.Now().UTC(), "", "", "", "", 0, 0, "", CrawlIndex, time.Now().UTC(), time.Now().UTC(), hidden, hasDuplicateOnGemini}
		var success bool = false
		page, success = addPageToDb(ctx, page)
		if !success {
			return
		}
		ctx.setUrlCrawledPageData(urlString, page)

		// If this page was linked to from another page, add the link to the db here
		if crawlData.PageFromId != 0 {
			link, link_success := addLinkToDb(ctx, Link{0, crawlData.PageFromId, page.Id, crawlData.PageFrom_LinkText, !crawlData.PageFrom_InternalLink, CrawlIndex, time.Now().UTC()})
			if !link_success {
				// TODO: Log error and Ignore for now
				logError("Couldn't Add Link to Db: %v; Page: %v", link, page)
			}
		}

		for _, link := range links {
			url, _ := ctx.currentURL.Parse(link.url) // NOTE: This call will translate all relative and absolute links in the context of the current page's URL.
			if url == nil {
				continue
			}
			url.Fragment = "" // Strip the fragment
			internalLink := ctx.currentURL.Hostname() == url.Hostname() && ctx.currentURL.Port() == url.Port() && ctx.currentURL.Scheme == url.Scheme
			if crawledPage, ok := ctx.globalData.urlsCrawled.Get(url.String()); /*ctx.urlsCrawled[url.String()]*/ ok {
				// Link is already crawled. TODO: What if the crawledPage's info hasn't been set yet?
				if crawledPage.(Page).Id != 0 {
					dbLink, db_success := addLinkToDb(ctx, Link{0, page.Id, crawledPage.(Page).Id, link.name, !internalLink, CrawlIndex, time.Now().UTC()})
					if !db_success {
						logError("Couldn't Add Link to Db: %v; From Page: %v", dbLink, page)
					}
				}
				continue
			}
			if internalLink && ctx.globalData.followInternalLinks {
				allow := ctx.currentRobots.indexerGroup.Test(url.Path)
				// If not in robots.txt, or if depth is greater than max depth, then skip link
				if !allow || (ctx.globalData.maxDepth != 0 && crawlData.currentDepth+1 > ctx.globalData.maxDepth) {
					continue
				}
				ctx.addUrl(url.String(), UrlToCrawlData{page.Id, true, link.name, crawlData.currentDepth + 1})
			} else if (url.Scheme == "gemini" || url.Scheme == "nex" || url.Scheme == "scroll" || url.Scheme == "spartan") && ctx.globalData.followExternalLinks {
				ctx.addUrl(url.String(), UrlToCrawlData{page.Id, false, link.name, 0})
			}
		}
	} else if strings.HasPrefix(mediatype, "text/") {
		textBytes := data
		textStr := string(textBytes)
		size := len(textBytes)
		//keywords := rake.RunRake(textStr)

		// Detect programming language of file, if there is one
		//preOrCodeText := ""
		language := ""
		language = enry.GetLanguage(path.Base(ctx.currentURL.Path), textBytes) // NOTE: .txt plain/text files return "Text" as lang
		if language == "" {
			// TODO: empty string is returned when file is binary or when language is unknown
		}
		extension := path.Ext(ctx.currentURL.Path)
		switch extension {
		case ".ha":
			//preOrCodeText = textStr
			language = "Hare"
		}

		// Get number of lines
		linecount := strings.Count(textStr, "\n")

		hasher := sha256.New()
		hasher.Write([]byte(textStr))
		hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		// If root page of domain, add the domain still
		if ctx.isRootPage {
			//fmt.Printf("Getting Domain for Root %s\n", ctx.currentURL)
			var success bool = false
			domain, success = addDomainToDb(ctx, domain, false)
			if !success {
				return // TODO
			}
			//fmt.Printf("DomainId of %s: %d\n", ctx.currentURL, domain.Id)
		}

		urlString := ctx.GetCurrentURL()
		scheme := strings.ToLower(strings.TrimSuffix(ctx.currentURL.Scheme, "://"))
		hidden := false

		// If there's non-hidden duplicates from same scheme, hide this page
		if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, scheme)) > 0 {
			hidden = true
		}

		hasDuplicateOnGemini := false
		if scheme == "gemini" {
			// If there's pages on other protocols with the hash, and current scheme is gemini, then set all of those others as having gemini duplicate.
			if len(getPagesWithHashAndNotScheme(ctx, urlString, hashStr, scheme)) > 0 {
				setPageHashHasGeminiDuplicate(ctx, urlString, hashStr, true)
			}
		} else {
			// If there's a gemini page with the hash that is not hidden, then set hasDuplicateOnGemini
			if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, "gemini")) > 0 {
				hasDuplicateOnGemini = true
			}
		}

		var title string
		if crawlData.PageFrom_InternalLink {
			title = crawlData.PageFrom_LinkText
		}
		page := Page{0, urlString, ctx.currentURL.Scheme, domain.Id, mediatype, charset, language, linecount, UDCClass, title, "", "", size, hashStr, false, time.Time{}, time.Now().UTC(), "", "", "", "", 0, 0, "", CrawlIndex, time.Now().UTC(), time.Now().UTC(), hidden, hasDuplicateOnGemini}
		var success bool = false
		page, success = addPageToDb(ctx, page)
		if !success {
			return
		}
		ctx.setUrlCrawledPageData(urlString, page)

		// If this page was linked to from another page, add the link to the db here
		if crawlData.PageFromId != 0 {
			link, link_success := addLinkToDb(ctx, Link{0, crawlData.PageFromId, page.Id, crawlData.PageFrom_LinkText, !crawlData.PageFrom_InternalLink, CrawlIndex, time.Now().UTC()})
			if !link_success {
				// TODO: Log error and Ignore for now
				logError("Couldn't Add Link to Db: %v; Page: %v", link, page)
			}
		}

		//} else if mediatype == "text/markdown" {
		/*textBytes, _ := io.ReadAll(ctx.resp.Body)
		textStr := string(textBytes)
		size := len(textBytes)

		hasher := sha256.New()
		hasher.Write([]byte(textStr))
		hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		urlString := ctx.GetCurrentURL()
		*/
	} else if mediatype == "audio/mpeg" || mediatype == "audio/mp3" || mediatype == "audio/ogg" || mediatype == "audio/flac" || mediatype == "audio/x-flac" {
		p := data
		size := len(data)
		m, _ := tag.ReadFrom(bytes.NewReader(p[:size]))
		if m == nil {
			return
		}

		hasher := sha256.New()
		hasher.Write(p[:size])
		hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		//fmt.Printf("Title: %s; Hash: %s\n", m.Title(), hashStr)
		track, _ := m.Track()
		disc, _ := m.Disc()
		title := m.Title()

		if title == "" {
			if crawlData.PageFrom_InternalLink {
				title = crawlData.PageFrom_LinkText
			}
		}

		//tag.SumID3v2()

		urlString := ctx.GetCurrentURL()
		scheme := strings.ToLower(strings.TrimSuffix(ctx.currentURL.Scheme, "://"))
		hidden := false

		// If there's non-hidden duplicates from same scheme, hide this page
		if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, scheme)) > 0 {
			hidden = true
		}

		hasDuplicateOnGemini := false
		if scheme == "gemini" {
			// If there's pages on other protocols with the hash, and current scheme is gemini, then set all of those others as having gemini duplicate.
			if len(getPagesWithHashAndNotScheme(ctx, urlString, hashStr, scheme)) > 0 {
				setPageHashHasGeminiDuplicate(ctx, urlString, hashStr, true)
			}
		} else {
			// If there's a gemini page with the hash that is not hidden, then set hasDuplicateOnGemini
			if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, "gemini")) > 0 {
				hasDuplicateOnGemini = true
			}
		}

		/*urlHasher := sha256.New()
		urlHasher.Write([]byte(urlString))
		urlHash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))*/

		page := Page{0, urlString, ctx.currentURL.Scheme, domain.Id, mediatype, charset, language, 0, UDCClass, title, "", "", size, hashStr, false, time.Time{}, time.Now().UTC(), m.Album(), m.Artist(), m.AlbumArtist(), m.Composer(), track, disc, "", CrawlIndex, time.Now().UTC(), time.Now().UTC(), hidden, hasDuplicateOnGemini}
		var success bool = false
		page, success = addPageToDb(ctx, page)
		if !success {
			return
		}
		ctx.setUrlCrawledPageData(urlString, page)

		// If this page was linked to from another page, add the link to the db here
		if crawlData.PageFromId != 0 {
			link, link_success := addLinkToDb(ctx, Link{0, crawlData.PageFromId, page.Id, crawlData.PageFrom_LinkText, !crawlData.PageFrom_InternalLink, CrawlIndex, time.Now().UTC()})
			if !link_success {
				// TODO: Log error and Ignore for now
				logError("Couldn't Add Link to Db: %v; Page: %v", link, page)
			}
		}
	} else if mediatype == "application/pdf" || mediatype == "image/vnd.djvu" || mediatype == "application/epub" || mediatype == "application/epub+zip" {
		et, err := exiftool.NewExiftool()
		if err != nil {
			logError("Error when intializing: %v\n", err)
			return
		}
		defer et.Close()

		p := data
		size := len(data)
		tmpFilename := fmt.Sprintf("tmp_pdf_thread_%d%s", crawlThread, path.Ext(ctx.currentURL.Path))
		err = os.WriteFile(tmpFilename, p, 0644)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			logError("Error writing file '%s' for '%s': %s; %v", tmpFilename, ctx.GetCurrentURL(), err.Error(), err)
			return
		}

		fileInfos := et.ExtractMetadata(tmpFilename)
		fileInfo := fileInfos[0]
		if fileInfo.Err != nil {
			fmt.Printf("Error with fileinfo for file %s: %v\n", fileInfo.File, fileInfo.Err)
			logError("Error getting fileinfo '%s' for '%s': %s; %v", fileInfo.File, ctx.GetCurrentURL(), fileInfo.Err.Error(), fileInfo.Err)
			return
		}
		os.Remove(tmpFilename)

		/* Author
		author, authorExists := fileInfo.Fields["Author"]
		if !authorExists {
			author, authorExists = fileInfo.Fields["author"]
			if !authorExists {
				author = ""
			}
		}
		*/

		title, titleExists := fileInfo.Fields["Title"]
		if !titleExists {
			title, titleExists = fileInfo.Fields["title"]
			if !titleExists {
				title, titleExists = fileInfo.Fields["booktitle"]
				if !titleExists {
					if crawlData.PageFrom_InternalLink {
						title = crawlData.PageFrom_LinkText
					}
				}
			}
		}

		copyright, copyrightExists := fileInfo.Fields["Copyright"]
		if !copyrightExists {
			copyright = ""
		}

		if language == "" {
			language2, languageExists := fileInfo.Fields["Lang"]
			if !languageExists {
				language2, languageExists = fileInfo.Fields["Language"]
				if !languageExists {
					language2 = ""
				}
			}
			language = language2.(string)
		}

		// TODO: Add keywords stuff here?

		hasher := sha256.New()
		hasher.Write(p[:size])
		hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		urlString := ctx.GetCurrentURL()
		scheme := strings.ToLower(strings.TrimSuffix(ctx.currentURL.Scheme, "://"))
		hidden := false

		// If there's non-hidden duplicates from same scheme, hide this page
		if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, scheme)) > 0 {
			hidden = true
		}

		hasDuplicateOnGemini := false
		if scheme == "gemini" {
			// If there's pages on other protocols with the hash, and current scheme is gemini, then set all of those others as having gemini duplicate.
			if len(getPagesWithHashAndNotScheme(ctx, urlString, hashStr, scheme)) > 0 {
				setPageHashHasGeminiDuplicate(ctx, urlString, hashStr, true)
			}
		} else {
			// If there's a gemini page with the hash that is not hidden, then set hasDuplicateOnGemini
			if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, "gemini")) > 0 {
				hasDuplicateOnGemini = true
			}
		}

		page := Page{0, urlString, ctx.currentURL.Scheme, domain.Id, mediatype, charset, language, 0, UDCClass, title.(string), "", "", size, hashStr, false, time.Time{}, time.Now().UTC(), "", "", "", "", 0, 0, copyright.(string), CrawlIndex, time.Now().UTC(), time.Now().UTC(), hidden, hasDuplicateOnGemini}
		var success bool = false
		page, success = addPageToDb(ctx, page)
		if !success {
			return
		}
		ctx.setUrlCrawledPageData(urlString, page)

		// If this page was linked to from another page, add the link to the db here
		if crawlData.PageFromId != 0 {
			link, link_success := addLinkToDb(ctx, Link{0, crawlData.PageFromId, page.Id, crawlData.PageFrom_LinkText, !crawlData.PageFrom_InternalLink, CrawlIndex, time.Now().UTC()})
			if !link_success {
				// TODO: Log error and Ignore for now
				logError("Couldn't Add Link to Db: %v; Page: %v", link, page)
			}
		}
	} else {
		if ctx.isRootPage {
			fmt.Printf("Weird %s: %s\n", ctx.currentURL, meta)
			panic("Weirdness happening!")
		}

		p := data
		size := len(data)
		hasher := sha256.New()
		hasher.Write(p[:size])
		hashStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		urlString := ctx.GetCurrentURL()
		scheme := strings.ToLower(strings.TrimSuffix(ctx.currentURL.Scheme, "://"))
		hidden := false

		// If there's non-hidden duplicates from same scheme, hide this page
		if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, scheme)) > 0 {
			hidden = true
		}

		hasDuplicateOnGemini := false
		if scheme == "gemini" {
			// If there's pages on other protocols with the hash, and current scheme is gemini, then set all of those others as having gemini duplicate.
			if len(getPagesWithHashAndNotScheme(ctx, urlString, hashStr, scheme)) > 0 {
				setPageHashHasGeminiDuplicate(ctx, urlString, hashStr, true)
			}
		} else {
			// If there's a gemini page with the hash that is not hidden, then set hasDuplicateOnGemini
			if len(getPagesWithHashAndScheme(ctx, urlString, hashStr, "gemini")) > 0 {
				hasDuplicateOnGemini = true
			}
		}

		var title string
		if crawlData.PageFrom_InternalLink {
			title = crawlData.PageFrom_LinkText
		}
		page := Page{0, urlString, ctx.currentURL.Scheme, domain.Id, mediatype, charset, language, 0, UDCClass, title, "", "", size, hashStr, false, time.Time{}, time.Now().UTC(), "", "", "", "", 0, 0, "", CrawlIndex, time.Now().UTC(), time.Now().UTC(), hidden, hasDuplicateOnGemini}
		var success bool = false
		page, success = addPageToDb(ctx, page)
		if !success {
			return
		}
		ctx.setUrlCrawledPageData(urlString, page)

		// If this page was linked to from another page, add the link to the db here
		if crawlData.PageFromId != 0 {
			link, link_success := addLinkToDb(ctx, Link{0, crawlData.PageFromId, page.Id, crawlData.PageFrom_LinkText, !crawlData.PageFrom_InternalLink, CrawlIndex, time.Now().UTC()})
			if !link_success {
				// TODO: Log error and Ignore for now
				logError("Couldn't Add Link to Db: %v; Page: %v", link, page)
			}
		}
	}
}
