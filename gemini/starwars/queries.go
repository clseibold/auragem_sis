package starwars

import (
	"context"
	"database/sql"
	"time"
)

func GetMovies(conn *sql.DB, timeline bool) ([]Movie, time.Time) {
	lastUpdate := time.Time{}
	var movies []Movie
	var q string
	if timeline {
		q = `SELECT r.ID, r."NAME", r.TIMELINEDATE, r.PUBLICATIONDATE, r.PRODUCTIONCOMPANY, r.DISTRIBUTOR, r.DATE_ADDED FROM MOVIES r order by timelinedate asc`
	} else {
		q = `SELECT r.ID, r."NAME", r.TIMELINEDATE, r.PUBLICATIONDATE, r.PRODUCTIONCOMPANY, r.DISTRIBUTOR, r.DATE_ADDED FROM MOVIES r order by publicationdate asc`
	}
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var movie Movie
			scan_err := rows.Scan(&movie.Id, &movie.Name, &movie.TimelineDate, &movie.PublicationDate, &movie.ProductionCompany, &movie.Distributor, &movie.Date_added)
			if scan_err == nil {
				if lastUpdate.Before(movie.Date_added) {
					lastUpdate = movie.Date_added
				}
				movies = append(movies, movie)
			}
		}
	}

	return movies, lastUpdate
}

func GetShows(conn *sql.DB) ([]TVShow, time.Time) {
	lastUpdate := time.Time{}
	var shows []TVShow
	q := `SELECT r.ID, r."NAME", r.DATE_ADDED FROM TVSHOWS r`
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var show TVShow
			scan_err := rows.Scan(&show.Id, &show.Name, &show.Date_added)
			if scan_err == nil {
				if lastUpdate.Before(show.Date_added) {
					lastUpdate = show.Date_added
				}
				shows = append(shows, show)
			}
		}
	}

	return shows, lastUpdate
}

func GetComicSeries_Full(conn *sql.DB) ([]ComicSeries, time.Time) {
	lastUpdate := time.Time{}
	var series []ComicSeries
	q := `SELECT r.ID, r."NAME", r.MINISERIES, r.STARTYEAR, r.DATE_ADDED, (SELECT COUNT(*) FROM COMICTPBS WHERE COMICTPBS.COMICSERIESID=r.ID) as tpbcount, (SELECT COUNT(*) FROM COMICISSUES WHERE COMICISSUES.COMICSERIESID=r.ID AND COMICISSUES.ANNUAL=false) as issuecount FROM COMICSERIES r WHERE r.miniseries=false`
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var serie ComicSeries
			scan_err := rows.Scan(&serie.Id, &serie.Name, &serie.Miniseries, &serie.StartYear, &serie.Date_added, &serie.TPBCount, &serie.IssueCount)
			if scan_err == nil {
				if lastUpdate.Before(serie.Date_added) {
					lastUpdate = serie.Date_added
				}
				series = append(series, serie)
			}
		}
	}

	return series, lastUpdate
}

func GetComicCrossovers(conn *sql.DB, timeline bool) ([]ComicTPB, time.Time) {
	lastUpdate := time.Time{}
	var tpbs []ComicTPB
	var q string
	if timeline {
		q = `SELECT r.ID, r."NAME", r.CROSSOVER, r.TIMELINEDATE, r.PUBLICATIONDATE, r.DATE_ADDED FROM COMICTPBS r WHERE r.CROSSOVER=true order by timelinedate ASC`
	} else {
		q = `SELECT r.ID, r."NAME", r.CROSSOVER, r.TIMELINEDATE, r.PUBLICATIONDATE, r.DATE_ADDED FROM COMICTPBS r WHERE r.CROSSOVER=true order by publicationdate ASC`
	}
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var tpb ComicTPB
			scan_err := rows.Scan(&tpb.Id, &tpb.Name, &tpb.Crossover, &tpb.TimelineDate, &tpb.PublicationDate, &tpb.Date_added)
			if scan_err == nil {
				if lastUpdate.Before(tpb.Date_added) {
					lastUpdate = tpb.Date_added
				}
				tpbs = append(tpbs, tpb)
			} else {
				panic(scan_err)
			}
		}
	} else {
		panic(rows_err)
	}

	return tpbs, lastUpdate
}

func GetComicSeries_Miniseries(conn *sql.DB) ([]ComicSeries, time.Time) {
	lastUpdate := time.Time{}
	var series []ComicSeries
	q := `SELECT r.ID, r."NAME", r.MINISERIES, r.STARTYEAR, r.DATE_ADDED, (SELECT COUNT(*) FROM COMICTPBS WHERE COMICTPBS.COMICSERIESID=r.ID) as tpbcount, (SELECT COUNT(*) FROM COMICISSUES WHERE COMICISSUES.COMICSERIESID=r.ID AND COMICISSUES.ANNUAL=false) as issuecount FROM COMICSERIES r WHERE r.miniseries=true`
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var serie ComicSeries
			scan_err := rows.Scan(&serie.Id, &serie.Name, &serie.Miniseries, &serie.StartYear, &serie.Date_added, &serie.TPBCount, &serie.IssueCount)
			if scan_err == nil {
				if lastUpdate.Before(serie.Date_added) {
					lastUpdate = serie.Date_added
				}
				series = append(series, serie)
			}
		}
	}

	return series, lastUpdate
}

func GetTPBs(conn *sql.DB, timeline bool) ([]ComicTPB, time.Time) {
	lastUpdate := time.Time{}
	var tpbs []ComicTPB
	var q string
	if timeline {
		q = `SELECT r.ID, r.VOLUME, r."NAME", r.CROSSOVER, r.TIMELINEDATE, r.PUBLICATIONDATE, r.DATE_ADDED, comicseries.ID, comicseries."NAME", comicseries.MINISERIES, comicseries.STARTYEAR, comicseries.DATE_ADDED, (SELECT COUNT(*) FROM COMICISSUES WHERE COMICISSUES.COMICTPBID=r.ID) as issuecount FROM COMICTPBS r LEFT JOIN COMICSERIES ON COMICSERIES.ID=r.COMICSERIESID order by timelinedate ASC`
	} else {
		q = `SELECT r.ID, r.VOLUME, r."NAME", r.CROSSOVER, r.TIMELINEDATE, r.PUBLICATIONDATE, r.DATE_ADDED, comicseries.ID, comicseries."NAME", comicseries.MINISERIES, comicseries.STARTYEAR, comicseries.DATE_ADDED, (SELECT COUNT(*) FROM COMICISSUES WHERE COMICISSUES.COMICTPBID=r.ID) as issuecount FROM COMICTPBS r LEFT JOIN COMICSERIES ON COMICSERIES.ID=r.COMICSERIESID order by publicationdate ASC`
	}
	rows, rows_err := conn.QueryContext(context.Background(), q) // TODO: ComicSeriesId
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var tpb ComicTPB

			var volume interface{}
			var timelineDate interface{}
			var series comicSeriesNullable
			scan_err := rows.Scan(&tpb.Id, &volume, &tpb.Name, &tpb.Crossover, &timelineDate, &tpb.PublicationDate, &tpb.Date_added, &series.Id, &series.Name, &series.Miniseries, &series.StartYear, &series.Date_added, &tpb.IssueCount)
			if scan_err == nil {
				if volume != nil {
					tpb.Volume = int(volume.(int32))
				}
				if timelineDate != nil {
					tpb.TimelineDate = int(timelineDate.(int32))
				}
				if lastUpdate.Before(tpb.Date_added) {
					lastUpdate = tpb.Date_added
				}
				tpb.Series = comicSeriesScan(series)

				tpbs = append(tpbs, tpb)
			} else {
				panic(scan_err)
			}
		}
	} else {
		panic(rows_err)
	}

	return tpbs, lastUpdate
}

func GetComicIssues(conn *sql.DB, timeline bool) ([]ComicIssue, time.Time) {
	lastUpdate := time.Time{}
	var issues []ComicIssue
	var q string
	if timeline {
		q = `SELECT r.ID, r."NUMBER", r."NAME", r.ANNUAL, r.TIMELINEDATE, r.PUBLICATIONDATE, r.PUBLISHER, r.DATE_ADDED, comicseries.ID, comicseries."NAME", comicseries.MINISERIES, comicseries.STARTYEAR, comicseries.DATE_ADDED FROM COMICISSUES r LEFT JOIN COMICSERIES ON COMICSERIES.ID=r.COMICSERIESID order by r.TIMELINEDATE ASC, r.NUMBER ASC`
	} else {
		q = `SELECT r.ID, r."NUMBER", r."NAME", r.ANNUAL,r.TIMELINEDATE, r.PUBLICATIONDATE, r.PUBLISHER, r.DATE_ADDED, comicseries.ID, comicseries."NAME", comicseries.MINISERIES, comicseries.STARTYEAR, comicseries.DATE_ADDED FROM COMICISSUES r LEFT JOIN COMICSERIES ON COMICSERIES.ID=r.COMICSERIESID order by r.PUBLICATIONDATE ASC, r.NUMBER ASC`
	}
	rows, rows_err := conn.QueryContext(context.Background(), q) // TODO: ComicSeriesId
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var issue ComicIssue

			//var volume interface{}
			var timelineDate interface{}
			var series comicSeriesNullable
			scan_err := rows.Scan(&issue.Id, &issue.Number, &issue.Name, &issue.Annual, &timelineDate, &issue.PublicationDate, &issue.Publisher, &issue.Date_added, &series.Id, &series.Name, &series.Miniseries, &series.StartYear, &series.Date_added)
			if scan_err == nil {
				/*if volume != nil {
					tpb.Volume = int(volume.(int32))
				}*/
				if timelineDate != nil {
					issue.TimelineDate = int(timelineDate.(int32))
				}
				issue.Series = comicSeriesScan(series)

				if lastUpdate.Before(issue.Date_added) {
					lastUpdate = issue.Date_added
				}
				issues = append(issues, issue)
			} else {
				panic(scan_err)
			}
		}
	} else {
		panic(rows_err)
	}

	return issues, lastUpdate
}

func GetComicOneshots(conn *sql.DB, timeline bool) ([]ComicIssue, time.Time) {
	lastUpdate := time.Time{}
	var issues []ComicIssue
	var q string
	if timeline {
		q = `SELECT r.ID, r."NUMBER", r."NAME", r.ANNUAL, r.TIMELINEDATE, r.PUBLICATIONDATE, r.PUBLISHER, r.DATE_ADDED FROM COMICISSUES r WHERE r.COMICSERIESID IS NULL order by r.TIMELINEDATE ASC, r.NUMBER ASC`
	} else {
		q = `SELECT r.ID, r."NUMBER", r."NAME", r.ANNUAL, r.TIMELINEDATE, r.PUBLICATIONDATE, r.PUBLISHER, r.DATE_ADDED FROM COMICISSUES r WHERE r.COMICSERIESID IS NULL order by r.PUBLICATIONDATE ASC, r.NUMBER ASC`
	}
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var issue ComicIssue
			var timelineDate interface{}
			scan_err := rows.Scan(&issue.Id, &issue.Number, &issue.Name, &issue.Annual, &timelineDate, &issue.PublicationDate, &issue.Publisher, &issue.Date_added)
			if scan_err == nil {
				/*if volume != nil {
					tpb.Volume = int(volume.(int32))
				}*/
				if timelineDate != nil {
					issue.TimelineDate = int(timelineDate.(int32))
				}

				if lastUpdate.Before(issue.Date_added) {
					lastUpdate = issue.Date_added
				}
				issues = append(issues, issue)
			} else {
				panic(scan_err)
			}
		}
	} else {
		panic(rows_err)
	}

	return issues, lastUpdate
}

func GetBookSeries(conn *sql.DB) ([]BookSeries, time.Time) {
	lastUpdate := time.Time{}
	var series []BookSeries
	q := `SELECT r.ID, r."NAME", r.DATE_ADDED, (SELECT COUNT(*) FROM BOOKS WHERE BOOKS.BOOKSERIESID=r.ID) as bookcount FROM BOOKSERIES r`
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var serie BookSeries
			scan_err := rows.Scan(&serie.Id, &serie.Name, &serie.Date_added, &serie.BookCount)
			if scan_err == nil {
				if lastUpdate.Before(serie.Date_added) {
					lastUpdate = serie.Date_added
				}
				series = append(series, serie)
			}
		}
	}

	return series, lastUpdate
}

func GetBooks(conn *sql.DB) ([]Book, time.Time) {
	lastUpdate := time.Time{}
	var books []Book
	q := `SELECT r.ID, r."NUMBER", r."NAME", r.BOOKTYPE, r.AUTHOR, r.BOOKSERIESID, r.TIMELINEDATE, r.PUBLICATIONDATE, r.PUBLISHER, r.DATE_ADDED, BOOKSERIES.ID, BOOKSERIES."NAME", BOOKSERIES.DATE_ADDED FROM BOOKS r LEFT JOIN BOOKSERIES ON BOOKSERIES.ID=r.BOOKSERIESID ORDER BY r.TIMELINEDATE ASC, r.NUMBER ASC`
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var bookNull bookNullable
			var bookSeriesNull bookSeriesNullable
			scan_err := rows.Scan(&bookNull.Id, &bookNull.Number, &bookNull.Name, &bookNull.BookType, &bookNull.Author, &bookNull.BookSeriesId, &bookNull.TimelineDate, &bookNull.PublicationDate, &bookNull.Publisher, &bookNull.Date_added, &bookSeriesNull.Id, &bookSeriesNull.Name, &bookSeriesNull.Date_added)
			if scan_err == nil {
				book := bookScan(bookNull)
				book.Series = bookSeriesScan(bookSeriesNull)
				if lastUpdate.Before(book.Date_added) {
					lastUpdate = book.Date_added
				}
				books = append(books, book)
			}
		}
	}

	return books, lastUpdate
}

func GetBookStandalones(conn *sql.DB) ([]Book, time.Time) {
	lastUpdate := time.Time{}
	var books []Book
	q := `SELECT r.ID, r."NUMBER", r."NAME", r.BOOKTYPE, r.AUTHOR, r.BOOKSERIESID, r.TIMELINEDATE, r.PUBLICATIONDATE, r.PUBLISHER, r.DATE_ADDED FROM BOOKS r WHERE r.BOOKSERIESID IS NULL ORDER BY r.TIMELINEDATE ASC, r.NUMBER ASC`
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var bookNull bookNullable
			scan_err := rows.Scan(&bookNull.Id, &bookNull.Number, &bookNull.Name, &bookNull.BookType, &bookNull.Author, &bookNull.BookSeriesId, &bookNull.TimelineDate, &bookNull.PublicationDate, &bookNull.Publisher, &bookNull.Date_added)
			if scan_err == nil {
				book := bookScan(bookNull)
				if lastUpdate.Before(book.Date_added) {
					lastUpdate = book.Date_added
				}
				books = append(books, book)
			}
		}
	}

	return books, lastUpdate
}
