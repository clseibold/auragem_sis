package crawler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

type Seed struct {
	Id         int
	Url        string
	Title      string // the Title already in the DB
	Date_added time.Time
}

type Domain struct {
	Id     int
	Domain string
	Title  string
	Port   int
	//ParentDomain Domain // ForeignKey
	ParentDomainId int
	//Robots string // contents of robots.txt?
	HasRobots     bool
	HasSecurity   bool
	HasFavicon    bool
	Favicon       string
	CrawlIndex    int
	Date_added    time.Time
	Slowdowncount int
}

// Domains can have multiple schemes
/*type DomainScheme struct {
	Id int
	//Domain Domain
	DomainId int
	Scheme string
}*/

type Page struct {
	Id     int
	Url    string // fetchable_url, normalized_url
	Scheme string
	// Domain Domain // foreign key
	DomainId interface{}

	Content_type string
	Charset      string
	Language     string
	Linecount    int
	Udc          string

	Title string // Used for text/gemini and text/markdown files with page titles
	// content []u8 // TODO
	Prompt      string // For input prompt urls
	Headings    string
	Size        int // bytes
	Hash        string
	Feed        bool      // rss, atom, or gmisub
	PublishDate time.Time // Used if linked from a feed, or if audio/video with year tag
	Index_time  time.Time

	// Audio/Video-only info
	Album               string
	Artist              string
	AlbumArtist         string
	Composer            string
	Track               int
	Disc                int
	Copyright           string
	CrawlIndex          int
	Date_added          time.Time
	LastSuccessfulVisit time.Time

	Hidden bool

	HasDuplicateOnGemini bool
}

type Tag struct {
	Id int
	//Page Page
	PageId     int
	Name       string
	Rank       int
	CrawlIndex int
	Date_added time.Time
}

type Link struct {
	Id int
	//From Page
	FromPageId int
	//To Page
	ToPageId   int
	Title      string
	Cross_host bool
	CrawlIndex int
	Date_added time.Time
}

func logError(format string, a ...interface{}) {
	f, err := os.OpenFile("errors.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString("[" + time.Now().String() + "] " + fmt.Sprintf(format, a...) + "\n"); err != nil {
		log.Println(err)
	}
}

func GetSeeds(gd *GlobalData) []Seed {
	q := `SELECT id, url, date_added FROM seeds ORDER BY date_added ASC`

	rows, rows_err := gd.dbConn.QueryContext(context.Background(), q)

	var seeds []Seed
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var seed Seed
			scan_err := rows.Scan(&seed.Id, &seed.Url, &seed.Date_added)
			if scan_err == nil {
				seeds = append(seeds, seed)
			} else {
				panic(scan_err)
			}
		}
	}

	return seeds
}

func GetFeedsAsSeeds(gd *GlobalData) []Seed {
	q := `SELECT id, url, title, date_added FROM pages WHERE feed = true AND hidden = false`
	//q := `SELECT id, url, date_added FROM seeds ORDER BY date_added ASC`

	rows, rows_err := gd.dbConn.QueryContext(context.Background(), q)

	var seeds []Seed
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var seed Seed
			scan_err := rows.Scan(&seed.Id, &seed.Url, &seed.Title, &seed.Date_added)
			if scan_err == nil {
				seeds = append(seeds, seed)
			} else {
				panic(scan_err)
			}
		}
	}

	return seeds
}

func setPageToHidden(ctx CrawlContext, URL string) {
	if !utf8.ValidString(URL) {
		logError("Error from Page: Page Url not valid utf8; %v", URL)
		return
	}

	parsedUrl, parseErr := url.Parse(URL)
	if parseErr != nil {
		panic(parseErr)
	}

	// Check if exists in db, then update. Otherwise, don't add it in the first place
	row := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM pages WHERE url=?", URL)
	count := 0
	err := row.Scan(&count)
	if err != sql.ErrNoRows && err != nil { // TODO
		panic(err)
		//return
	}
	if err == sql.ErrNoRows || count <= 0 {
		// Don't insert page
		return
	} else if count > 0 {
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "UPDATE pages SET scheme=?, indextime=?, crawlIndex=?, last_successful_visit=?, hidden=true WHERE url=?", strings.ToLower(strings.TrimSuffix(parsedUrl.Scheme, "://")), time.Now().UTC(), CrawlIndex, time.Now().UTC(), URL)
		if err != nil {
			fmt.Printf("Error from Page URL: %v\n", URL)
			panic(err)
		}
	}
}

func addPageToDb(ctx CrawlContext, page Page) (Page, bool) {
	if !utf8.ValidString(page.Title) {
		logError("Error from Page: Page Title not valid utf8; %v", page)
		return Page{}, false
	}
	if !utf8.ValidString(page.Url) {
		logError("Error from Page: Page Url not valid utf8; %v", page)
		return Page{}, false
	}
	//titleGraphemeCount := uniseg.GraphemeClusterCount(page.Title)
	if len(page.Title) > 250 {
		logError("Error from Page: Title over 250 characters; %v", page)
		return Page{}, false
	}

	// Check if exists in db, then update or insert
	row := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM pages WHERE url=?", page.Url)
	count := 0
	err := row.Scan(&count)
	if err != sql.ErrNoRows && err != nil { // TODO
		logError("Error from Page: %v\n%v\n", page, err.Error())
		return Page{}, false
		//panic(err)
		//return Page{}, false
	}
	if page.DomainId == 0 {
		fmt.Printf("Page's DomainId is 0: %v\n", page)
		panic("DomainId Value Cannot Be Zero")
	}
	if err == sql.ErrNoRows || count <= 0 {
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "INSERT INTO pages (url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, headings, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlIndex, date_added, last_successful_visit, hidden, has_duplicate_on_gemini) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", page.Url, page.Scheme, page.DomainId, page.Content_type, page.Charset, page.Language, page.Linecount, page.Udc, page.Title, page.Prompt, page.Headings, page.Size, page.Hash, page.Feed, page.PublishDate, time.Now().UTC(), page.Album, page.Artist, page.AlbumArtist, page.Composer, page.Track, page.Disc, page.Copyright, CrawlIndex, time.Now().UTC(), time.Now().UTC(), page.Hidden, page.HasDuplicateOnGemini)
		if err != nil {
			logError("Error from Page: %v\n%v\n", page, err.Error())
			return Page{}, false
			/*fmt.Printf("Error from Page: %v\n", page)
			panic(err)*/
		}
	} else if count > 0 {
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "UPDATE pages SET scheme=?, domainid=?, contenttype=?, charset=?, language=?, linecount=?, udc=?, title=?, prompt=?, headings=?, size=?, hash=?, feed=?, publishdate=?, indextime=?, album=?, artist=?, albumartist=?, composer=?, track=?, disc=?, copyright=?, crawlIndex=?, last_successful_visit=?, hidden=?, has_duplicate_on_gemini=? WHERE url=?", page.Scheme, page.DomainId, page.Content_type, page.Charset, page.Language, page.Linecount, page.Udc, page.Title, page.Prompt, page.Headings, page.Size, page.Hash, page.Feed, page.PublishDate, time.Now().UTC(), page.Album, page.Artist, page.AlbumArtist, page.Composer, page.Track, page.Disc, page.Copyright, CrawlIndex, time.Now().UTC(), page.Hidden, page.HasDuplicateOnGemini, page.Url)
		if err != nil {
			logError("Error from Page: %v\n%v\n", page, err.Error())
			return Page{}, false
			/*fmt.Printf("Error from Page: %v\n", page)
			panic(err)*/
		}
	}

	// Get the page
	var result Page
	row2 := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT FIRST 1 id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, headings, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden, has_duplicate_on_gemini FROM pages WHERE url=?", page.Url)
	row2.Scan(&result.Id, &result.Url, &result.Scheme, &result.DomainId, &result.Content_type, &result.Charset, &result.Language, &result.Linecount, &result.Udc, &result.Title, &result.Prompt, &result.Headings, &result.Size, &result.Hash, &result.Feed, &result.PublishDate, &result.Index_time, &result.Album, &result.Artist, &result.AlbumArtist, &result.Composer, &result.Track, &result.Disc, &result.Copyright, &result.CrawlIndex, &result.Date_added, &result.LastSuccessfulVisit, &result.Hidden, &result.HasDuplicateOnGemini)
	return result, true
}

func getPagesWithHashAndScheme(ctx CrawlContext, url string, pageHash string, scheme string) []Page {
	query := "SELECT id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden, has_duplicate_on_gemini FROM pages WHERE url<>? AND hash=?"
	if scheme != "" {
		query += " AND scheme=? AND hidden=false"
	}

	var rows *sql.Rows
	var err error
	if scheme != "" {
		rows, err = ctx.globalData.dbConn.Query(query, url, pageHash, scheme)
	} else {
		rows, err = ctx.globalData.dbConn.Query(query, url, pageHash)
	}

	var pages []Page = make([]Page, 0, 1)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden, &page.HasDuplicateOnGemini)
			if scan_err == nil {
				pages = append(pages, page)
			} else {
				prevPage := Page{}
				if len(pages) > 0 {
					prevPage = pages[len(pages)-1]
				}
				panic(fmt.Errorf("scan error after page %v; %s", prevPage, scan_err.Error()))
			}
		}
		if err := rows.Err(); err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}

	return pages
}

func getPagesWithHashAndNotScheme(ctx CrawlContext, url string, pageHash string, scheme string) []Page {
	query := "SELECT id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden, has_duplicate_on_gemini FROM pages WHERE url<>? AND hash=?"
	if scheme != "" {
		query += " AND scheme<>?"
	}

	var rows *sql.Rows
	var err error
	if scheme != "" {
		rows, err = ctx.globalData.dbConn.Query(query, url, pageHash, scheme)
	} else {
		rows, err = ctx.globalData.dbConn.Query(query, url, pageHash)
	}

	var pages []Page = make([]Page, 0, 1)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden, &page.HasDuplicateOnGemini)
			if scan_err == nil {
				pages = append(pages, page)
			} else {
				prevPage := Page{}
				if len(pages) > 0 {
					prevPage = pages[len(pages)-1]
				}
				panic(fmt.Errorf("scan error after page %v; %s", prevPage, scan_err.Error()))
			}
		}
		if err := rows.Err(); err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}

	return pages
}

// Sets has_duplicate_on_gemini to true on all pages of schemes outside of 'gemini' with the given hash.
func setPageHashHasGeminiDuplicate(ctx CrawlContext, url string, pageHash string, value bool) {
	_, err := ctx.globalData.dbConn.Exec("UPDATE pages SET has_duplicate_on_gemini=? WHERE url<>? AND hash=? AND scheme<>'gemini'", value, url, pageHash)
	if err != sql.ErrNoRows && err != nil { // TODO
		panic(err)
		//return
	}
}

func domainIncrementSlowDownCount(ctx CrawlContext, domain Domain) {
	// Check if exists in db, then update or insert
	row := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM domains WHERE domain=?", domain.Domain)
	count := 0
	err := row.Scan(&count)
	if err != sql.ErrNoRows && err != nil {
		fmt.Printf("Error from Domain: %v\n", domain)
		panic(err)
	}
	if err == sql.ErrNoRows || count <= 0 { // Insert domain
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "INSERT INTO domains (domain, title, port, has_robots, has_favicon, has_security, crawlIndex, date_added, slowdowncount, emptymetacount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", domain.Domain, domain.Title, domain.Port, domain.HasRobots, domain.HasSecurity, domain.HasFavicon, CrawlIndex, time.Now().UTC(), 1, 0)
		if err != nil {
			fmt.Printf("Error from Domain: %v\n", domain)
			panic(err)
		}
	} else if count > 0 { // Otherwise, just increment the slowdowncount
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "UPDATE domains SET slowdowncount=slowdowncount+1 WHERE domain=?", domain.Domain)
		if err != nil {
			fmt.Printf("Error from Domain: %v\n", domain)
			panic(err)
		}
	}
}

func domainIncrementEmptyMeta(ctx CrawlContext, domain Domain) {
	// Check if exists in db, then update or insert
	row := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM domains WHERE domain=?", domain.Domain)
	count := 0
	err := row.Scan(&count)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		fmt.Printf("Error from Domain: %v\n", domain)
		panic(err)
	}
	if errors.Is(err, sql.ErrNoRows) || count <= 0 { // Insert domain
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "INSERT INTO domains (domain, title, port, has_robots, has_favicon, has_security, crawlIndex, date_added, slowdowncount, emptymetacount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", domain.Domain, domain.Title, domain.Port, domain.HasRobots, domain.HasSecurity, domain.HasFavicon, CrawlIndex, time.Now().UTC(), 0, 1)
		if err != nil {
			fmt.Printf("Error from Domain: %v\n", domain)
			panic(err)
		}
	} else if count > 0 { // Otherwise, just increment the emptymetacount
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "UPDATE domains SET emptymetacount=emptymetacount+1 WHERE domain=?", domain.Domain)
		if err != nil {
			fmt.Printf("Error from Domain: %v\n", domain)
			panic(err)
		}
	}
}

func addDomainToDb(ctx CrawlContext, domain Domain, update bool) (Domain, bool) {
	if !utf8.ValidString(domain.Title) {
		logError("Error from Domain: Domain Title not valid utf8; %v", domain)
		return Domain{}, false
	}
	//titleGraphemeCount := uniseg.GraphemeClusterCount(domain.Title)
	if len(domain.Title) > 250 {
		logError("Error from Domain: Domain Title over 250 characters; %v", domain)
		return Domain{}, false
	}

	// Check if exists in db, then update or insert
	row := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM domains WHERE domain=? AND port=?", domain.Domain, domain.Port)
	count := 0
	err := row.Scan(&count)
	if err != sql.ErrNoRows && err != nil {
		fmt.Printf("Error from Domain: %v\n", domain)
		panic(err)
	}
	if err == sql.ErrNoRows || count <= 0 {
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "INSERT INTO domains (domain, title, port, has_robots, has_favicon, has_security, crawlIndex, date_added, slowdowncount, emptymetacount) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", domain.Domain, domain.Title, domain.Port, domain.HasRobots, domain.HasSecurity, domain.HasFavicon, CrawlIndex, time.Now().UTC(), 0, 0)
		if err != nil {
			fmt.Printf("Error from Domain: %v\n", domain)
			panic(err)
		}
	} else if count > 0 && update {
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "UPDATE domains SET title=?, has_robots=?, has_security=?, has_favicon=?, crawlIndex=? WHERE domain=? AND port=?", domain.Title, domain.HasRobots, domain.HasSecurity, domain.HasFavicon, CrawlIndex, domain.Domain, domain.Port)
		if err != nil {
			fmt.Printf("Error from Domain: %v\n", domain)
			panic(err)
		}
	}

	// Get the domain
	var result Domain
	row2 := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT FIRST 1 id, domain, title, port, has_robots, has_security, has_favicon, crawlIndex, date_added FROM domains WHERE domain=? AND port=?", domain.Domain, domain.Port)
	err_result := row2.Scan(&result.Id, &result.Domain, &result.Title, &result.Port, &result.HasRobots, &result.HasSecurity, &result.HasFavicon, &result.CrawlIndex, &result.Date_added)
	if err_result != nil {
		fmt.Printf("Get Error from Domain: %v\n", domain)
		panic(err_result)
	}
	return result, true
}

func addLinkToDb(ctx CrawlContext, link Link) (Link, bool) {
	if !utf8.ValidString(link.Title) {
		logError("Error from Link: Link Title not valid utf8; %v", link)
		return Link{}, false
	}
	//titleGraphemeCount := uniseg.GraphemeClusterCount(link.Title)
	if len(link.Title) > 250 {
		logError("Error from Link: Title over 250 characters; %v", link)
		return Link{}, false
	}

	// Check if exists in db, then update or insert
	row := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM links WHERE pageid_from=? AND pageid_to=?", link.FromPageId, link.ToPageId)
	count := 0
	err := row.Scan(&count)
	if err != sql.ErrNoRows && err != nil { // TODO
		panic(err)
		//return Page{}, false
	}
	if link.FromPageId == 0 || link.ToPageId == 0 {
		logError("Link's From/To Page Id is 0: %v\n", link)
		//panic("DomainId Value Cannot Be Zero")
	}
	if err == sql.ErrNoRows || count <= 0 {
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "INSERT INTO links (pageid_from, pageid_to, title, crosshost, crawlIndex, date_added) VALUES (?, ?, ?, ?, ?, ?)", link.FromPageId, link.ToPageId, link.Title, link.Cross_host, CrawlIndex, time.Now().UTC())
		if err != nil {
			fmt.Printf("Error from Link: %v\n", link)
			panic(err)
		}
	} else if count > 0 {
		_, err := ctx.globalData.dbConn.ExecContext(context.Background(), "UPDATE links SET title=?, crawlIndex=? WHERE pageid_from=? AND pageid_to=?", link.Title, CrawlIndex, link.FromPageId, link.ToPageId)
		if err != nil {
			fmt.Printf("Error from Link: %v\n", link)
			panic(err)
		}
	}

	// Get the link
	var result Link
	row2 := ctx.globalData.dbConn.QueryRowContext(context.Background(), "SELECT FIRST 1 id, pageid_from, pageid_to, title, crosshost, crawlindex, date_added FROM links WHERE pageid_from=? AND pageid_to=?", link.FromPageId, link.ToPageId)
	row2.Scan(&result.Id, &result.FromPageId, &result.ToPageId, &result.Title, &result.Cross_host, &result.CrawlIndex, &result.Date_added)
	return result, true
}
