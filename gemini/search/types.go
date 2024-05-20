package search

import (
	"database/sql"
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
	ParentDomainId sql.Null[int64]
	//Robots string // contents of robots.txt?
	HasRobots   bool
	HasSecurity bool
	HasFavicon  bool
	Favicon     sql.Null[string]
	CrawlIndex  int
	Date_added  time.Time
}

type Page struct {
	Score  float64
	Id     int64
	Url    string // fetchable_url, normalized_url
	Scheme string
	// Domain Domain // foreign key
	DomainId sql.Null[int64]

	Content_type string
	Charset      string
	Language     string
	Linecount    int
	Udc          string

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

type PageWithDomain struct {
	page   Page
	domain Domain
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
