package search

import (
	"time"
)

type Seed struct {
	Id         int
	Url        string
	Date_added time.Time
}

type Domain struct {
	Id     int
	Domain string
	Title  string
	Port   int
	//ParentDomain Domain // ForeignKey
	ParentDomainId interface{}
	//Robots string // contents of robots.txt?
	HasRobots   bool
	HasSecurity bool
	HasFavicon  bool
	Favicon     interface{}
	CrawlIndex  int
	Date_added  time.Time
}

type Page struct {
	Score  float64
	Id     int64
	Url    string // fetchable_url, normalized_url
	Scheme string
	// Domain Domain // foreign key
	DomainId int64

	Content_type string
	Charset      string
	Language     string
	Linecount    int

	Title string // Used for text/gemini and text/markdown files with page titles
	// content []u8 // TODO
	Prompt      string // For input prompt urls
	Headings    string // Empty unless specifically queried for as we don't want to query this from the DB due to potential large size
	Size        int    // bytes
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

	Highlight string // Used for highlights when searching
}

type pageNullable struct {
	Score  float64
	Id     int64
	Url    string // fetchable_url, normalized_url
	Scheme string
	// Domain Domain // foreign key
	DomainId interface{}

	Content_type string
	Charset      string
	Language     string
	Linecount    int

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

	Highlight string // Used for highlights when searching
}

type PageWithDomain struct {
	page   Page
	domain Domain
}

func scanPage(page pageNullable) Page {
	var result Page
	result.Score = page.Score
	result.Id = page.Id
	result.Url = page.Url

	result.Scheme = page.Scheme
	if page.DomainId != nil {
		if _, ok := page.DomainId.(int64); ok {
			result.DomainId = int64(page.DomainId.(int64))
		} else {
			if _, ok2 := page.DomainId.(int); ok2 {
				result.DomainId = int64(page.DomainId.(int))
			}
		}
	} else {
		result.DomainId = -1
	}

	result.Content_type = page.Content_type
	result.Charset = page.Charset
	result.Language = page.Language
	result.Linecount = page.Linecount

	result.Title = page.Title
	result.Prompt = page.Prompt
	result.Headings = page.Headings
	result.Size = page.Size
	result.Hash = page.Hash
	result.Feed = page.Feed
	result.PublishDate = page.PublishDate
	result.Index_time = page.Index_time

	result.Album = page.Album
	result.Artist = page.Artist
	result.AlbumArtist = page.AlbumArtist
	result.Composer = page.Composer
	result.Track = page.Track
	result.Disc = page.Disc
	result.Copyright = page.Copyright
	result.CrawlIndex = page.CrawlIndex
	result.Date_added = page.Date_added
	result.LastSuccessfulVisit = page.LastSuccessfulVisit

	result.Hidden = page.Hidden

	result.Highlight = page.Highlight

	return result
}

type Tag struct {
	Id         int
	PageId     int
	Name       string
	Rank       float64
	CrawlIndex int
	Date_added time.Time

	Count int
}

type Backlink struct {
	Id           int
	PageId_From  int
	PageURL_FROM string
	Title        string
	Crosshost    bool
	CrawlIndex   int
	Date_added   time.Time
}
