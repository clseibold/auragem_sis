package search

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"
	// "strconv"
)

func getRecent(conn *sql.DB) []Page {
	q := `SELECT FIRST 50 id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE hidden=false ORDER BY date_added DESC`

	rows, rows_err := conn.QueryContext(context.Background(), q)

	var pages []Page = make([]Page, 0, 50)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
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
	}

	return pages
}

func getCapsules(conn *sql.DB) []Domain {
	q := `SELECT COUNT(*) OVER () as total, id, domain, title, port, has_robots, has_security, has_favicon, favicon, crawlindex, date_added FROM domains ORDER BY date_added DESC`

	rows, rows_err := conn.QueryContext(context.Background(), q)

	var capsules []Domain = nil
	if rows_err == nil {
		var count int64
		defer rows.Close()
		for rows.Next() {
			var domain Domain
			scan_err := rows.Scan(&count, &domain.Id, &domain.Domain, &domain.Title, &domain.Port, &domain.HasRobots, &domain.HasSecurity, &domain.HasFavicon, &domain.Favicon, &domain.CrawlIndex, &domain.Date_added)
			if scan_err == nil {
				if capsules == nil {
					capsules = make([]Domain, 0, count)
				}
				capsules = append(capsules, domain)
			} else {
				panic(scan_err)
			}
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}
	}

	return capsules
}

// NOTE: Only tags used more than twice are included
func getTags(conn *sql.DB) []Tag {
	q := `SELECT tags.name, COUNT(*) as c FROM tags WHERE tags.name NOT LIKE 'boycottnovell%' GROUP BY tags.name HAVING COUNT(*) > 2 ORDER BY c DESC`

	rows, rows_err := conn.QueryContext(context.Background(), q)

	var tags []Tag
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var tag Tag
			scan_err := rows.Scan(&tag.Name, &tag.Count)
			if scan_err == nil {
				tags = append(tags, tag)
			} else {
				panic(scan_err)
			}
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}
	}

	return tags
}

// TODO: Set a way to order the results
func getPagesOfTag(conn *sql.DB, name string) []Page {
	q := `SELECT COUNT(*) OVER () as total, pages.id, pages.url, pages.scheme, pages.domainid, pages.contenttype, pages.charset, pages.language, pages.linecount, pages.udc, pages.title, pages.prompt, pages.size, pages.hash, pages.feed, pages.publishdate, pages.indextime, pages.album, pages.artist, pages.albumartist, pages.composer, pages.track, pages.disc, pages.copyright, pages.crawlindex, pages.date_added, pages.last_successful_visit, pages.hidden FROM tags JOIN pages ON pages.id = tags.pageid where tags.name=?`

	rows, rows_err := conn.QueryContext(context.Background(), q, name)

	var pages []Page = nil
	if rows_err == nil {
		var count int64
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&count, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
			if scan_err == nil {
				if pages == nil {
					pages = make([]Page, 0, count)
				}
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
	}

	return pages
}

func getMimetypeFiles(conn *sql.DB, mimetype string) []Page {
	q := `SELECT FIRST 300 id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE contenttype=? AND hidden=false`

	rows, rows_err := conn.QueryContext(context.Background(), q, mimetype)

	var pages []Page = make([]Page, 0, 300)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
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
	}

	return pages
}

type MimetypeListItem struct {
	mimetype string
	count    int
}

func getMimetypes(conn *sql.DB) []MimetypeListItem {
	var mimetypes []MimetypeListItem = make([]MimetypeListItem, 0, 50)
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT contenttype, COUNT(*) FROM pages WHERE hidden=false GROUP BY contenttype ORDER BY COUNT(*) DESC")
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var item MimetypeListItem
			scan_err := rows.Scan(&item.mimetype, &item.count)
			if scan_err == nil {
				mimetypes = append(mimetypes, item)
			} else {
				panic(scan_err)
			}
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}
	}

	return mimetypes
}

func getFeeds(conn *sql.DB) []Page {
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT COUNT(*) OVER () as total, id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE feed = true AND hidden = false")

	var pages []Page = nil
	if rows_err == nil {
		var count int64
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&count, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
			if scan_err == nil {
				if pages == nil {
					pages = make([]Page, 0, count)
				}
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
	}

	return pages
}

func getPagesWithPublishDate(conn *sql.DB) []Page {
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT COUNT(*) OVER () as total, id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE publishdate <> ? AND hidden = false ORDER BY publishdate DESC", time.Time{})

	var pages []Page = nil
	if rows_err == nil {
		var count int64
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&count, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
			if scan_err == nil {
				if pages == nil {
					pages = make([]Page, 0, count)
				}
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
	}

	return pages
}

// TODO: Allow for different languages
// NOTE: Blank language fields are considered English
func getPagesWithPublishDateFromLastYear(conn *sql.DB, results int, skip int) ([]Page, int) {
	query := fmt.Sprintf("SELECT FIRST %d SKIP %d COUNT(*) OVER () totalCount, id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE publishdate > dateadd(-1 year to ?) AND publishdate < dateadd(2 day to ?) AND (language = '' OR language LIKE 'en%%') AND hidden = false ORDER BY publishdate DESC", results, skip)
	rows, rows_err := conn.QueryContext(context.Background(), query, time.Now().UTC(), time.Now().UTC())

	var pages []Page = make([]Page, 0, results)
	var totalCount int
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&totalCount, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
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
	}

	return pages, totalCount
}

// Returns []Page, totalResultsCount, and whether there's a next page
func getAudioFiles(conn *sql.DB, page int64) ([]Page, int64, bool) {
	var results int64 = 30
	skip := (page - 1) * results
	q := fmt.Sprintf(`SELECT FIRST %d SKIP %d COUNT(*) OVER () totalCount, id, url, scheme, domainid, contenttype, charset, language, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE contenttype IN ('audio/mpeg', 'audio/mp3', 'audio/ogg', 'audio/flac', 'audio/mid', 'audio/m4a', 'audio/x-flac') AND hidden = false`, results, skip)

	rows, rows_err := conn.QueryContext(context.Background(), q)

	var pages []Page = make([]Page, 0, results)
	var totalCount int64
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&totalCount, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
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
	}

	return pages, totalCount, skip+results < totalCount
}

// Returns []Page, totalResultsCount, and whether there's a next page
func getImageFiles(conn *sql.DB, page int64) ([]Page, int64, bool) {
	var results int64 = 30
	skip := (page - 1) * results
	q := fmt.Sprintf(`SELECT FIRST %d SKIP %d COUNT(*) OVER () totalCount, id, url, scheme, domainid, contenttype, charset, language, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE contenttype IN ('image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/bmp', 'image/webp', 'image/svg+xml', 'image/vnd.mozilla.apng') AND hidden = false`, results, skip)

	rows, rows_err := conn.QueryContext(context.Background(), q)

	var pages []Page = make([]Page, 0, results)
	var totalCount int64
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&totalCount, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
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
	}

	return pages, totalCount, skip+results < totalCount
}

func getTwtxtFiles(conn *sql.DB) []Page {
	q := `SELECT COUNT(*) OVER () as total, id, url, scheme, domainid, contenttype, charset, language, linecount, udc, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE (url LIKE '%twtxt.txt' OR url LIKE '%tw.txt') AND hidden = false`

	rows, rows_err := conn.QueryContext(context.Background(), q)

	var pages []Page = nil
	if rows_err == nil {
		var count int64
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&count, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
			if scan_err == nil {
				if pages == nil {
					pages = make([]Page, 0, count)
				}
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
	}

	return pages
}

func getSecurityTxtFiles(conn *sql.DB) []PageWithDomain {
	q := `SELECT COUNT(*) OVER () as total, pages.id, pages.url, pages.scheme, pages.domainid, pages.contenttype, pages.charset, pages.language, pages.linecount, pages.udc, pages.title, pages.prompt, pages.size, pages.hash, pages.feed, pages.publishdate, pages.indextime, pages.album, pages.artist, pages.albumartist, pages.composer, pages.track, pages.disc, pages.copyright, pages.crawlindex, pages.date_added, pages.last_successful_visit, pages.hidden, domains.id, domains.domain, domains.title, domains.port, domains.has_robots, domains.has_security, domains.has_favicon, domains.favicon, domains.crawlindex, domains.date_added FROM pages INNER JOIN domains ON domains.ID = pages.domainid WHERE pages.url LIKE '%security.txt' AND hidden = false`

	rows, rows_err := conn.QueryContext(context.Background(), q)

	var pages []PageWithDomain = nil
	if rows_err == nil {
		var count int64
		defer rows.Close()
		for rows.Next() {
			var page Page
			var domain Domain
			scan_err := rows.Scan(&count, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden, &domain.Id, &domain.Domain, &domain.Title, &domain.Port, &domain.HasRobots, &domain.HasSecurity, &domain.HasFavicon, &domain.Favicon, &domain.CrawlIndex, &domain.Date_added)
			if scan_err == nil {
				if pages == nil {
					pages = make([]PageWithDomain, 0, count)
				}
				pages = append(pages, PageWithDomain{page, domain})
			} else {
				prevPage := Page{}
				if len(pages) > 0 {
					prevPage = pages[len(pages)-1].page
				}
				panic(fmt.Errorf("scan error after page %v; %s", prevPage, scan_err.Error()))
			}
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}
	}

	return pages
}

var InvalidURLString = errors.New("URL is not a valid UTF-8 string.")
var URLTooLong = errors.New("URL exceeds 1024 bytes.")
var InvalidURL = errors.New("URL is not valid.")
var URLRelative = errors.New("URL is relative. Only absolute URLs can be added.")
var URLNotGemini = errors.New("Must be a Gemini URL.")

func addSeedToDb(conn *sql.DB, seed Seed) (Seed, error) {
	// Make sure URL is a valid UTF-8 string
	if !utf8.ValidString(seed.Url) {
		return Seed{}, InvalidURLString
	}
	// Make sure URL doesn't exceed 1024 bytes
	if len(seed.Url) > 1024 {
		return Seed{}, URLTooLong
	}
	// Make sure URL has gemini:// scheme
	if !strings.HasPrefix(seed.Url, "gemini://") && !strings.HasPrefix(seed.Url, "nex://") && !strings.Contains(seed.Url, "://") && !strings.HasPrefix(seed.Url, ".") && !strings.HasPrefix(seed.Url, "/") {
		seed.Url = "gemini://" + seed.Url
	}

	// Make sure the url is parseable and that only the hostname is being added
	u, urlErr := url.Parse(seed.Url)
	if urlErr != nil { // Check if able to parse
		return Seed{}, InvalidURL
	}
	if !u.IsAbs() { // Check if Absolute URL
		return Seed{}, URLRelative
	}
	if u.Scheme != "gemini" && u.Scheme != "nex" { // Make sure scheme is gemini or nex
		return Seed{}, URLNotGemini
	}
	seed.Url = _getHostname(u)
	if !strings.ContainsRune(seed.Url, '.') { // Check that there's a TLD (e.g. .com, .org, .io, etc)
		return Seed{}, InvalidURL
	}

	// Check if exists in db, then update or insert
	row := conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM seeds WHERE url=?", seed.Url)
	count := 0
	err := row.Scan(&count)
	if err != sql.ErrNoRows && err != nil { // TODO
		//panic(err)
		return Seed{}, err
	}
	if err == sql.ErrNoRows || count <= 0 {
		_, err := conn.ExecContext(context.Background(), "INSERT INTO seeds (url, date_added) VALUES (?, ?)", seed.Url, time.Now().UTC())
		if err != nil {
			//fmt.Printf("Error from Page: %v\n", page)
			//panic(err)
			return Seed{}, err
		}
	} else if count > 0 {
		// Already exists, do nothing
	}

	// Get the seed
	var result Seed
	row2 := conn.QueryRowContext(context.Background(), "SELECT FIRST 1 id, url, date_added FROM seeds WHERE url=?", seed.Url)
	row2.Scan(&result.Id, &result.Url, &result.Date_added)
	return result, nil
}

func _getHostname(url *url.URL) string {
	host := ""

	if url.Port() == "" || url.Port() == "1965" {
		host = url.Scheme + "://" + url.Hostname() + "/"
	} else {
		host = url.Scheme + "://" + url.Hostname() + ":" + url.Port() + "/"
	}

	return host
}
