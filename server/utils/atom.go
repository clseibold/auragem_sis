package utils

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"
)

type AtomPost struct {
	link  string
	date  time.Time
	title string
}

func GenerateAtomFrom(file string, domain string, baseurl string, authorName string, authorEmail string) (string, string, time.Time) {
	//ISO8601Layout := "2006-01-02T15:04:05Z0700"
	feedTitle := ""
	var posts []AtomPost
	last_updated, _ := time.Parse("2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z")

	gemini, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	geminiLines := strings.Split(strings.TrimSuffix(string(gemini), "\n"), "\n")
	for _, line := range geminiLines {
		if strings.HasPrefix(line, "=> ") {
			parts := strings.SplitN(strings.Replace(line, "=> ", "", 1), " ", 3)

			t, err := time.Parse(time.RFC3339, parts[1])
			if err != nil {
				t, err = time.Parse("2006-01-02", parts[1])
				if err != nil {
					continue
				}
			}

			if t.After(last_updated) {
				last_updated = t
			}

			if len(parts) >= 3 {
				posts = append(posts, AtomPost{domain + parts[0], t, parts[2]})
			}
		} else if strings.HasPrefix(line, "# ") && feedTitle == "" {
			feedTitle = strings.Replace(line, "# ", "", 1)
		}
	}

	last_updated_string := last_updated.Format("2006-01-02T15:04:05Z")

	var builder strings.Builder
	fmt.Fprintf(&builder, `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
	<id>%s</id>
	<title>%s</title>
	<updated>%s</updated>
	<link href="%s"/>
	<author>
		<name>%s</name>
		<email>%s</email>
	</author>
`, baseurl, html.EscapeString(feedTitle), last_updated_string, baseurl+"/atom.xml", html.EscapeString(authorName), html.EscapeString(authorEmail))

	for _, post := range posts {
		post_date_string := post.date.Format(time.RFC3339)

		fmt.Fprintf(&builder,
			`	<entry>
		<title>%s</title>
		<link rel="alternate" href="%s"/>
		<id>%s</id>
		<updated>%s</updated>
	</entry>
`, html.EscapeString(post.title), post.link, post.link, post_date_string)
	}

	fmt.Fprintf(&builder, `</feed>`)

	return builder.String(), feedTitle, last_updated
}
