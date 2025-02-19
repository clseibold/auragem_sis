package starwars

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gitlab.com/clseibold/auragem_sis/db"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

var publishDate, _ = time.ParseInLocation(time.RFC3339, "2024-03-19T08:23:00", time.Local)

func HandleStarWars(s sis.ServerHandle) {
	conn := db.NewConn(db.StarWarsDB)
	conn.SetMaxOpenConns(500)
	conn.SetMaxIdleConns(3)
	conn.SetConnMaxLifetime(time.Hour * 4)

	s.AddRoute("/starwars2", func(request *sis.Request) {
		request.Redirect("/starwars2/")
	})

	s.AddRoute("/starwars2/", func(request *sis.Request) {
		updateDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-19T08:23:00", time.Local)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Entertainment, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: updateDate, Language: "en", Abstract: "# Star Wars Database\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		request.Gemini(`# Star Wars Database

Welcome to the Star Wars Database.

## Canon - Ordered By Timeline
=> /starwars2/timeline/movies Movies
=> /starwars2/timeline/shows TV Shows
=> /starwars2/timeline/comics Comics
=> /starwars2/timeline/bookseries Books
=> /starwars2/timeline/all All

## Canon - Ordered By Publication
=> /starwars2/publication/movies Movies
=> /starwars2/publication/comics Comics

=> /starwars/ The Old Database
`)
	})

	// Movies
	s.AddRoute("/starwars2/timeline/movies", func(request *sis.Request) {
		handleMovies(request, conn, true)
	})
	s.AddRoute("/starwars2/publication/movies", func(request *sis.Request) {
		handleMovies(request, conn, false)
	})

	// Movies CSV
	s.AddRoute("/starwars2/timeline/movies/csv", func(request *sis.Request) {
		handleMoviesCSV(request, conn, true)
	})
	s.AddRoute("/starwars2/publication/movies/csv", func(request *sis.Request) {
		handleMoviesCSV(request, conn, false)
	})

	s.AddRoute("/starwars2/timeline/shows", func(request *sis.Request) {
		shows, lastUpdate := GetShows(conn)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Entertainment, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: TV Shows (Timeline)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		header, tableData := constructTableDataFromShows(shows)
		table := constructTable(header, tableData)

		var builder strings.Builder
		fmt.Fprintf(&builder, "```\n%s\n```\n\n", table)

		request.Gemini(fmt.Sprintf(`# Star Wars Shows

=> /starwars2/ Home
=> /starwars2/timeline/shows/episodes/ Episodes

%s
`, builder.String()))
	})

	s.AddRoute("/starwars2/timeline/comics", func(request *sis.Request) {
		fullSeries, lastUpdate := GetComicSeries_Full(conn)
		miniseries, lastUpdate2 := GetComicSeries_Miniseries(conn)
		if lastUpdate.Before(lastUpdate2) {
			lastUpdate = lastUpdate2
		}
		crossovers, lastUpdate2 := GetComicCrossovers(conn, true)
		if lastUpdate.Before(lastUpdate2) {
			lastUpdate = lastUpdate2
		}
		oneshots, lastUpdate2 := GetComicOneshots(conn, true)
		if lastUpdate.Before(lastUpdate2) {
			lastUpdate = lastUpdate2
		}

		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Literature, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Comics (Timeline)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		var builder strings.Builder
		fmt.Fprintf(&builder, "## Full Series\n")
		full_heading, full_data := constructTableDataFromSeries(fullSeries)
		full_table := constructTable(full_heading, full_data)
		fmt.Fprintf(&builder, "```\n%s```\n\n", full_table)

		fmt.Fprintf(&builder, "## Crossovers\n")
		crossovers_heading, crossovers_data := constructTableDataFromCrossover(crossovers)
		crossovers_table := constructTable(crossovers_heading, crossovers_data)
		fmt.Fprintf(&builder, "```\n%s```\n\n", crossovers_table)

		fmt.Fprintf(&builder, "## Miniseries\n")
		miniseries_heading, miniseries_data := constructTableDataFromSeries(miniseries)
		miniseries_table := constructTable(miniseries_heading, miniseries_data)
		fmt.Fprintf(&builder, "```\n%s```\n\n", miniseries_table)

		fmt.Fprintf(&builder, "## One-shots\n")
		oneshots_heading, oneshots_data := constructTableDataFromOneshots(oneshots)
		oneshots_table := constructTable(oneshots_heading, oneshots_data)
		fmt.Fprintf(&builder, "```\n%s```\n\n", oneshots_table)

		request.Gemini(fmt.Sprintf(`# Star Wars Comics - Series'

=> /starwars2/ Home
=> /starwars2/timeline/comics/issues Issues
=> /starwars2/timeline/comics/tpbs TPBs

%s
`, builder.String()))
	})

	s.AddRoute("/starwars2/publication/comics", func(request *sis.Request) {
		fullSeries, lastUpdate := GetComicSeries_Full(conn)
		miniseries, lastUpdate2 := GetComicSeries_Miniseries(conn)
		if lastUpdate.Before(lastUpdate2) {
			lastUpdate = lastUpdate2
		}
		crossovers, lastUpdate2 := GetComicCrossovers(conn, false)
		if lastUpdate.Before(lastUpdate2) {
			lastUpdate = lastUpdate2
		}
		oneshots, lastUpdate2 := GetComicOneshots(conn, false)
		if lastUpdate.Before(lastUpdate2) {
			lastUpdate = lastUpdate2
		}

		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Literature, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Comics (Publication)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		var builder strings.Builder
		fmt.Fprintf(&builder, "## Full Series\n")
		full_heading, full_data := constructTableDataFromSeries(fullSeries)
		full_table := constructTable(full_heading, full_data)
		fmt.Fprintf(&builder, "```\n%s```\n\n", full_table)

		fmt.Fprintf(&builder, "## Crossovers\n")
		crossovers_heading, crossovers_data := constructTableDataFromCrossover(crossovers)
		crossovers_table := constructTable(crossovers_heading, crossovers_data)
		fmt.Fprintf(&builder, "```\n%s```\n\n", crossovers_table)

		fmt.Fprintf(&builder, "## Miniseries\n")
		miniseries_heading, miniseries_data := constructTableDataFromSeries(miniseries)
		miniseries_table := constructTable(miniseries_heading, miniseries_data)
		fmt.Fprintf(&builder, "```\n%s```\n\n", miniseries_table)

		fmt.Fprintf(&builder, "## One-shots\n")
		oneshots_heading, oneshots_data := constructTableDataFromOneshots(oneshots)
		oneshots_table := constructTable(oneshots_heading, oneshots_data)
		fmt.Fprintf(&builder, "```\n%s```\n\n", oneshots_table)

		request.Gemini(fmt.Sprintf(`# Star Wars Comics - Series'

=> /starwars2/ Home
=> /starwars2/publication/comics/issues Issues
=> /starwars2/publication/comics/tpbs TPBs

%s
`, builder.String()))
	})

	s.AddRoute("/starwars2/timeline/comics/tpbs", func(request *sis.Request) {
		tpbs, lastUpdate := GetTPBs(conn, true)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Literature, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Comic TPBs (Timeline)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		heading, data := constructTableDataFromTPBs(tpbs)
		table := constructTable(heading, data)

		var builder strings.Builder
		fmt.Fprintf(&builder, "```\n%s```\n\n", table)

		request.Gemini(fmt.Sprintf(`# Star Wars Comics - TPBs

=> /starwars2/ Home
=> /starwars2/timeline/comics Comic Series'
=> /starwars2/timeline/comics/issues Issues
=> /starwars2/timeline/comics/tpbs TPBs

%s
`, builder.String()))
	})

	s.AddRoute("/starwars2/publication/comics/tpbs", func(request *sis.Request) {
		tpbs, lastUpdate := GetTPBs(conn, false)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Literature, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Comic TPBs (Publication)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		heading, data := constructTableDataFromTPBs(tpbs)
		table := constructTable(heading, data)

		var builder strings.Builder
		fmt.Fprintf(&builder, "```\n%s```\n\n", table)

		request.Gemini(fmt.Sprintf(`# Star Wars Comics - TPBs

=> /starwars2/ Home
=> /starwars2/publication/comics Comic Series'
=> /starwars2/publication/comics/issues Issues
=> /starwars2/publication/comics/tpbs TPBs

%s
`, builder.String()))
	})

	s.AddRoute("/starwars2/timeline/comics/issues", func(request *sis.Request) {
		issues, lastUpdate := GetComicIssues(conn, true)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Literature, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Comic Issues (Timeline)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		heading, data := constructTableDataFromIssues(issues)
		table := constructTable(heading, data)

		var builder strings.Builder
		fmt.Fprintf(&builder, "```\n%s```\n\n", table)

		request.Gemini(fmt.Sprintf(`# Star Wars Comics - Issues

=> /starwars2/ Home
=> /starwars2/timeline/comics Comic Series'
=> /starwars2/timeline/comics/issues Issues
=> /starwars2/timeline/comics/tpbs TPBs

%s
`, builder.String()))
	})

	s.AddRoute("/starwars2/publication/comics/issues", func(request *sis.Request) {
		issues, lastUpdate := GetComicIssues(conn, false)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Literature, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Comic Issues (Publication)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		heading, data := constructTableDataFromIssues(issues)
		table := constructTable(heading, data)

		var builder strings.Builder
		fmt.Fprintf(&builder, "```\n%s```\n\n", table)

		request.Gemini(fmt.Sprintf(`# Star Wars Comics - Issues

=> /starwars2/ Home
=> /starwars2/publication/comics Comic Series'
=> /starwars2/publication/comics/issues Issues
=> /starwars2/publication/comics/tpbs TPBs

%s
`, builder.String()))
	})

	s.AddRoute("/starwars2/timeline/bookseries", func(request *sis.Request) {
		var builder strings.Builder

		series, lastUpdate := GetBookSeries(conn)
		series_header, series_tableData := constructTableDataFromBookSeries(series)
		series_table := constructTable(series_header, series_tableData)
		fmt.Fprintf(&builder, "## Series'\n```\n%s\n```\n\n", series_table)

		standalones, lastUpdate2 := GetBookStandalones(conn)
		if lastUpdate.Before(lastUpdate2) {
			lastUpdate = lastUpdate2
		}
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Literature, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Book Series (Timeline)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		standalones_header, standalones_tableData := constructTableDataFromBookStandalones(standalones)
		standalones_table := constructTable(standalones_header, standalones_tableData)
		fmt.Fprintf(&builder, "## Standalones\n```\n%s\n```\n\n", standalones_table)

		request.Gemini(fmt.Sprintf(`# Star Wars Book Series'

=> /starwars2/ Home
=> /starwars2/timeline/bookseries Book Series'
=> /starwars2/timeline/books Books

%s
`, builder.String()))
	})

	s.AddRoute("/starwars2/timeline/books", func(request *sis.Request) {
		books, lastUpdate := GetBooks(conn)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Literature, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Books (Timeline)\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}

		header, tableData := constructTableDataFromBooks(books)
		table := constructTable(header, tableData)

		var builder strings.Builder
		fmt.Fprintf(&builder, "```\n%s\n```\n\n", table)

		request.Gemini(fmt.Sprintf(`# Star Wars Books

=> /starwars2/ Home
=> /starwars2/timeline/bookseries Book Series'
=> /starwars2/timeline/books Books

%s
`, builder.String()))
	})
}

func handleMovies(request *sis.Request, conn *sql.DB, timeline bool) {
	movies, lastUpdate := GetMovies(conn, timeline)
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Entertainment, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Movies\n"})
	if request.ScrollMetadataRequested() {
		request.SendAbstract("")
		return
	}

	header, tableData := constructTableDataFromMovies(movies)
	table := constructTable(header, tableData)

	var builder strings.Builder
	fmt.Fprintf(&builder, "```\n%s\n```\n", table)

	request.Gemini(fmt.Sprintf(`# Star Wars Movies

=> /starwars2/ Home

%s
=> movies/csv CSV File
`, builder.String()))
}

func handleMoviesCSV(request *sis.Request, conn *sql.DB, timeline bool) {
	movies, lastUpdate := GetMovies(conn, timeline)
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Entertainment, Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# Star Wars Database: Movies CSV\n"})
	if request.ScrollMetadataRequested() {
		request.SendAbstract("text/csv")
		return
	}

	header, tableData := constructTableDataFromMovies(movies)

	var builder strings.Builder
	for colNum, col := range header {
		fmt.Fprintf(&builder, "%s", col)
		if colNum < len(header)-1 {
			fmt.Fprintf(&builder, ",")
		}
	}
	fmt.Fprintf(&builder, "\n")

	for _, row := range tableData {
		for colNum, col := range row {
			fmt.Fprintf(&builder, "%s", col)
			if colNum < len(row)-1 {
				fmt.Fprintf(&builder, ",")
			}
		}
		fmt.Fprintf(&builder, "\n")
	}

	request.TextWithMimetype("text/csv", builder.String())
	//return c.Blob("text/csv", []byte(builder.String()))
}
