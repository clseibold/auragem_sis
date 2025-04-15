package search

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/language"
)

// Aggregator tool queries the database to construct pages for the aggregator.
func Aggregate(destRoot string, conn *sql.DB) {
	page := 1
	for {
		hasNext := getPage(destRoot, page, conn)
		if !hasNext {
			break
		}
		page++
	}
}

func getPage(root string, page int, conn *sql.DB) bool {
	results := 40
	skip := (page - 1) * results

	pages, totalResultsCount := _getPagesWithPublishDateFromLastYear(conn, results, skip)

	resultsStart := skip + 1
	resultsEnd := Min(totalResultsCount, skip+results) // + 1 - 1
	hasNextPage := resultsEnd < totalResultsCount && totalResultsCount != 0
	hasPrevPage := resultsStart > results

	var builder strings.Builder
	_buildPageResults(&builder, pages, false, false)

	if hasPrevPage {
		if page-1 <= 1 {
			fmt.Fprintf(&builder, "\n=> /search/yearposts/ Previous Page\n")
		} else {
			fmt.Fprintf(&builder, "\n=> /search/yearposts/%d.gmi Previous Page\n", page-1)
		}
	}
	if hasNextPage && !hasPrevPage {
		fmt.Fprintf(&builder, "\n=> /search/yearposts/%d.gmi Next Page\n", page+1)
	} else if hasNextPage && hasPrevPage {
		fmt.Fprintf(&builder, "=> /search/yearposts/%d.gmi Next Page\n", page+1)
	}

	doc := fmt.Sprintf(`# Recent Publications

=> /search/ Home
=> /search/s/ Search

Note: Currently lists only English posts.

%s
`, builder.String())

	filename := filepath.Join(root, "index.gmi")
	if page > 1 {
		filename = filepath.Join(root, fmt.Sprintf("%d.gmi", page))
	}
	err := os.WriteFile(filename, []byte(doc), 0600)
	if err != nil {
		panic(err)
	}

	return hasNextPage
}

// TODO: Allow for different languages
// NOTE: Blank language fields are considered English
func _getPagesWithPublishDateFromLastYear(conn *sql.DB, results int, skip int) ([]Page, int) {
	query := fmt.Sprintf("SELECT FIRST %d SKIP %d COUNT(*) OVER () totalCount, id, url, scheme, domainid, contenttype, charset, language, linecount, title, prompt, size, hash, feed, publishdate, indextime, album, artist, albumartist, composer, track, disc, copyright, crawlindex, date_added, last_successful_visit, hidden FROM pages WHERE publishdate > dateadd(-1 year to ?) AND publishdate < dateadd(2 day to ?) AND (language = '' OR language LIKE 'en%%' OR language LIKE 'EN%%' OR language LIKE 'eng%%' OR language LIKE 'ENG%%') AND has_duplicate_on_gemini=false AND hidden = false AND domainid <> 9 ORDER BY publishdate DESC", results, skip)
	rows, rows_err := conn.QueryContext(context.Background(), query, time.Now().UTC(), time.Now().UTC())

	var pages []Page = make([]Page, 0, results)
	var totalCount int
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var page Page
			scan_err := rows.Scan(&totalCount, &page.Id, &page.Url, &page.Scheme, &page.DomainId, &page.Content_type, &page.Charset, &page.Language, &page.Linecount, &page.Title, &page.Prompt, &page.Size, &page.Hash, &page.Feed, &page.PublishDate, &page.Index_time, &page.Album, &page.Artist, &page.AlbumArtist, &page.Composer, &page.Track, &page.Disc, &page.Copyright, &page.CrawlIndex, &page.Date_added, &page.LastSuccessfulVisit, &page.Hidden)
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

func _buildPageResults(builder *strings.Builder, pages []Page, useHighlight bool, showScores bool) {
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
