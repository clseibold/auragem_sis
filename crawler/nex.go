package crawler

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

type NexLink struct {
	name string
	url  string
}

func (ctx *CrawlContext) GetNexPageInfo(dataReader *bytes.Reader, tagsMap *map[string]float64, mentionsMap *map[string]bool, links *[]NexLink, strippedTextBuilder *strings.Builder, update bool) (string, int, string, string, int, bool) {
	var isFeed bool = false
	var nexTitle string = ""
	var lastTitleLevel int = 4
	var linecount = 0
	size := dataReader.Len()
	var headingsBuilder strings.Builder
	var preformattedTextBuilder strings.Builder

	scanner := bufio.NewScanner(dataReader)
	inPreformat := false
	for scanner.Scan() {
		linecount += 1
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if nexTitle == "" && strings.TrimSpace(line) != "" {
			if ContainsLetterRunes(line) {
				// Assume for nex documents that the first non-blank line is the title
				nexTitle = strings.TrimSpace(line)
			}
		}
		if inPreformat {
			fmt.Fprintf(strippedTextBuilder, "%s\n", line)
			fmt.Fprintf(&preformattedTextBuilder, "%s\n", line)
			continue
		}

		if strings.HasPrefix(line, "```") {
			inPreformat = !inPreformat
		} else if strings.HasPrefix(line, "###") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "###")))
			if nexTitle == "" || lastTitleLevel > 3 {
				nexTitle = string(strings.TrimSpace(strings.TrimPrefix(line, "###")))
				lastTitleLevel = 3
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "##") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "##")))
			if nexTitle == "" || lastTitleLevel > 2 {
				nexTitle = string(strings.TrimSpace(strings.TrimPrefix(line, "##")))
				lastTitleLevel = 2
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "#") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "#")))
			if nexTitle == "" || lastTitleLevel > 1 {
				nexTitle = string(strings.TrimSpace(strings.TrimPrefix(line, "#")))
				lastTitleLevel = 1
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "=>") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "=>"))
			fmt.Fprintf(strippedTextBuilder, "%s\n", line)
			link, title, _ := CutAny(line, " \t")

			link_without_fragment, _, _ := strings.Cut(link, "#")
			//link_without_query_and_fragment, _, _ = strings.Cut(link_without_query_and_fragment, "?")
			*links = append(*links, NexLink{title, link_without_fragment})

			if isTimeDate(title) {
				isFeed = true
			}
		} else if strings.HasPrefix(line, ">") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimPrefix(line, ">"))
		} else {
			fmt.Fprintf(strippedTextBuilder, "%s\n", line)
			continue
		}
	}

	return nexTitle, linecount, headingsBuilder.String(), preformattedTextBuilder.String(), size, isFeed
}
