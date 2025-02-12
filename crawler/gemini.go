package crawler

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

type GeminiLink struct {
	name         string
	url          string
	spartanInput bool
}

func (ctx *CrawlContext) GetGeminiPageInfo2(dataReader *bytes.Reader, tagsMap *map[string]float64, mentionsMap *map[string]bool, links *[]GeminiLink, strippedTextBuilder *strings.Builder, update bool) (string, int, string, string, int, bool) {
	var isFeed = 0
	var spartanTitle = ""
	var lastTitleLevel = 5
	var linecount = 0
	size := dataReader.Len()
	var headingsBuilder strings.Builder
	var preformattedTextBuilder strings.Builder

	scanner := bufio.NewScanner(dataReader)
	inPreformat := false
	for scanner.Scan() {
		linecount += 1
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if spartanTitle == "" && strings.TrimSpace(line) != "" {
			if ContainsLetterRunes(line) && len(line) < 250 {
				// Assume for nex documents that the first non-blank line (that is under 250 bytes) is the title
				spartanTitle = strings.TrimSpace(line)
			}
		}
		if inPreformat {
			if strings.HasPrefix(line, "```") {
				inPreformat = false
			}
			fmt.Fprintf(strippedTextBuilder, "%s\n", line)
			fmt.Fprintf(&preformattedTextBuilder, "%s\n", line)
			continue
		}

		if strings.HasPrefix(line, "```") {
			inPreformat = !inPreformat
		} else if strings.HasPrefix(line, "####") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "####")))
			if spartanTitle == "" || lastTitleLevel > 4 {
				spartanTitle = strings.TrimSpace(strings.TrimPrefix(line, "####"))
				lastTitleLevel = 4
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "###") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "###")))
			if spartanTitle == "" || lastTitleLevel > 3 {
				spartanTitle = strings.TrimSpace(strings.TrimPrefix(line, "###"))
				lastTitleLevel = 3
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "##") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "##")))
			if spartanTitle == "" || lastTitleLevel > 2 {
				spartanTitle = strings.TrimSpace(strings.TrimPrefix(line, "##"))
				lastTitleLevel = 2
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "#") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "#")))
			if spartanTitle == "" || lastTitleLevel > 1 {
				spartanTitle = strings.TrimSpace(strings.TrimPrefix(line, "#"))
				lastTitleLevel = 1
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "=:") {
			// Input Link: Don't put in urls to crawl
			line = strings.TrimSpace(strings.TrimPrefix(line, "=:"))
			fmt.Fprintf(strippedTextBuilder, "%s\n", line)
			link, title, _ := CutAny(line, " \t")

			link_without_fragment, _, _ := strings.Cut(link, "#")
			//link_without_query_and_fragment, _, _ = strings.Cut(link_without_query_and_fragment, "?")
			*links = append(*links, GeminiLink{title, link_without_fragment, true})
		} else if strings.HasPrefix(line, "=>") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "=>"))
			fmt.Fprintf(strippedTextBuilder, "%s\n", line)
			link, title, _ := CutAny(line, " \t")

			link_without_fragment, _, _ := strings.Cut(link, "#")
			//link_without_query_and_fragment, _, _ = strings.Cut(link_without_query_and_fragment, "?")
			*links = append(*links, GeminiLink{title, link_without_fragment, false})

			if isTimeDate(title) {
				isFeed++
			}
		} else if strings.HasPrefix(line, ">") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimPrefix(line, ">"))
		} else if strings.HasPrefix(line, "**** ") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimPrefix(line, "**** "))
		} else if strings.HasPrefix(line, "*** ") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimPrefix(line, "*** "))
		} else if strings.HasPrefix(line, "** ") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimPrefix(line, "** "))
		} else if strings.HasPrefix(line, "* ") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimPrefix(line, "* "))
		} else {
			fmt.Fprintf(strippedTextBuilder, "%s\n", line)
			continue
		}
	}

	return spartanTitle, linecount, headingsBuilder.String(), preformattedTextBuilder.String(), size, isFeed > 1
}
