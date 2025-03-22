package crawler

import (
	"crypto/x509"
	"database/sql"
	"errors"
	"io"
	"math"
	"math/rand"
	"net"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	//geminiParser "git.sr.ht/~adnano/go-gemini"
	"github.com/clseibold/go-gemini"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/temoto/robotstxt"
	"gitlab.com/clseibold/gonex/nex_client"
	spartan_client "gitlab.com/clseibold/profectus/spartan"
	scroll "gitlab.com/clseibold/scroll-term/scroll_client"
)

var CrawlIndex = 4

type Response struct {
	Status           int
	Description      string
	Author           string
	PublishDate      time.Time
	ModificationDate time.Time
	Body             io.ReadCloser
	Cert             *x509.Certificate
}

type Robots struct {
	robots       *robotstxt.RobotsData
	indexerGroup *robotstxt.Group
	//allGroup     *robotstxt.Group
}

var defaultSlowDown float64 = 2 // Changed to 2 seconds, used to be 2.5

type DomainInfo struct {
	slowDown      float64 // In Seconds
	lastCrawlTime time.Time
}

type UrlToCrawlData struct {
	PageFromId            int
	PageFrom_InternalLink bool
	PageFrom_LinkText     string
	currentDepth          int
}

// GlobalData is used by all threads
type GlobalData struct {
	domainsCrawled cmap.ConcurrentMap // DomainInfo
	urlsCrawled    cmap.ConcurrentMap // map[string]struct{}
	urlsToCrawl    cmap.ConcurrentMap // bool is whether robots.txt should be checked or not
	robotsMap      cmap.ConcurrentMap
	dbConn         *sql.DB
	crawlStartTime time.Time

	// Whether to follow links
	followExternalLinks bool
	followInternalLinks bool
	maxDepth            int // 0 to disregard depth
	sub                 bool
}

func NewGlobalData(db *sql.DB, followExternalLinks bool, followInternalLinks bool, maxDepth int) *GlobalData {
	return &GlobalData{cmap.New(), cmap.New(), cmap.New(), cmap.New(), db, time.Now(), followExternalLinks, followInternalLinks, maxDepth, false}
}

// NewSubGlobalData creates a new global data with the same domainsCrawled, urlsCrawled, and robots maps but a different urlsToCrawl List
/*func NewSubGlobalData(globalData *GlobalData, followExternalLinks bool, followInternalLinks bool, maxDepth int) *GlobalData {
	return &GlobalData{globalData.domainsCrawled, globalData.urlsCrawled, cmap.New(), globalData.robotsMap, globalData.dbConn, followExternalLinks, followInternalLinks, maxDepth}
}*/

// NewSubGlobalData creates a new global data with the same robots map and domainsCrawled, but with different urlsToCrawl and urlsCrawled Lists
func NewSubGlobalData(globalData *GlobalData, followExternalLinks bool, followInternalLinks bool, maxDepth int) *GlobalData {
	return &GlobalData{globalData.domainsCrawled, cmap.New(), cmap.New(), globalData.robotsMap, globalData.dbConn, time.Now(), followExternalLinks, followInternalLinks, maxDepth, true}
}

func (gd *GlobalData) Reset() {
	gd.urlsCrawled.Clear()
	gd.urlsToCrawl.Clear()
	gd.crawlStartTime = time.Now()

	if gd.sub {
		gd.robotsMap.Clear()
		gd.domainsCrawled.Clear()
	}
}

func (gd *GlobalData) AddUrl(url string, crawlData UrlToCrawlData) {
	gd.urlsToCrawl.Set(url, crawlData)
}

func (gd *GlobalData) ToCrawlCount() int {
	return gd.urlsToCrawl.Count()
}

func (gd *GlobalData) CrawledCount() int {
	return gd.urlsCrawled.Count()
}

func (gd *GlobalData) IsCrawling() bool {
	return !gd.urlsToCrawl.IsEmpty()
}

func (gd *GlobalData) StartCrawlTime() time.Time {
	return gd.crawlStartTime
}

// CrawlContext supports concurrency
// TODO: Separate the thread-specific info from the universal info
type CrawlContext struct {
	client         gemini.Client
	nex_client     nex_client.Client
	scroll_client  scroll.Client
	spartan_client spartan_client.Client
	resp           Response
	currentURL     *neturl.URL
	isRootPage     bool // If hostname == currentURL
	currentRobots  Robots
	globalData     *GlobalData
}

var timeout, _ = time.ParseDuration("10m")

func newCrawlContext(globalData *GlobalData) CrawlContext {
	url, _ := neturl.Parse("gemini://gemini.circumlunar.space/")
	return CrawlContext{gemini.Client{NoTimeCheck: true, ReadTimeout: timeout, ConnectTimeout: 15 * time.Second}, nex_client.Client{ReadTimeout: timeout, ConnectTimeout: 15 * time.Second}, scroll.Client{ReadTimeout: timeout, ConnectTimeout: 15 * time.Second, NoTimeCheck: true}, spartan_client.Client{ConnectTimeout: 15 * time.Second, ReadTimeout: timeout}, Response{}, url, true, Robots{}, globalData}
}

// GetCurrentURL should always return the URL with a slash after the hostname
func (ctx *CrawlContext) GetCurrentURL() string {
	var buf strings.Builder

	// Hostname
	addPort := _getAddPort(ctx.currentURL)
	buf.WriteString(ctx.currentURL.Scheme)
	buf.WriteString("://")
	buf.WriteString(ctx.currentURL.Hostname())
	if addPort {
		buf.WriteByte(':')
		buf.WriteString(ctx.currentURL.Port())
	}

	// Path
	path := ctx.currentURL.EscapedPath()
	if path == "" || (path != "" && path[0] != '/' && ctx.currentURL.Host != "") {
		buf.WriteByte('/')
	}
	buf.WriteString(path)

	// Queries and Fragments
	if ctx.currentURL.ForceQuery || ctx.currentURL.RawQuery != "" {
		buf.WriteByte('?')
		buf.WriteString(ctx.currentURL.RawQuery)
	}
	if ctx.currentURL.Fragment != "" {
		buf.WriteByte('#')
		buf.WriteString(ctx.currentURL.EscapedFragment())
	}

	return buf.String()
}

func _getAddPort(URL *neturl.URL) bool {
	if URL.Port() != "" {
		if URL.Port() != _getPortStringFromScheme(URL.Scheme) {
			return true
		}
	}

	return false
}

func _getPortStringFromScheme(scheme string) string {
	if scheme == "gemini" || scheme == "titan" {
		return "1965"
	} else if scheme == "gopher" || scheme == "gophers" {
		return "70"
	} else if scheme == "nex" {
		return "1900"
	} else if scheme == "spartan" {
		return "300"
	} else if scheme == "scroll" {
		return "5699"
	}

	return "1965"
}

func _getPortFromScheme(scheme string) int {
	if scheme == "gemini" || scheme == "titan" {
		return 1965
	} else if scheme == "gopher" || scheme == "gophers" {
		return 70
	} else if scheme == "nex" {
		return 1900
	} else if scheme == "spartan" {
		return 300
	} else if scheme == "scroll" {
		return 5699
	}

	return 1965
}

// NOTE: Includes the scheme and port, with a trailing slash
func (ctx *CrawlContext) GetCurrentHostname() string {
	host := ""
	addPort := _getAddPort(ctx.currentURL)
	if addPort {
		host = ctx.currentURL.Scheme + "://" + ctx.currentURL.Hostname() + ":" + ctx.currentURL.Port() + "/"
	} else {
		host = ctx.currentURL.Scheme + "://" + ctx.currentURL.Hostname() + "/"
	}

	return host
}

// GetDomain gets the Current Domain from the context
func (ctx *CrawlContext) GetDomain() Domain {
	port := _getPortFromScheme(ctx.currentURL.Scheme)
	if ctx.currentURL.Port() != "" {
		parsed, err := strconv.Atoi(ctx.currentURL.Port())
		if err == nil {
			port = parsed
		}
	}

	domain := ctx.currentURL.Hostname()
	hasRobots := false
	if _, ok := ctx.globalData.robotsMap.Get(ctx.GetCurrentHostname()); ok {
		hasRobots = true
	}

	return Domain{0, domain, "", port, 0, hasRobots, false, false, "", CrawlIndex, time.Now().UTC(), 0} // NOTE: Added .UTC()
}

var ErrNilURL = errors.New("URL is Nil")

func GetHostname(url string) (string, error) {
	currentURL, err := neturl.Parse(url)
	if err != nil {
		return "", err
	} else if currentURL == nil {
		return "", ErrNilURL
	}
	host := ""

	addPort := _getAddPort(currentURL)
	if addPort {
		host = currentURL.Scheme + "://" + currentURL.Hostname() + ":" + currentURL.Port() + "/"
	} else {
		host = currentURL.Scheme + "://" + currentURL.Hostname() + "/"
	}

	return host, nil
}

func (ctx *CrawlContext) addUrl(url string, crawlData UrlToCrawlData) {
	//var exists = struct{}{}
	//c.urlsToCrawl[url] = exists
	ctx.globalData.urlsToCrawl.Set(url, crawlData)
}

// Map of domains and their DomainInfo (slowDown time, and lastCrawlTime)
// Returns true if already existed
func (ctx *CrawlContext) addDomain(domain string) (bool, DomainInfo) {
	//var exists = struct{}{}
	//_, preexists := c.domainsCrawled[domain]
	//c.domainsCrawled[domain] = exists
	//return preexists

	if r, ok := ctx.globalData.domainsCrawled.Get(domain); ok {
		// Domain already exists
		return true, r.(DomainInfo)
	} else {
		domainInfo := DomainInfo{defaultSlowDown, time.Now().UTC()} // NOTE: Added .UTC()
		ctx.globalData.domainsCrawled.Set(domain, domainInfo)
		return false, domainInfo
	}
}

func (ctx *CrawlContext) getDomain(domain string) (bool, DomainInfo) {
	r, ok := ctx.globalData.domainsCrawled.Get(domain)
	return ok, r.(DomainInfo)
}

func (ctx *CrawlContext) removeUrl(url string) {
	//var exists = struct{}{}
	//delete(c.urlsToCrawl, url)
	ctx.globalData.urlsToCrawl.Remove(url)
	//c.urlsCrawled[url] = exists
	ctx.globalData.urlsCrawled.Set(url, Page{})
}

func (ctx *CrawlContext) setUrlCrawledPageData(url string, page Page) {
	ctx.globalData.urlsCrawled.Set(url, page)
}

/*
func (c *CrawlContext) removeDomain(domain string) {
	//delete(c.domainsCrawled, domain)
	c.domainsCrawled.Remove(domain)
}
*/

func (ctx *CrawlContext) getNextUrl() (string, UrlToCrawlData) {
	// Randomize getting from the map a little bit
	var skip int64 = 0
	if ctx.globalData.urlsToCrawl.Count() > 1 {
		skip = rand.Int63n(int64(ctx.globalData.urlsToCrawl.Count() - 1)) // TODO
	}
	var i int64 = 0
	for k := range ctx.globalData.urlsToCrawl.IterBuffered() {
		if i < skip {
			i++
			continue
		}
		i++

		//c.urlsToCrawl.Remove(k.Key)
		removed := ctx.globalData.urlsToCrawl.RemoveCb(k.Key, func(key string, v any, exists bool) bool {
			return exists
			/*if !exists {
				return false
			}*/

			// Check if url's hostname can be crawled at this time. If not, don't remove it and continue on to next url
			/*host := GetHostname(k.Key)
			var crawlDelay time.Duration = time.Second
			if r, ok := c.robotsMap.Get(host); ok {
				crawlDelay = r.(Robots).indexerGroup.CrawlDelay
				allow := r.(Robots).indexerGroup.Test(c.currentURL.Path)
				if !allow {
					// Not allowed by robots.txt. Skip this URL
					return false
				}
				c.currentRobots = r.(Robots)
			}*/

			// Check if enough time has passed to be able to crawl
			/*if domainExists, domainInfo := c.getDomain(host); domainExists {
				elapsed := time.Since(domainInfo.lastCrawlTime).Seconds()
				if elapsed >= domainInfo.slowDown && elapsed >= float64(crawlDelay.Seconds()) {
					// set the new lastCrawlTime
					newDomainInfo := DomainInfo{math.Max(domainInfo.slowDown, crawlDelay.Seconds()), time.Now().UTC()}
					c.domainsCrawled.Set(host, newDomainInfo)
					return exists
				}
			}*/
			// return false
		})
		if !removed {
			continue
		}

		// If already crawled, continue to next url and leave this one as removed
		if _, exists := ctx.globalData.urlsCrawled.Get(k.Key); exists {
			continue
		}

		// Set as crawled
		ctx.globalData.urlsCrawled.Set(k.Key, Page{CrawlIndex: CrawlIndex, Date_added: time.Now().UTC()})
		return k.Key, k.Val.(UrlToCrawlData) // TODO: k.Value should be the URLToCrawlData
	}

	return "", UrlToCrawlData{}
}

// GetRobotsTxt gets the robots of a host. The given host must have "/" at the end
func (ctx *CrawlContext) GetRobotsTxt(host string) (Robots, error) {
	// Defaults
	robotsData := Robots{}
	robotsStr := "User-agent: *\nAllow: /"
	robotsData.robots, _ = robotstxt.FromString(robotsStr) // TODO: Add domains with robots.txt problems to a database table to keep track of them

	if strings.HasPrefix(host, "gemini://") {
		resp, err := ctx.client.Fetch(host + "robots.txt")
		if err != nil && strings.HasSuffix(err.Error(), "bind: An operation on a socket could not be performed because the system lacked sufficient buffer space or because a queue was full.") {
			logError("Gemini Get Error on robots.txt: %s; %v", err.Error(), err)
			return Robots{}, err
		}

		if err != nil || resp == nil {
			//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", robotsStr)
		} else if resp.Status == 44 {
			return Robots{}, errors.New("Slow down.")
		} else if resp.Status != gemini.StatusSuccess {
			//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", robotsStr)
			resp.Body.Close()
		} else {
			data, read_err := io.ReadAll(resp.Body)
			dataStr := string(data)

			if !strings.Contains(dataStr, "User-Agent:") {
				// If data doesn't contain "User-Agent:" anywhere, then prepend "User-Agent: *\n" to it.
				dataStr = "User-Agent: *\n" + dataStr
			}
			if read_err == nil {
				robotsData.robots, _ = robotstxt.FromString(dataStr)
				//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", dataStr)
			}
			resp.Body.Close()
		}
	} else if strings.HasPrefix(host, "nex://") {
		conn, err := ctx.nex_client.Request(host + "robots.txt")
		if err != nil && strings.HasSuffix(err.Error(), "bind: An operation on a socket could not be performed because the system lacked sufficient buffer space or because a queue was full.") {
			logError("Nex Get Error on robots.txt: %s; %v", err.Error(), err)
			return Robots{}, err
		} else if err != nil {
		} else {
			data, read_err := io.ReadAll(conn)
			dataStr := string(data)

			if !strings.Contains(dataStr, "User-Agent:") {
				// If data doesn't contain "User-Agent:" anywhere, then prepend "User-Agent: *\n" to it.
				dataStr = "User-Agent: *\n" + dataStr
			}
			if read_err == nil {
				robotsData.robots, _ = robotstxt.FromString(dataStr)
				//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", dataStr)
			}
			conn.Close()
		}
	} else if strings.HasPrefix(host, "scroll://") {
		resp, err := ctx.scroll_client.Fetch(host+"robots.txt", []string{"en"}, false)
		if err != nil && strings.HasSuffix(err.Error(), "bind: An operation on a socket could not be performed because the system lacked sufficient buffer space or because a queue was full.") {
			logError("Scroll Get Error on robots.txt: %s; %v", err.Error(), err)
			return Robots{}, err
		}

		if err != nil || resp == nil {
			//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", robotsStr)
		} else if resp.Status == 44 {
			return Robots{}, errors.New("Slow down.")
		} else if scroll.CleanStatus(resp.Status) != 20 {
			//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", robotsStr)
			resp.Body.Close()
		} else {
			data, read_err := io.ReadAll(resp.Body)
			dataStr := string(data)

			if !strings.Contains(dataStr, "User-Agent:") {
				// If data doesn't contain "User-Agent:" anywhere, then prepend "User-Agent: *\n" to it.
				dataStr = "User-Agent: *\n" + dataStr
			}
			if read_err == nil {
				robotsData.robots, _ = robotstxt.FromString(dataStr)
				//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", dataStr)
			}
			resp.Body.Close()
		}
	} else if strings.HasPrefix(host, "spartan://") {
		resp, err := ctx.spartan_client.Request(host+"robots.txt", []byte{})
		if err != nil && strings.HasSuffix(err.Error(), "bind: An operation on a socket could not be performed because the system lacked sufficient buffer space or because a queue was full.") {
			logError("Scroll Get Error on robots.txt: %s; %v", err.Error(), err)
			return Robots{}, err
		}

		if err != nil || resp == nil {
			//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", robotsStr)
		} else if resp.Status == 44 {
			// Return error when there's a slowdown
			return Robots{}, errors.New("Slow down.")
		} else if resp.Status != spartan_client.StatusSuccess {
			//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", robotsStr)
			resp.Body.Close()
		} else {
			data, read_err := io.ReadAll(resp.Body)
			dataStr := string(data)

			if !strings.Contains(dataStr, "User-Agent:") {
				// If data doesn't contain "User-Agent:" anywhere, then prepend "User-Agent: *\n" to it.
				dataStr = "User-Agent: *\n" + dataStr
			}
			if read_err == nil {
				robotsData.robots, _ = robotstxt.FromString(dataStr)
				//fmt.Printf("Robots: %s\n%s\n\n", host+"robots.txt", dataStr)
			}
			resp.Body.Close()
		}
	}

	// TODO: FindGroup fails when there's a robots but there's no group specified in it with "User-Agent:"
	robotsData.indexerGroup = robotsData.robots.FindGroup("indexer")
	//robotsData.allGroup = robotsData.robots.FindGroup("*")

	// Add robots.txt to robotsMap
	ctx.globalData.robotsMap.Set(host, robotsData)
	return robotsData, nil
}

// Get gemini page data
// Be sure to call cancel
func (ctx *CrawlContext) Get(url string, crawlThread int, crawlData UrlToCrawlData) (Response, error) {
	ctx.currentURL, _ = neturl.Parse(url)
	if ctx.currentURL.Scheme != "gemini" && ctx.currentURL.Scheme != "nex" && ctx.currentURL.Scheme != "scroll" && ctx.currentURL.Scheme != "spartan" {
		//c.removeUrl(url) // TODO
		return Response{}, ErrNotSupportedScheme
	}

	// Check if url already crawled
	/*if _, ok := c.urlsCrawled[c.GetCurrentURL()]; ok {
		fmt.Printf("Skipping Already Crawled URL\n")
		return nil, ErrAlreadyCrawled
	}*/

	// Get host (with scheme and port and trailing slash)
	host := ctx.GetCurrentHostname()

	// Set domain information
	ctx.isRootPage = false
	if host == ctx.GetCurrentURL() {
		ctx.isRootPage = true
		//fmt.Printf("Is Root Page[%d]! %s == %s\n", crawlThread, host, c.GetCurrentURL())
	}

	var crawlDelay time.Duration
	// Check if host is in robotsMap. If not, get robots.txt. If so, check if allowed to crawl, and return if not.
	if r, ok := ctx.globalData.robotsMap.Get(host); ok {
		crawlDelay = r.(Robots).indexerGroup.CrawlDelay
		allow := r.(Robots).indexerGroup.Test(ctx.currentURL.Path)
		if !allow {
			//c.removeUrl(url)
			return Response{}, ErrNotAllowed
		}
		ctx.currentRobots = r.(Robots)
	} else {
		// Get robots.txt and insert into map if exists
		r, err := ctx.GetRobotsTxt(host)
		if err != nil { // Robots.txt couldn't be fetched, because of no space in buffer (due to socket TIME_WAITs)
			ctx.globalData.urlsCrawled.Remove(url)
			ctx.addUrl(url, crawlData)
			return Response{}, err
		}
		crawlDelay = r.indexerGroup.CrawlDelay
		allow := r.indexerGroup.Test(ctx.currentURL.Path)
		if !allow {
			//c.removeUrl(url)
			return Response{}, ErrNotAllowed
		}
		ctx.currentRobots = r
		time.Sleep((time.Duration(crawlDelay.Milliseconds()) * time.Millisecond) + time.Second)
	}

	// Default to crawl delay of 2 seconds
	if crawlDelay == time.Duration(0) {
		crawlDelay = time.Duration(defaultSlowDown) * time.Second
	}

	domainPreexists, domainInfo := ctx.addDomain(host)
	if !domainPreexists { // TODO
		ctx.addUrl(host, UrlToCrawlData{})
	} else {
		// Check if enough time has passed
		elapsed := time.Since(domainInfo.lastCrawlTime).Seconds()
		if elapsed >= domainInfo.slowDown && elapsed >= crawlDelay.Seconds() {
			// Go Ahead and crawl the page and set the new lastCrawlTime
			newDomainInfo := DomainInfo{math.Max(domainInfo.slowDown, crawlDelay.Seconds()), time.Now().UTC()} // NOTE: Added .UTC()
			ctx.globalData.domainsCrawled.Set(host, newDomainInfo)
			//fmt.Printf("Allowed Crawl: %s; %v\n", ctx.GetCurrentURL(), newDomainInfo)
		} else {
			// Skip the page, add the url back in, and remove from urlsCrawled
			//fmt.Printf("Not Allowed to Crawl Yet: %s (%fs); %v\n", ctx.GetCurrentURL(), domainInfo.slowDown, domainInfo)
			ctx.globalData.urlsCrawled.Remove(url)
			ctx.addUrl(url, crawlData)
			return Response{}, ErrSlowDown
		}
	}

	//c.removeUrl(url)
	var err error
	var resp Response
	if ctx.currentURL.Scheme == "gemini" {
		var g_resp *gemini.Response
		g_resp, err = ctx.client.Fetch(url)
		if err == nil {
			resp.Status = g_resp.Status
			resp.Description = g_resp.Meta
			if resp.Description == "" && (resp.Status >= 20 && resp.Status <= 29) {
				if beforeQuery, _, _ := strings.Cut(url, "?"); strings.HasSuffix(beforeQuery, ".scroll") {
					resp.Description = "text/scroll"
				} else if strings.HasSuffix(beforeQuery, ".gmi") || strings.HasSuffix(beforeQuery, ".gemini") {
					resp.Description = "text/gemini"
				} else if strings.HasSuffix(beforeQuery, "/") {
					resp.Description = "text/gemini"
				}
			}
			resp.Body = g_resp.Body
			resp.Cert = g_resp.Cert
			ctx.resp = resp
		}
	} else if ctx.currentURL.Scheme == "nex" {
		var conn net.Conn
		conn, err = ctx.nex_client.Request(url)
		if err == nil {
			resp.Status = 20
			resp.Description = ""
			if beforeQuery, _, _ := strings.Cut(url, "?"); strings.HasSuffix(beforeQuery, "/") {
				// Nex Listings use a "/" at the end of the path.
				resp.Description = "text/nex"
			}
			resp.Body = conn
			resp.Cert = nil
			ctx.resp = resp
		}
	} else if ctx.currentURL.Scheme == "scroll" {
		var s_resp *scroll.Response
		s_resp, err = ctx.scroll_client.Fetch(url, []string{}, false)
		if err == nil {
			resp.Status = s_resp.Status
			resp.Description = s_resp.Description
			if resp.Description == "" && (resp.Status >= 20 && resp.Status <= 29) {
				if beforeQuery, _, _ := strings.Cut(url, "?"); strings.HasSuffix(beforeQuery, ".scroll") {
					resp.Description = "text/scroll"
				} else if strings.HasSuffix(beforeQuery, ".gmi") || strings.HasSuffix(beforeQuery, ".gemini") {
					resp.Description = "text/gemini"
				} else if strings.HasSuffix(beforeQuery, "/") {
					resp.Description = "text/scroll"
				}
			}
			resp.Body = s_resp.Body
			resp.Author = s_resp.Author
			resp.PublishDate = s_resp.PublishDate
			resp.ModificationDate = s_resp.ModificationDate
			resp.Cert = s_resp.Cert
			ctx.resp = resp
		}
	} else if ctx.currentURL.Scheme == "spartan" {
		var s_resp *spartan_client.Response
		s_resp, err = ctx.spartan_client.Request(url, []byte{})
		if err == nil {
			if s_resp.Status == spartan_client.StatusSuccess {
				resp.Status = 20
			} else if s_resp.Status == spartan_client.StatusRedirect {
				resp.Status = gemini.StatusRedirect
			} else if s_resp.Status == 4 { // Client Error
				resp.Status = gemini.StatusBadRequest
			} else if s_resp.Status == 5 { // Server Error
				resp.Status = gemini.StatusTemporaryFailure
			}
			resp.Description = s_resp.Meta
			if resp.Description == "" && (resp.Status >= 20 && resp.Status <= 29) {
				if beforeQuery, _, _ := strings.Cut(url, "?"); strings.HasSuffix(beforeQuery, ".scroll") {
					resp.Description = "text/scroll"
				} else if strings.HasSuffix(beforeQuery, ".gmi") || strings.HasSuffix(beforeQuery, ".gemini") {
					resp.Description = "text/gemini"
				} else if strings.HasSuffix(beforeQuery, "/") {
					resp.Description = "text/gemini"
				}
			}
			resp.Body = s_resp.Body
			ctx.resp = resp
		}
	}

	return resp, err
}
