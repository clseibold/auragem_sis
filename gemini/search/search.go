package search

// TODO: Add Favicons text to database and have crawler look for favicon.txt file
// TOOD: track Hashtags and Mentions (mentions can start with @ or ~)

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	wiki "github.com/trietmn/go-wiki"
	// "gitlab.com/clseibold/auragem_sis/crawler"
	"gitlab.com/clseibold/auragem_sis/crawler"
	"gitlab.com/clseibold/auragem_sis/db"
	sis "gitlab.com/clseibold/smallnetinformationservices"
	"golang.org/x/text/language"
)

// %%query%% replaced with the exact query the user entered, escaped and put in quotes
// %%matches%% replaced with something like `t.NAME = 'term1' OR t.NAME = 'term2'`
// Search query will rank domain root pages higher if they match the query

// Search from all protocols
var fts_searchQuery string = `
select FIRST %%first%% SKIP %%skip%% COUNT(*) OVER () totalCount, (FTS.FTS$SCORE) as GROUPED_SCORE, P.ID, P.URL, P.SCHEME, P.DOMAINID, P.CONTENTTYPE, P.CHARSET, P.LANGUAGE, P.LINECOUNT, P.UDC, P.TITLE, P.PROMPT, P.SIZE, P.HASH, P.FEED, CASE WHEN EXTRACT(YEAR FROM P.PUBLISHDATE) < 1800 THEN TIMESTAMP '01.01.9999 00:00:00.000' ELSE P.PUBLISHDATE END AS PUBLISHDATE, P.INDEXTIME, P.ALBUM, P.ARTIST, P.ALBUMARTIST, P.COMPOSER, P.TRACK, P.DISC, P.COPYRIGHT, P.CRAWLINDEX, P.DATE_ADDED, P.LAST_SUCCESSFUL_VISIT, P.HIDDEN
    FROM FTS$SEARCH('FTS_PAGE_ID_EN', '(%%query%%) AND HIDDEN:false') FTS
    JOIN PAGES P ON P.ID = FTS.FTS$ID
    WHERE P.HAS_DUPLICATE_ON_GEMINI=false
	ORDER BY GROUPED_SCORE DESC, CHAR_LENGTH(P.URL) ASC, PUBLISHDATE DESC
`

// Search from a specific protocol
var fts_searchQuery_protocol string = `
select FIRST %%first%% SKIP %%skip%% COUNT(*) OVER () totalCount, (FTS.FTS$SCORE) as GROUPED_SCORE, P.ID, P.URL, P.SCHEME, P.DOMAINID, P.CONTENTTYPE, P.CHARSET, P.LANGUAGE, P.LINECOUNT, P.UDC, P.TITLE, P.PROMPT, P.SIZE, P.HASH, P.FEED, CASE WHEN EXTRACT(YEAR FROM P.PUBLISHDATE) < 1800 THEN TIMESTAMP '01.01.9999 00:00:00.000' ELSE P.PUBLISHDATE END AS PUBLISHDATE, P.INDEXTIME, P.ALBUM, P.ARTIST, P.ALBUMARTIST, P.COMPOSER, P.TRACK, P.DISC, P.COPYRIGHT, P.CRAWLINDEX, P.DATE_ADDED, P.LAST_SUCCESSFUL_VISIT, P.HIDDEN
    FROM FTS$SEARCH('FTS_PAGE_ID_EN', '(%%query%%) AND HIDDEN:false AND SCHEME:%%protocol%%') FTS
    JOIN PAGES P ON P.ID = FTS.FTS$ID
	ORDER BY GROUPED_SCORE DESC, CHAR_LENGTH(P.URL) ASC, PUBLISHDATE DESC
`

var fts_audioSearchQuery string = `
select FIRST %%first%% SKIP %%skip%% COUNT(*) OVER () totalCount, SUM(s.SCORE) as GROUPED_SCORE, s.HIGHLIGHT, s.ID, s.URL, s.SCHEME, s.DOMAINID, s.CONTENTTYPE, s.CHARSET, s.LANGUAGE, s.LINECOUNT, s.UDC, s.TITLE, s.PROMPT, s.SIZE, s.HASH, s.FEED, s.PUBLISHDATE, s.INDEXTIME, s.ALBUM, s.ARTIST, s.ALBUMARTIST, s.COMPOSER, s.TRACK, s.DISC, s.COPYRIGHT, s.CRAWLINDEX, s.DATE_ADDED, s.LAST_SUCCESSFUL_VISIT, s.HIDDEN
FROM (select FTS.FTS$ID as fts_id, FTS.FTS$SCORE as SCORE,
        FTS$HIGHLIGHTER.FTS$BEST_FRAGMENT(A.TEXT, '%%query%%', 'ENGLISH', 'TEXT', 70, '[', ']') AS HIGHLIGHT,
        P.*
    FROM FTS$SEARCH('FTS_AUDIOTRANSCRIPT_ID_EN', '%%query%%') FTS
    JOIN AUDIOTRANSCRIPTS A ON A.ID = FTS.FTS$ID
    JOIN PAGES P ON A.PAGEID = P.ID
	) s
WHERE s.HIDDEN = false
GROUP BY HIGHLIGHT, ID, URL, SCHEME, DOMAINID, CONTENTTYPE, CHARSET, LANGUAGE, LINECOUNT, TITLE, PROMPT, SIZE, HASH, FEED, PUBLISHDATE, INDEXTIME, ALBUM, ARTIST, ALBUMARTIST, COMPOSER, TRACK, DISC, COPYRIGHT, CRAWLINDEX, DATE_ADDED, LAST_SUCCESSFUL_VISIT, HIDDEN
ORDER BY GROUPED_SCORE DESC
`

func HandleSearchEngineDown(s sis.ServerHandle) {
	s.AddRoute("/searchengine", func(request sis.Request) {
		request.Redirect("/search/")
	})
	s.AddRoute("/searchengine/", func(request sis.Request) {
		request.Redirect("/search/")
	})
	s.AddRoute("/search", func(request sis.Request) {
		request.Redirect("/search/")
	})
	s.AddRoute("/search/*", func(request sis.Request) {
		request.ServerUnavailable("AuraGem Search is currently down due to upgrades.")
	})
}

func UdcClassStringToShortTitle(udc string) string {
	switch udc {
	case "0":
		return "Science and Knowledge. Computer Science. Documentation."
	case "1":
		return "Philosophy. Psychology"
	case "2":
		return "Religion. Theology. Scripture"
	case "3":
		return "Social Sciences"
	case "4":
		return "Misc/Unclassed"
	case "5":
		return "Mathematics. Natural Sciences"
	case "6":
		return "Applied Sciences. Medicine. General Technology"
	case "7":
		return "The Arts. Recreation. Entertainment. Sport"
	case "8":
		return "Language. Linguistics. Literature. Personal logs"
	case "9":
		return "Geography. Biography. History"
	default:
		return "Unknown"
	}
}

func UdcClassStringToTitle(udc string) string {
	switch udc {
	case "0":
		return "Science and Knowledge. Organization. Computer Science and Computer Technology. Information. Documentation. Librarianship. Institutions. Publications"
	case "1":
		return "Philosophy. Psychology"
	case "2":
		return "Religion. Theology. Scripture"
	case "3":
		return "Social Sciences"
	case "4":
		return "Misc/Unclassed"
	case "5":
		return "Mathematics. Natural Sciences"
	case "6":
		return "Applied Sciences. Medicine. General Technology"
	case "7":
		return "The Arts. Recreation. Entertainment. Sport"
	case "8":
		return "Language. Linguistics. Literature. Personal logs"
	case "9":
		return "Geography. Biography. History"
	default:
		return "Unknown"
	}
}

func HandleSearchEngine(s sis.ServerHandle) {
	//crawling := make(map[string]bool, 10)
	//crawling_ips := make(map[string]bool, 10)

	conn := db.NewConn(db.SearchDB)
	conn.SetMaxOpenConns(100)
	conn.SetConnMaxIdleTime(0)
	conn.SetMaxIdleConns(8 + 10)
	conn.SetConnMaxLifetime(0)

	// Crawler - full crawl every month, feed crawl every 13 hours, and on-demand capsule crawling
	globalData := crawler.NewGlobalData(conn, true, true, 0) // Follows all links
	go crawler.RegularCrawler(globalData, nil)
	go crawler.FeedCrawler(globalData, 13, nil)

	// On-Demand Capsule Crawler
	capsulesCrawling := make(map[string]bool, 10)
	addToCrawlerChan := make(chan string, 1000)
	crawlerMutex := &sync.RWMutex{}
	cond := sync.NewCond(crawlerMutex.RLocker())
	currentCapsuleCrawl := ""
	go func() {
		for {
			crawlerMutex.Lock()
			url := <-addToCrawlerChan
			capsulesCrawling[url] = true
			crawlerMutex.Unlock()
			cond.Broadcast()
		}
	}()
	go func() {
		currentCapsuleCrawl = ""
		for {
			crawlerMutex.RLock()
			// If no capsules to crawl, wait until there is a capsule to crawl.
			for len(capsulesCrawling) == 0 {
				cond.Wait()
			}
			fmt.Printf("[2000] Starting capsule crawl.\n")
			for k := range capsulesCrawling {
				currentCapsuleCrawl = k
				break
			}
			crawler.OnDemandCapsuleCrawl(globalData, currentCapsuleCrawl, "")
			delete(capsulesCrawling, currentCapsuleCrawl)
			currentCapsuleCrawl = ""
			crawlerMutex.RUnlock()
		}
	}()

	// Outdated Link Handles
	s.AddRoute("/searchengine", func(request sis.Request) {
		request.Redirect("/search/")
	})
	s.AddRoute("/searchengine/", func(request sis.Request) {
		request.Redirect("/search/")
	})
	s.AddRoute("/searchengine/search", func(request sis.Request) {
		request.Redirect("/search/s")
	})
	s.AddRoute("/searchengine/random", func(request sis.Request) {
		request.Redirect("/search/random")
	})
	s.AddRoute("/searchengine/capsules", func(request sis.Request) {
		request.Redirect("/search/capsules")
	})
	s.AddRoute("/searchengine/tags", func(request sis.Request) {
		request.Redirect("/search/tags")
	})
	s.AddRoute("/searchengine/mimetype", func(request sis.Request) {
		request.Redirect("/search/mimetype")
	})
	s.AddRoute("/searchengine/recent", func(request sis.Request) {
		request.Redirect("/search/recent")
	})
	s.AddRoute("/searchengine/yearposts", func(request sis.Request) {
		request.Redirect("/search/yearposts")
	})
	s.AddRoute("/searchengine/feeds", func(request sis.Request) {
		request.Redirect("/search/feeds")
	})
	s.AddRoute("/searchengine/audio", func(request sis.Request) {
		request.Redirect("/search/audio")
	})
	s.AddRoute("/searchengine/images", func(request sis.Request) {
		request.Redirect("/search/images")
	})
	s.AddRoute("/searchengine/twtxt", func(request sis.Request) {
		request.Redirect("/search/twtxt")
	})
	s.AddRoute("/searchengine/security", func(request sis.Request) {
		request.Redirect("/search/security")
	})

	// Search Engine Handles
	s.AddRoute("/search", func(request sis.Request) {
		request.Redirect("/search/")
	})
	// TODO: Removed Tag Index (=> /search/tags 🏷️ Tag Index)
	publishDate, _ := time.ParseInLocation(time.RFC3339, "2021-07-01T00:00:00", time.Local)
	updateDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-13T00:00:00", time.Local)
	s.AddRoute("/search/", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		request.Gemini("# AuraGem Search\n\n")
		request.PromptLine("/search/s/", "🔍 Search Smallnet")
		request.PromptLine("/search/gemini/", "🔍 Search Geminispace")
		request.PromptLine("/search/scroll/", "🔍 Search Scrollspace")
		request.PromptLine("/search/spartan/", "🔍 Search Spartanspace")
		request.Gemini(`
=> /search/scrollspace Scrollspace Index
=> /search/random/ 🎲 Goto Random Capsule
=> /search/backlinks/ Check Backlinks

=> /search/yearposts/ 📌 Recent Publications
=> /search/recent/ 50 Most Recently Indexed
=> /search/capsules/ 🪐 Recently Discovered Capsules
=> /search/mimetype/ Mimetypes

=> /search/features/ About and Features
=> /search/stats/ 📈 Statistics
=> /search/crawl/ Missing your capsule? Add it to AuraGem Search

=> /search/feeds/ 🗃 Indexed Feeds
=> /search/audio/ 🎵 Indexed Audio Files
=> /search/images/ 🖼️ Indexed Image Files
=> /search/twtxt/ 📝 Indexed Twtxt Files
=> /search/security/ 📃 Indexed Security.txt Files

=> /search/configure_default/ Configure Default Search Engine in Lagrange

Note that AuraGem Search does not ensure or rank based on the popularity or accuracy of the information within any of the pages listed in these search results. One cannot presume that information published within Geminispace is or is not for ill-intent or misinformation, even if it's popular or well-linked, so one must use their best judgement in determining the trustworthiness of such content themselves.

## Other Search Engines

=> gemini://kennedy.gemi.dev/ Kennedy (PageRank)
=> gemini://tlgs.one/ TLGS - "Totally Legit" Gemini Search
=> gemini://gemplex.space/ Gemplex (PageRank)
=> gemini://geminispace.info geminispace.info (GUS)

## Compendiums

=> gemini://smol.earth/compendium/ Smol Earth Compendium

## Aggregators

=> gemini://warmedal.se/~antenna/ Antenna
=> gemini://skyjake.fi/~Cosmos/ Cosmos
=> gemini://calcuode.com/gmisub-aggregate.gmi GmiSub
=> gemini://gemini.circumlunar.space/capcom/ Capcom
=> gemini://rawtext.club/~sloum/spacewalk.gmi SpaceWalk

`)

		// ## Support
		//
		// Want to help support the project? Consider donating on the Patreon. The first goal is to get a server from a server hosting provider that could better support all of the projects I have planned.
		//
		// => https://www.patreon.com/krixano Patreon
	})

	s.AddRoute("/search/crawl", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Docs, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# Crawl a link\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Enter a page/capsule to crawl:")
			return
		} else {
			queryUrl, parse_err := url.Parse(query)
			if parse_err != nil {
				//request.Redirect("/search/crawl")
				request.TemporaryFailure("Unable to parse URL.")
				return
			}
			queryUrl.Fragment = "" // Strip the fragment
			if (queryUrl.Scheme != "gemini" && queryUrl.Scheme != "nex" && queryUrl.Scheme != "spartan" && queryUrl.Scheme != "scroll") || !queryUrl.IsAbs() {
				request.TemporaryFailure("Please enter only a Gemini, Nex, Spartan, or Scroll URL.")
				return
			}
			if queryUrl.Path == "" {
				queryUrl.Path = "/"
			}

			/*if _, ok := crawling_ips[request.IPHash()]; ok {
				request.TemporaryFailure("You have already submitted a URL to crawl.")
				return
			}*/

			// Add the capsule's root to the seeds table.
			seedUrl := *queryUrl
			seedUrl.Path = "/"
			addSeedToDb(conn, Seed{0, seedUrl.String(), time.Time{}})
			// Add to regular crawler in case it's not already there.
			globalData.AddUrl(seedUrl.String(), crawler.UrlToCrawlData{})
			request.Gemini(fmt.Sprintf("# Capsule submitted\n\n'%s' has been submitted for on-demand crawling.\n=> /search/ AuraGem Search Home\n", queryUrl.String()))

			// Add to capsule on-demand crawler
			addToCrawlerChan <- seedUrl.String()
		}
	})

	// Goes through all pages that haven't been crawled in a long time and sets them to hidden if they are no longer available
	/*inCleanup := false
	s.AddRoute("/search/cleanup", func(request sis.Request) {
		if inCleanup {
			request.TemporaryFailure("Cleanup already running.")
			return
		}

		go func(inCleanup *bool) {
			*inCleanup = true
			globalData := crawler.NewGlobalData(conn, false, false, 0)
			crawler.PurgeOldLinks(globalData)
			*inCleanup = false
		}(&inCleanup)

		request.Redirect("/search/")
	})*/

	s.AddRoute("/search/configure_default", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Docs, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# Configure Default Search Engine in Lagrange\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		request.Gemini(`# Configure Default Search Engine in Lagrange

1. Go to File -> Preferences -> General
2. Paste the following link into the Search URL field:
> gemini://auragem.ddns.net/search/s
`)
	})

	s.AddRoute("/search/features", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Docs, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search Features\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		request.Gemini(`# AuraGem Search Features

## Current State of Features
* Full Text Search of page and file metadata, with Stemming, because apparently other search engines think it's important and unique to advertise one of the most common features in searching systems, lol.
* Complex search queries using AND, OR, and NOT operators, as well as grouping using parentheses and quotes for multiword search terms. By default, if you do not use any of these operators, search terms are combined using OR, much like you would expect from web search engines. However, searches that have all the terms provided will still be ranked higher than searches with just one or a portion of the terms provided.
* + and - operators. + is for a required term, - is for a search term that must not be matched.

* Title extraction using first apparent heading, regardless of its level.
* Can detect gemsub feeds.
* Line Counts of text files, and publication dates indexed based on dates in filenames.
* File size information
* Mp3, Ogg, and Flac file metadata (ID3, MP4, and Ogg/Flac) is indexed.
* A feed of Posts from Past Year organized based on publication date, from most recent to least recent.

* Filters include "TITLE", "URL", "ALBUM", "ARTIST", "ALBUMARTIST", "COPYRIGHT", "CONTENTTYPE", "LANGUAGE", and "PUBLISHDATE", as well as others that are untested. The syntax is "field: term". You can also use groups for filters. Field names must be in all capital letters.
* Wildcards * and ?
* Fuzzy Searching by placing ~ after a search term
* Proximity Searching: if you want to search for two words that are within a distance of 10 words of each other, then query with "term_one term_two"~10
* Range Searching: For searching in ranges of numbers or dates. Can be used with filters, like the PUBLISHDATE filter. An example of filtering based on a publication date range would be, PUBLISHDATE:[20220101 to 20231201]

* Crawler: Robots.txt is followed, including "Allow", "Disallow", and "Crawl-Delay" directives. The Slow Down gemini status code is also followed.
* Crawler: 2 second delay between crawling of pages on the same domain.

## Features Coming Soon
* PDF and Djvu file metadata indexed
* Image file metadata indexed
* Plain text file full contents indexed
* Backlinks and searching of link text
* Page Metadata Lookup
* Full Markdown, Tinylog, and Twtxt parsing to get links, titles, and heading information.

## History

AuraGem was a search engine that I started about 2 years ago under its original name, Ponix Search. It was originally designed to experiment with how I could make search results better. The official announcement of the Search Engine happened on 2021-07-01:
=> /devlog/20210701.gmi 2021-07-01 Search Engine & Ponix Capsule Now Open Source (MIT)
=> /devlog/20211205.gmi 2021-12-05 AuraGem Search Begins Crawling Again

Note that some of the information in the above posts have been recently updated to match the current URL and Ip Address of the crawler and gemini capsule.

One of the first priorities with AuraGem Search was to have extraction of file metadata for as many files as possible. Audio files were one of the first to get this feature. PDFs and Djvu files were supposed to be next, and support was added for them on 2022-07-19, but the feature was buggy and never worked, unfortunately. As you can see in the below post, I chose to go with Keyword Extraction (which was later removed and replaced with simple mentions and tags extraction) instead of Full Text Searching on page contents. Part of this was to save space, and part of it was to respect copyright. However, I am rethinking this approach now that the Stats page can determine how large the text-only portion of geminispace is (no more than 5GB total).
=> /devlog/20220719.gmi 2022-07-19 AuraGem Search Engine Update
=> /search/stats/ Stats Page

In the above article, you can see that I start to play with the notion of different types of searches. I think this idea remains important today:
> Another problem that the above process would not catch are names and proper nouns. These are often very important words that people would want to search for (e.g. Mathematics, C++, Celine Dion, FTS). I do not have an easy method for this atm.

The next update on 2022-07-21 added Full Text Searching of link and file metadata, which drastically improved the speed of searches. Yes, this came with stemming because my database's FTS uses Lucene++.
=> /devlog/20220720_search.gmi 2022-07-21 AuraGem Search Update

Not long after I wrote an article about FTS, ranking systems, and some of the problems that Search Engines have to handle:
=> /devlog/20220722.gmi 2022-07-22 Search Engine Ranking Systems Are Being Left Unquestioned

The most important portion of this article, however, is recognizing how people do searches:
> This also introduces the argument that the ranking systems are really only important for underspecified queries (broad queries), so the emphasis on the problems with ranking algorithms is unwarranted. This argument hardly makes sense when the majority of searches that people make are broad. I would also argue that broad searches are most used for *discovering* pages, not for getting to a specific page. However, ranking based on popularity prioritizes what it thinks people would want, which is more suited for specific searches using broad queries, at the expense of discovery of broad topics. Broad discovery using broad topic queries and specific searches using proper-noun queries or very specific queries are both much better ways of dealing with searches without relying on popularity.

When making a search engine, one must balance the search results between discovery (broadness) and exact matches (exactness). Relevancy applies to both of these, but is more important for discovery. I continue to think that link analysis assumes that people want exact matches of pages while using broad queries. For example, if someone types in "search engine", a PageRank system would put the most popular search engine at the top along with popular articles about search engines, assuming that the person wanted that specific search engine, when it's more likely they wanted a collection of search engines. Rather, my approach is to return broad relevant discovery-based results with broad queries, and exact pages with exact queries.

Exact queries include words from titles, domain names, capsule names, service names, basically mainly proper nouns or a specific combination of words that matches the page information. Broad queries, however, use category names and common nouns.

When I type "Station", I want an exact match for Station itself. However, when I type "social network", I want search results that give a very broad set of capsules that are social networks. I believe that this is how most people would use search engines, especially if they do not rely much on filtering, and this is the exact methodology that I use for my article analyzing gemini's search engines:
=> /devlog/20220807.gmi 2022-08-07 Gemini Search Results Study, Part 1

`)
	})

	/*s.AddRoute("/search/index", func(request sis.Request) {
		handleSearchIndex(request, conn)
	})*/

	s.AddRoute("/search/scrollspace/", func(request sis.Request) {
		request.Gemini(`# Scrollspace Index

=> /search/scrollspace/0/ 0. Science and Knowledge. Organization. Computer Science and Computer Technology. Information. Documentation. Librarianship. Institutions. Publications
=> /search/scrollspace/1/ 1. Philosophy. Psychology
=> /search/scrollspace/2/ 2. Religion. Theology. Scripture
=> /search/scrollspace/3/ 3. Social Sciences
=> /search/scrollspace/4/ 4. Misc/Unclassed
=> /search/scrollspace/5/ 5. Mathematics. Natural Sciences
=> /search/scrollspace/6/ 6. Applied Sciences. Medicine. General Technology
=> /search/scrollspace/7/ 7. The Arts. Recreation. Entertainment. Sport
=> /search/scrollspace/8/ 8. Language. Linguistics. Literature. Personal logs
=> /search/scrollspace/9/ 9. Geography. Biography. History

=> gemini://scrollprotocol.us.to/ About the Scroll Protocol
=> /search/ Search Home
`)
	})

	s.AddRoute("/search/scrollspace/:class/", func(request sis.Request) {
		classStr := request.GetParam("class")
		if classStr != "0" && classStr != "1" && classStr != "2" && classStr != "3" && classStr != "4" && classStr != "5" && classStr != "6" && classStr != "7" && classStr != "8" && classStr != "9" {
			request.BadRequest("UDC class doesn't exist.")
			return
		}

		pages, hasNextPage, _ := getPagesOfUDC(conn, 1, classStr)

		request.Gemini(fmt.Sprintf("# Scrollspace Index: %s. %s\n", classStr, UdcClassStringToShortTitle(classStr)))
		request.Gemini("\n=> /search/scrollspace/ Scrollspace Index Home\n\n")

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)
		request.Gemini(builder.String() + "\n")

		if hasNextPage {
			request.Gemini(fmt.Sprintf("\n=> /search/scrollspace/%s/2/ Next Page\n", classStr))
		}
		request.Gemini("\n")
	})

	s.AddRoute("/search/scrollspace/:class/:page", func(request sis.Request) {
		classStr := request.GetParam("class")
		if classStr != "0" && classStr != "1" && classStr != "2" && classStr != "3" && classStr != "4" && classStr != "5" && classStr != "6" && classStr != "7" && classStr != "8" && classStr != "9" {
			request.BadRequest("UDC class doesn't exist.")
			return
		}
		pageStr := request.GetParam("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			request.BadRequest("Couldn't parse int.")
			return
		}

		pages, hasNextPage, hasPrevPage := getPagesOfUDC(conn, page, classStr)

		request.Gemini(fmt.Sprintf("# Scrollspace Index: %s. %s\n", classStr, UdcClassStringToShortTitle(classStr)))
		request.Gemini("\n=> /search/scrollspace/ Scrollspace Index Home\n\n")

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)
		request.Gemini(builder.String() + "\n")

		if hasPrevPage {
			request.Gemini(fmt.Sprintf("\n=> /search/scrollspace/%s/%d/ Previous Page\n", classStr, page-1))
		}
		if hasNextPage {
			request.Gemini(fmt.Sprintf("\n=> /search/scrollspace/%s/%d/ Next Page\n", classStr, page+1))
		}
		request.Gemini("\n")
	})

	var refreshCacheEvery = time.Hour * 1
	var pagesCountCache = 0
	var lastCrawlCache = time.Time{}
	var totalSizeCache float64 = -1
	var totalSizeTextCache float64 = -1
	var lastCacheTime time.Time
	s.AddRoute("/search/stats", func(request sis.Request) {
		currentTime := time.Now()
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: currentTime, Abstract: "# AuraGem Search Stats\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		if totalSizeCache == -1 || lastCacheTime.Add(refreshCacheEvery).Before(currentTime) {
			row := conn.QueryRowContext(context.Background(), "SELECT COUNT(*), MAX(LAST_SUCCESSFUL_VISIT), SUM(SIZE) FROM pages WHERE SCHEME = 'gemini' OR SCHEME = 'GEMINI' AND HIDDEN=false")
			row.Scan(&pagesCountCache, &lastCrawlCache, &totalSizeCache)
			// Convert totalSize to GB
			lastCacheTime = currentTime
		}

		totalSize := totalSizeCache
		totalSize /= 1024 // Bytes to KB
		totalSize /= 1024 // KB to MB
		totalSize /= 1024 // MB to GB

		row2 := conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM domains")
		domainsCount := 0
		row2.Scan(&domainsCount)

		row3 := conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM pages WHERE FEED = true AND HIDDEN=false")
		feedCount := 0
		row3.Scan(&feedCount)

		if totalSizeTextCache == -1 || lastCacheTime.Add(refreshCacheEvery).Before(currentTime) {
			row4 := conn.QueryRowContext(context.Background(), "SELECT SUM(SIZE) FROM pages WHERE contenttype LIKE 'text/%%' AND (SCHEME = 'gemini' OR SCHEME = 'GEMINI') AND HIDDEN=false")
			row4.Scan(&totalSizeTextCache)
			lastCacheTime = currentTime
		}
		totalSizeText := totalSizeTextCache
		totalSizeText /= 1024 // Bytes to KB
		totalSizeText /= 1024 // KB to MB
		totalSizeText /= 1024 // MB to GB

		row5 := conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM domains WHERE slowdowncount<>0")
		var slowdowncount = 0
		row5.Scan(&slowdowncount)

		row6 := conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM domains WHERE emptymetacount<>0")
		var emptyMetaCount = 0
		row6.Scan(&emptyMetaCount)

		request.Gemini(fmt.Sprintf(`# AuraGem Search Stats

Index Updated on %s

Page Count: %d
Capsule Count: %d
Gemsub Feed Count: %d

Total Size of Geminispace: %.3f GB
Total Size of Text Files within Geminispace: %.3f GB (%.2f%% of Geminispace)

Number of Domains with SlowDown responses: %d
Number of Domains that responded with an empty META field: %d

=> /search/mimetype/ Mimetypes with Counts

`, lastCrawlCache.Format("2006-01-02"), pagesCountCache, domainsCount, feedCount, totalSize, totalSizeText, totalSizeText/totalSize*100.0, slowdowncount, emptyMetaCount))

		if globalData.IsCrawling() {
			request.Gemini(fmt.Sprintf("## Crawler\n\nCurrently crawled %d documents.\n", globalData.CrawledCount()))
		}
		if currentCapsuleCrawl != "" {
			request.Gemini(fmt.Sprintf("## Capsule On-Demand Crawler\n\nCurrently crawling capsule '%s'.\n", currentCapsuleCrawl))
		}
	})

	handleSearchFeedback(s)

	s.AddRoute("/search/add_capsule", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Capsule:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Add Capsule to Index\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			queryUrl, parse_err := url.Parse(query)
			if parse_err != nil {
				request.Redirect("/search/add_capsule")
				return
			}
			queryUrl.Fragment = "" // Strip the fragment
			if (queryUrl.Scheme != "gemini" && queryUrl.Scheme != "nex" && queryUrl.Scheme != "spartan" && queryUrl.Scheme != "scroll") || !queryUrl.IsAbs() {
				request.Redirect("/search/add_capsule")
				return
			}
			if queryUrl.Path == "" {
				queryUrl.Path = "/"
			}

			_, err := addSeedToDb(conn, Seed{0, queryUrl.String(), time.Time{}})
			if err != nil {
				request.TemporaryFailure("%s", err.Error())
				return //err
			} else {
				request.Gemini(fmt.Sprintf("# Capsule submitted\n\n'%s' has been submitted as a seed.\n=> /search/ AuraGem Search Home\n", queryUrl.String()))
				return
			}
		}
	})

	s.AddRoute("/search/backlinks", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Gemini URL:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Backlinks\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			// Check that gemini url in query string is correct
			queryUrl, parse_err := url.Parse(query)
			if parse_err != nil {
				request.Redirect("/search/backlinks")
				return
			}

			if queryUrl.Scheme != "gemini" || !queryUrl.IsAbs() {
				request.Redirect("/search/backlinks")
				return
			}

			if queryUrl.Path == "" {
				queryUrl.Path = "/"
			}

			handleBacklinks(request, conn, queryUrl)
			return
		}
	})

	// Smallnet search
	s.AddRoute("/search/s", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - '" + query + "'\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			// Page 1
			handleSearch(request, conn, query, 1, false, false, false, false)
			return
		}
	})

	s.AddRoute("/search/s/:page", func(request sis.Request) {
		pageStr := request.GetParam("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			request.BadRequest("Couldn't parse int.")
			return
		}

		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - '" + query + "' Page " + pageStr + "\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			handleSearch(request, conn, query, page, false, false, false, false)
			return
		}
	})

	// Geminispace search
	s.AddRoute("/search/gemini", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - '" + query + "'\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			// Page 1
			handleSearch(request, conn, query, 1, false, true, false, false)
			return
		}
	})

	s.AddRoute("/search/gemini/:page", func(request sis.Request) {
		pageStr := request.GetParam("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			request.BadRequest("Couldn't parse int.")
			return
		}

		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - '" + query + "' Page " + pageStr + "\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			handleSearch(request, conn, query, page, false, true, false, false)
			return
		}
	})

	// Scroll protocol search
	s.AddRoute("/search/scroll", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - '" + query + "'\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			// Page 1
			handleSearch(request, conn, query, 1, false, false, true, false)
			return
		}
	})

	s.AddRoute("/search/scroll/:page", func(request sis.Request) {
		pageStr := request.GetParam("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			request.BadRequest("Couldn't parse int.")
			return
		}

		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - '" + query + "' Page " + pageStr + "\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			handleSearch(request, conn, query, page, false, false, true, false)
			return
		}
	})

	// Spartanspace search
	s.AddRoute("/search/spartan", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - '" + query + "'\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			// Page 1
			handleSearch(request, conn, query, 1, false, false, false, true)
			return
		}
	})

	s.AddRoute("/search/spartan/:page", func(request sis.Request) {
		pageStr := request.GetParam("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			request.BadRequest("Couldn't parse int.")
			return
		}

		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - '" + query + "' Page " + pageStr + "\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			handleSearch(request, conn, query, page, false, false, false, true)
			return
		}
	})

	// Debug searching - shows the Score numbers
	s.AddRoute("/search/debug_s", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			// Page 1
			handleSearch(request, conn, query, 1, true, false, false, false)
			return
		}
	})

	s.AddRoute("/search/debug_s/:page", func(request sis.Request) {
		pageStr := request.GetParam("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			request.BadRequest("Couldn't parse int.")
			return
		}

		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Search Query:")
			return
		} else {
			handleSearch(request, conn, query, page, true, false, false, false)
			return
		}
	})

	s.AddRoute("/search/recent", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - 50 Most Recently Indexed\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		pages := getRecent(conn)

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)

		request.Gemini(fmt.Sprintf(`# 50 Most Recently Indexed

=> /search/ Home
=> /search/s/ Search

%s
`, builder.String()))
	})

	s.AddRoute("/search/capsules", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - List of Capsules\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		capsules := getCapsules(conn)

		var builder strings.Builder
		for _, capsule := range capsules {
			scheme := "gemini"
			if capsule.Port == 5699 {
				scheme = "scroll"
			} else if capsule.Port == 300 {
				scheme = "spartan"
			} else if capsule.Port == 1900 {
				scheme = "nex"
			}
			if capsule.Title == "" {
				fmt.Fprintf(&builder, "=> %s://%s %s\n", scheme, capsule.Domain, capsule.Domain)
			} else {
				fmt.Fprintf(&builder, "=> %s://%s %s\n", scheme, capsule.Domain, capsule.Title)
			}
		}

		request.Gemini(fmt.Sprintf(`# List of Capsules

=> /search/ Home
=> /search/s/ Search

%s
`, builder.String()))
	})

	s.AddRoute("/search/feeds", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Indexed Feeds\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		pages := getFeeds(conn)

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)

		request.Gemini(fmt.Sprintf(`# Indexed Feeds

=> /search/ Home
=> /search/s/ Search

%s
`, builder.String()))
	})

	/*
			s.AddRoute("/search/yearposts", func(request sis.Request) {
				request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Posts From The Past Year\n"})
				if request.ScrollMetadataRequested {
					request.SendAbstract("")
					return
				}

				page := 1
				results := 40
				skip := (page - 1) * results

				pages, totalResultsCount := getPagesWithPublishDateFromLastYear(conn, results, skip)

				resultsStart := skip + 1
				resultsEnd := Min(totalResultsCount, skip+results) // + 1 - 1
				hasNextPage := resultsEnd < totalResultsCount && totalResultsCount != 0
				hasPrevPage := resultsStart > results

				var builder strings.Builder
				buildPageResults(&builder, pages, false, false)

				if hasPrevPage {
					fmt.Fprintf(&builder, "\n=> /search/yearposts/%d Previous Page\n", page-1)
				}
				if hasNextPage && !hasPrevPage {
					fmt.Fprintf(&builder, "\n=> /search/yearposts/%d/ Next Page\n", page+1)
				} else if hasNextPage && hasPrevPage {
					fmt.Fprintf(&builder, "=> /search/yearposts/%d/ Next Page\n", page+1)
				}

				request.Gemini(fmt.Sprintf(`# Posts From The Past Year

		=> /search/ Home
		=> /search/s/ Search

		Note: Currently tries to list only posts that are in English.

		%s
		`, builder.String()))
			})

			s.AddRoute("/search/yearposts/:page", func(request sis.Request) {
				pageStr := request.GetParam("page")
				request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Posts From The Past Year, Page " + pageStr + "\n"})
				if request.ScrollMetadataRequested {
					request.SendAbstract("")
					return
				}

				page, err := strconv.Atoi(pageStr)
				if err != nil {
					request.BadRequest("Couldn't parse int.")
					return
				}

				results := 40
				skip := (page - 1) * results

				pages, totalResultsCount := getPagesWithPublishDateFromLastYear(conn, results, skip)

				resultsStart := skip + 1
				resultsEnd := Min(totalResultsCount, skip+results) // + 1 - 1
				hasNextPage := resultsEnd < totalResultsCount && totalResultsCount != 0
				hasPrevPage := resultsStart > results

				var builder strings.Builder
				buildPageResults(&builder, pages, false, false)

				if hasPrevPage {
					fmt.Fprintf(&builder, "\n=> /search/yearposts/%d Previous Page\n", page-1)
				}
				if hasNextPage && !hasPrevPage {
					fmt.Fprintf(&builder, "\n=> /search/yearposts/%d/ Next Page\n", page+1)
				} else if hasNextPage && hasPrevPage {
					fmt.Fprintf(&builder, "=> /search/yearposts/%d/ Next Page\n", page+1)
				}

				request.Gemini(fmt.Sprintf(`# Posts From The Past Year

		=> /search/ Home
		=> /search/s/ Search

		Note: Currently tries to list only posts that are in English.

		%s
		`, builder.String()))
			})
	*/

	s.AddRoute("/search/audio", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Indexed Audio Files\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		pages, _, _ := getAudioFiles(conn, 1)

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)
		/*for _, page := range pages {
			artist := ""
			if page.AlbumArtist != "" {
				artist = "(" + page.AlbumArtist + ")"
			} else if page.Artist != "" {
				artist = "(" + page.Artist + ")"
			}
			if page.Title == "" {
				fmt.Fprintf(&builder, "=> %s %s %s\n", page.Url, page.Url, artist)
			} else {
				fmt.Fprintf(&builder, "=> %s %s %s\n", page.Url, page.Title, artist)
			}
		}*/

		request.Gemini(fmt.Sprintf(`# Indexed Audio Files

=> /search/ Home

%s

=> /search/audio/2/ Next Page
`, builder.String()))
	})

	s.AddRoute("/search/audio/:page", func(request sis.Request) {
		pageStr := request.GetParam("page")
		page_int, parse_err := strconv.ParseInt(pageStr, 10, 64)
		if parse_err != nil {
			request.BadRequest("Page Number Error")
			return
		}
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Indexed Audio Files, Page " + pageStr + "\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		pages, _, hasNextPage := getAudioFiles(conn, page_int)

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)

		// Handle pagination
		fmt.Fprintf(&builder, "\n")
		if page_int > 1 {
			fmt.Fprintf(&builder, "=> /search/audio/%d/ Prev Page\n", page_int-1)
		}
		if hasNextPage {
			fmt.Fprintf(&builder, "=> /search/audio/%d/ Next Page\n", page_int+1)
		}

		request.Gemini(fmt.Sprintf(`# Indexed Audio Files

=> /search/ Home

%s
`, builder.String()))
	})

	s.AddRoute("/search/audio/s", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Audio Search Query:")
			return
		} else {
			// Page 1
			handleAudioSearch(request, conn, query, 1)
			return
		}
	})
	s.AddRoute("/search/audio/s/:page", func(request sis.Request) {
		pageStr := request.GetParam("page")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			request.BadRequest("Couldn't parse int.")
			return
		}

		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Audio Search Query:")
			return
		} else {
			// Page 1
			handleAudioSearch(request, conn, query, page)
			return
		}
	})

	s.AddRoute("/search/images", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Indexed Image Files\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		pages, _, _ := getImageFiles(conn, 1)

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)
		/*for _, page := range pages {
			artist := ""
			if page.AlbumArtist != "" {
				artist = "(" + page.AlbumArtist + ")"
			} else if page.Artist != "" {
				artist = "(" + page.Artist + ")"
			}
			if page.Title == "" {
				fmt.Fprintf(&builder, "=> %s %s %s\n", page.Url, page.Url, artist)
			} else {
				fmt.Fprintf(&builder, "=> %s %s %s\n", page.Url, page.Title, artist)
			}
		}*/

		request.Gemini(fmt.Sprintf(`# Indexed Image Files

=> /search/ Home
=> /search/s/ Search

%s

=> /search/images/2/ Next Page
`, builder.String()))
	})

	s.AddRoute("/search/images/:page", func(request sis.Request) {
		pageStr := request.GetParam("page")
		page_int, parse_err := strconv.ParseInt(pageStr, 10, 64)
		if parse_err != nil {
			request.BadRequest("Page Number Error")
			return
		}
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Indexed Image Files, Page " + pageStr + "\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		pages, _, hasNextPage := getImageFiles(conn, page_int)
		if len(pages) == 0 {
			request.NotFound("Page not found.")
			return
		}

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)

		// Handle pagination
		fmt.Fprintf(&builder, "\n")
		if page_int > 1 {
			fmt.Fprintf(&builder, "=> /search/images/%d/ Prev Page\n", page_int-1)
		}
		if hasNextPage {
			fmt.Fprintf(&builder, "=> /search/images/%d/ Next Page\n", page_int+1)
		}

		request.Gemini(fmt.Sprintf(`# Indexed Image Files

=> /search/ Home
=> /search/s/ Search

%s
`, builder.String()))
	})

	s.AddRoute("/search/twtxt", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Indexed Twtxt Files\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		pages := getTwtxtFiles(conn)
		if len(pages) == 0 {
			request.NotFound("Page not found.")
			return
		}

		var builder strings.Builder
		buildPageResults(&builder, pages, false, false)
		/*for _, page := range pages {
			if page.Title == "" {
				fmt.Fprintf(&builder, "=> %s %s\n", page.Url, page.Url)
			} else {
				fmt.Fprintf(&builder, "=> %s %s\n", page.Url, page.Title)
			}
		}*/

		request.Gemini(fmt.Sprintf(`# Indexed Twtxt Files

=> /search/ Home
=> /search/s Search

%s
`, builder.String()))
	})

	s.AddRoute("/search/security", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Indexed Security.txt Files\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		pages := getSecurityTxtFiles(conn)
		if len(pages) == 0 {
			request.NotFound("Page not found.")
			return
		}

		var builder strings.Builder
		for _, page := range pages {
			if page.domain.Title != "" {
				fmt.Fprintf(&builder, "=> %s Security.txt for '%s'\n", page.page.Url, page.domain.Title)
			} else {
				fmt.Fprintf(&builder, "=> %s Security.txt for '%s'\n", page.page.Url, page.domain.Domain)
			}
		}

		request.Gemini(fmt.Sprintf(`# Indexed Security.txt Files

=> /search/ Home
=> /search/s/ Search

%s
`, builder.String()))
	})

	s.AddRoute("/search/mimetype", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Mimetypes\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			mimetypesList := getMimetypes(conn)
			var mimetypes strings.Builder
			for _, item := range mimetypesList {
				fmt.Fprintf(&mimetypes, "=> /search/s/?%s %s (%d)\n", url.QueryEscape("CONTENTTYPE:("+item.mimetype+")"), item.mimetype, item.count)
			}

			request.Gemini(fmt.Sprintf(`# Mimetypes

=> /search/ Home
=> /search/s/ Search

%s
`, mimetypes.String()))
		} else {
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# AuraGem Search - Indexed of Mimetype '" + query + "'\n"})
			if request.ScrollMetadataRequested {
				request.SendAbstract("")
				return
			}

			pages := getMimetypeFiles(conn, query)
			if len(pages) == 0 {
				request.NotFound("Page not found.")
				return
			}

			var builder strings.Builder
			buildPageResults(&builder, pages, false, false)

			request.Gemini(fmt.Sprintf(`# Indexed of Mimetype '%s'

=> /search/ Home
=> /search/s/ Search
=> /search/mimetype/ Mimetypes

%s
`, query, builder.String()))
		}
	})

	s.AddRoute("/search/random", func(request sis.Request) {
		q := `SELECT FIRST 1 'gemini://' || r.DOMAIN || '/'
FROM DOMAINS r
ORDER BY (r.ID + cast(? as bigint))*4294967291-((r.ID + cast(? as bigint))*4294967291/49157)*49157`
		time := time.Now().Unix()
		row := conn.QueryRowContext(context.Background(), q, time, time)
		var page Page
		scan_err := row.Scan(&page.Url)
		if scan_err == nil {
			request.Redirect("%s", page.Url)
			return
		} else if scan_err == sql.ErrNoRows {
			request.Redirect("/search/")
			return
		} else {
			panic(scan_err)
		}
	})
}

func handleBacklinks(request sis.Request, conn *sql.DB, url *url.URL) {
	q := `SELECT COUNT(*) OVER () totalCount, r.ID, P_FROM.ID, P_FROM.URL, r.TITLE, r.CROSSHOST, r.CRAWLINDEX,
    r.DATE_ADDED
FROM LINKS r
JOIN pages P_TO ON r.PAGEID_TO = P_TO.ID
JOIN Pages P_FROM ON r.PAGEID_FROM = P_FROM.ID
WHERE P_TO.URL=?
ORDER BY r.CROSSHOST ASC`
	rows, rows_err := conn.QueryContext(context.Background(), q, url.String())
	var backlinks []Backlink = make([]Backlink, 0)
	var totalResultsCount = 0 // Total count of all results, regardless of pagination
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var backlink Backlink
			scan_err := rows.Scan(&totalResultsCount, &backlink.Id, &backlink.PageId_From, &backlink.PageURL_FROM, &backlink.Title, &backlink.Crosshost, &backlink.CrawlIndex, &backlink.Date_added)
			if scan_err == nil {
				backlinks = append(backlinks, backlink)
			} else {
				panic(scan_err)
			}
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}
	} else {
		panic(rows_err)
	}

	var builder strings.Builder
	for _, backlink := range backlinks {
		fmt.Fprintf(&builder, "=> %s \"%s\" with link \"%s\"\n", backlink.PageURL_FROM, backlink.PageURL_FROM, backlink.Title)
	}

	request.Gemini(fmt.Sprintf(`# Backlinks for %s

=> /search/ Home

%s
`, url.String(), builder.String()))
}

func handleSearch(request sis.Request, conn *sql.DB, query string, page int, showScores bool, gemini_only bool, scroll_only bool, spartan_only bool) {
	//rawQuery := c.URL().RawQuery
	rawQuery, err := request.RawQuery()
	if err != nil {
		request.TemporaryFailure("%s", err.Error())
		return
	}
	results := 30
	skip := (page - 1) * results

	// Escape single quotes ('test' => '''test''')
	queryFiltered := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(query, "\n", " "), "\r", ""), "'", "''")
	queryFiltered = strings.Replace(queryFiltered, "wikipedia", "gemipedia^2 wikipedia", 1)
	queryFiltered = strings.Replace(queryFiltered, "Wikipedia", "gemipedia^2 Wikipedia", 1)
	queryFiltered = strings.Replace(queryFiltered, "project gemini", "\"project gemini\"", 1)
	queryFiltered = strings.Replace(queryFiltered, "Project Gemini", "\"Project Gemini\"", 1)
	queryFiltered = strings.Replace(queryFiltered, "Project Gemini", "\"Project Gemini\"", 1)
	queryFiltered = strings.Replace(queryFiltered, "project Gemini", "\"project Gemini\"", 1)
	queryFiltered = strings.Replace(queryFiltered, "Project gemini", "\"Project gemini\"", 1)
	//queryFiltered = strings.Replace(queryFiltered, "gemini", "\"gemini protocol\"", 1) // TODO: Doesn't work well yet

	actualQuery := ""
	if !gemini_only && !scroll_only && !spartan_only {
		actualQuery = strings.Replace(fts_searchQuery, `%%query%%`, queryFiltered, 2) // TODO: Support for protocol-specific searching.
	} else if gemini_only {
		actualQuery = strings.Replace(fts_searchQuery_protocol, `%%query%%`, queryFiltered, 2) // TODO: Support for protocol-specific searching.
		actualQuery = strings.Replace(actualQuery, `%%protocol%%`, "gemini", 1)
	} else if scroll_only {
		actualQuery = strings.Replace(fts_searchQuery_protocol, `%%query%%`, queryFiltered, 2) // TODO: Support for protocol-specific searching.
		actualQuery = strings.Replace(actualQuery, `%%protocol%%`, "scroll", 1)
	} else if spartan_only {
		actualQuery = strings.Replace(fts_searchQuery_protocol, `%%query%%`, queryFiltered, 2) // TODO: Support for protocol-specific searching.
		actualQuery = strings.Replace(actualQuery, `%%protocol%%`, "spartan", 1)
	}
	actualQuery = strings.Replace(actualQuery, `%%first%%`, strconv.Itoa(results), 1)
	actualQuery = strings.Replace(actualQuery, `%%skip%%`, strconv.Itoa(skip), 1)

	parts := strings.Split(queryFiltered, " ")
	var matchesBuilder strings.Builder
	fmt.Fprintf(&matchesBuilder, "t.NAME = '%s' ", queryFiltered)
	for _, part := range parts {
		if part == "" {
			continue
		}
		fmt.Fprintf(&matchesBuilder, "OR t.NAME = '%s' ", part)
	}

	actualQuery = strings.Replace(actualQuery, `%%matches%%`, matchesBuilder.String(), 1)
	//q := `SELECT id, url, urlhash, scheme, domainid, contenttype, charset, language, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, hidden FROM pages WHERE lower(url) LIKE lower(?) OR lower(title) LIKE lower(?) OR lower(artist) LIKE lower(?) OR lower(album) LIKE lower(?) OR lower(albumartist) LIKE lower(?) OR id IN (SELECT keywords.pageid FROM keywords where lower(keywords.keyword) LIKE ?)`

	//fmt.Printf("Query: %s", queryBuilder.String())

	fmt.Printf("%s\n", actualQuery)

	before := time.Now()
	rows, rows_err := conn.QueryContext(context.Background(), actualQuery)
	after := time.Now()
	timeTaken := after.Sub(before)
	fmt.Printf("Time taken: %v\n", timeTaken)

	var pages []Page = make([]Page, 0, results)
	var totalResultsCount = 0 // Total count of all results, regardless of pagination
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&totalResultsCount, &page.Score, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
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
		panic(rows_err)
	}

	resultsStart := skip + 1
	resultsEnd := Min(totalResultsCount, skip+results) // + 1 - 1
	hasNextPage := resultsEnd < totalResultsCount && totalResultsCount != 0
	hasPrevPage := resultsStart > results

	request.Gemini(fmt.Sprintf("# AuraGem Search - Results %d-%d/%d\n", resultsStart, resultsEnd, totalResultsCount))
	request.Gemini("\n=> /search/ Home\n")
	request.PromptLine("/search/s/", "New Search")

	var builder strings.Builder

	// Try to get wikipedia suggestion
	if page == 1 {
		results, _, err := wiki.Search(query, 3, true)
		if err == nil && len(results) > 0 {
			for _, result := range results {
				builder.WriteString(fmt.Sprintf("=> gemini://gemi.dev/cgi-bin/wp.cgi/view?%s Gemipedia: %s\n", url.QueryEscape(result), result))
			}
		}
		builder.WriteString("\n")
	}

	buildPageResults(&builder, pages, false, showScores)

	request.Gemini(fmt.Sprintf("\nQuery: '%s'\nTime Taken: %v\n\n%s\n", query, timeTaken, builder.String()))

	if hasPrevPage {
		request.Gemini(fmt.Sprintf("\n=> /search/s/%d/?%s Previous Page\n", page-1, rawQuery))
	}
	if hasNextPage && !hasPrevPage {
		request.Gemini(fmt.Sprintf("\n=> /search/s/%d/?%s Next Page\n", page+1, rawQuery))
	} else if hasNextPage && hasPrevPage {
		request.Gemini(fmt.Sprintf("=> /search/s/%d/?%s Next Page\n", page+1, rawQuery))
	}

	request.Gemini("\nNote that AuraGem Search does not ensure or rank based on the popularity or accuracy of the information within any of the pages listed in these search results. One cannot presume that information published within Geminispace is or is not for ill-intent or misinformation, even if it's popular or well-linked, so one must use their best judgement in determining the trustworthiness of such content themselves.\n")
}

func handleSearchIndex(request sis.Request, conn *sql.DB) {
	request.Gemini("Test\n")
	query := "SELECT FIRST %%first%% SKIP %%skip%% COUNT(*) OVER () totalCount, P.ID, P.URL, P.SCHEME, P.DOMAINID, P.CONTENTTYPE, P.CHARSET, P.LANGUAGE, P.LINECOUNT, P.UDC, P.TITLE, P.PROMPT, P.SIZE, P.HASH, P.FEED, P.PUBLISHDATE, P.INDEXTIME, P.ALBUM, P.ARTIST, P.ALBUMARTIST, P.COMPOSER, P.TRACK, P.DISC, P.COPYRIGHT, P.CRAWLINDEX, P.DATE_ADDED, P.LAST_SUCCESSFUL_VISIT, P.HIDDEN FROM PAGES P"
	results_per_query := 10
	current_query_index := 1
	max_results := 100000000 // TODO
	first := true

	//for {
	current_skip := (current_query_index - 1) * results_per_query
	if current_skip >= max_results {
		//break
	}

	actualQuery := strings.Replace(query, `%%first%%`, strconv.Itoa(results_per_query), 1)
	actualQuery = strings.Replace(actualQuery, `%%skip%%`, strconv.Itoa(current_skip), 1)
	fmt.Printf("Query: %s\n", actualQuery)

	rows, rows_err := conn.QueryContext(context.Background(), actualQuery)
	var pages []Page = make([]Page, 0, results_per_query)
	var totalResultsCount = 0 // Total count of all results, regardless of pagination
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&totalResultsCount, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
			if scan_err == nil {
				pages = append(pages, page)
			} else {
				panic(scan_err)
			}
		}
		if err := rows.Err(); err != nil {
			panic(err)
		}

		max_results = totalResultsCount
	} else {
		panic(rows_err)
		//break
	}

	var builder strings.Builder
	buildPageResults(&builder, pages, false, false)
	if first {
		request.Gemini(fmt.Sprintf("# AuraGem Search Full Index (%d Pages)\n\n", max_results))
		first = false
	}
	request.Gemini(fmt.Sprintf("%s\n", builder.String()))

	current_query_index += 1
	//}

	if first {
		request.TemporaryFailure("Error")
		return
	}
}

func handleAudioSearch(request sis.Request, conn *sql.DB, query string, page int) {
	rawQuery, err := request.RawQuery()
	if err != nil {
		request.TemporaryFailure("%s", err.Error())
		return
	}
	results := 30
	skip := (page - 1) * results

	queryFiltered := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(query, "\n", " "), "\r", ""), "'", "''")
	actualQuery := strings.Replace(fts_audioSearchQuery, `%%query%%`, queryFiltered, 2)
	actualQuery = strings.Replace(actualQuery, `%%first%%`, strconv.Itoa(results), 1)
	actualQuery = strings.Replace(actualQuery, `%%skip%%`, strconv.Itoa(skip), 1)

	before := time.Now()
	rows, rows_err := conn.QueryContext(context.Background(), actualQuery)
	after := time.Now()
	timeTaken := after.Sub(before)
	fmt.Printf("Time taken for audio search: %v\n", timeTaken)

	var pages []Page = make([]Page, 0, results)
	var totalResultsCount = 0 // Total count of all results, regardless of pagination
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&totalResultsCount, &page.Score, &page.Highlight, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Udc, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
			if scan_err == nil {
				pages = append(pages, page)
			} else {
				panic(scan_err)
			}
		}

		if err := rows.Err(); err != nil {
			panic(err)
		}
	} else {
		panic(rows_err)
	}

	resultsStart := skip + 1
	resultsEnd := Min(totalResultsCount, skip+results) // + 1 - 1
	hasNextPage := resultsEnd < totalResultsCount && totalResultsCount != 0
	hasPrevPage := resultsStart > results

	var builder strings.Builder
	buildPageResults(&builder, pages, true, false)

	if hasPrevPage {
		fmt.Fprintf(&builder, "\n=> /search/audio/s/%d/?%s Previous Page\n", page-1, rawQuery)
	}
	if hasNextPage && !hasPrevPage {
		fmt.Fprintf(&builder, "\n=> /search/audio/s/%d/?%s Next Page\n", page+1, rawQuery)
	} else if hasNextPage && hasPrevPage {
		fmt.Fprintf(&builder, "=> /search/audio/s/%d/?%s Next Page\n", page+1, rawQuery)
	}

	request.Gemini(fmt.Sprintf(`# AuraGem Audio Search - Results %d-%d/%d

=> /search/ Home
=> /search/audio/s/ New Audio Search

Query: '%s'
Time Taken: %v

%s
`, resultsStart, resultsEnd, totalResultsCount, query, timeTaken, builder.String()))
}

// TODO: Yiddish
var Esperanto language.Tag = language.MustParse("eo")
var Yiddish language.Tag = language.MustParse("yi")
var AustralianEnglish language.Tag = language.MustParse("en-AU")
var languageMatcher = language.NewMatcher([]language.Tag{
	language.English, // The first language is used as fallback.
	AustralianEnglish,
	language.BritishEnglish,
	language.AmericanEnglish,
	language.CanadianFrench,
	language.French,
	language.German,
	language.Dutch,
	Esperanto,
	language.LatinAmericanSpanish,
	language.EuropeanSpanish,
	language.Spanish,
	language.Danish,
	language.TraditionalChinese,
	language.SimplifiedChinese,
	language.Chinese,
	language.ModernStandardArabic,
	language.Arabic,
	language.Finnish,
	language.Ukrainian,
	language.Hebrew,
	language.Italian,
	language.BrazilianPortuguese,
	language.EuropeanPortuguese,
	language.Portuguese,
	language.Russian,
	language.Greek,
	language.Hindi,
	language.Korean,
	language.Persian,
	Yiddish, // Yiddish
	language.Italian,
})

func langTagToText(tag language.Tag) string {
	switch tag {
	case language.English:
		return "English"
	case language.BritishEnglish:
		return "English"
	case language.AmericanEnglish:
		return "English"
	case AustralianEnglish:
		return "English"
	case language.CanadianFrench:
		return "French"
	case language.French:
		return "French"
	case language.German:
		return "German"
	case language.Dutch:
		return "Dutch"
	case Esperanto:
		return "Esperanto"
	case language.LatinAmericanSpanish:
		return "Spanish"
	case language.EuropeanSpanish:
		return "Spanish"
	case language.Spanish:
		return "Spanish"
	case language.Danish:
		return "Danish"
	case language.TraditionalChinese:
		return "Chinese"
	case language.SimplifiedChinese:
		return "Chinese"
	case language.Chinese:
		return "Chinese"
	case language.ModernStandardArabic:
		return "Arabic"
	case language.Arabic:
		return "Arabic"
	case language.Finnish:
		return "Finnish"
	case language.Ukrainian:
		return "Ukrainian"
	case language.Hebrew:
		return "Hebrew"
	case language.Italian:
		return "Italian"
	case language.BrazilianPortuguese:
		return "Portuguese"
	case language.EuropeanPortuguese:
		return "Portuguese"
	case language.Portuguese:
		return "Portuguese"
	case language.Russian:
		return "Russian"
	case language.Greek:
		return "Greek"
	case language.Hindi:
		return "Hindi"
	case language.Korean:
		return "Korean"
	case language.Persian:
		return "Persian"
	case Yiddish:
		return "Yiddish"
	case language.Italian:
		return "Italian"
	}

	return ""
}

func buildPageResults(builder *strings.Builder, pages []Page, useHighlight bool, showScores bool) {
	for _, page := range pages {
		typeText := ""
		if page.Prompt != "" {
			typeText = "Input Prompt • "
		} else if page.Feed {
			typeText = "Gemsub Feed • "
		}

		publishDateString := ""
		if page.PublishDate.Year() > 1800 && page.PublishDate.Year() <= time.Now().Year() {
			publishDateString = fmt.Sprintf("Published on %s • ", page.PublishDate.Format("2006-01-02"))
		}

		artist := ""
		if page.AlbumArtist != "" {
			artist = "by " + page.AlbumArtist + " • "
		} else if page.Artist != "" {
			artist = "by " + page.Artist + " • "
		}

		langText := ""
		if page.Content_type == "text/gemini" || page.Content_type == "" || strings.HasPrefix(page.Content_type, "text/") {
			// NOTE: This will just get the first language listed. In the future, list all languages by splitting on commas
			tag, _ := language.MatchStrings(languageMatcher, page.Language)
			str := langTagToText(tag)
			if str != "" {
				langText = fmt.Sprintf("%s • ", str)
			}
		}

		size := float64(page.Size)
		sizeLabel := "B"
		if size > 1024 {
			size /= 1024.0
			sizeLabel = "KB"
		}
		if size > 1024 {
			size /= 1024.0
			sizeLabel = "MB"
		}
		if size > 1024 {
			size /= 1024.0
			sizeLabel = "GB"
		}

		score := ""
		if showScores {
			score = fmt.Sprintf(" (Score: %f)", page.Score)
		}

		if page.Title == "" {
			fmt.Fprintf(builder, "=> %s %s%s\n", page.Url, page.Url, score)
			fmt.Fprintf(builder, "%s%s%s%s%d Lines • %.1f %s\n", typeText, publishDateString, langText, artist, page.Linecount, size, sizeLabel)
		} else {
			fmt.Fprintf(builder, "=> %s %s%s\n", page.Url, page.Title, score)
			fmt.Fprintf(builder, "%s%s%s%s%d Lines • %.1f %s • %s\n", typeText, publishDateString, langText, artist, page.Linecount, size, sizeLabel, page.Url)
		}
		if useHighlight {
			fmt.Fprintf(builder, "> %s\n", page.Highlight)
		}
		fmt.Fprintf(builder, "\n")
	}
}

// Max returns the larger of x or y.
func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

// Min returns the smaller of x or y.
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
