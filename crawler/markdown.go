package crawler

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

type MarkdownLink struct {
	name string
	url  string
}

func (ctx *CrawlContext) GetMarkdownPageInfo(dataReader *bytes.Reader, tagsMap *map[string]float64, mentionsMap *map[string]bool, links *[]MarkdownLink, strippedTextBuilder *strings.Builder, update bool) (string, int, string, string, int, bool) {
	var isFeed bool = false
	var mdTitle string = ""
	var lastTitleLevel int = 5
	var linecount = 0
	size := dataReader.Len()
	var headingsBuilder strings.Builder
	var preformattedTextBuilder strings.Builder

	scanner := bufio.NewScanner(dataReader)
	inPreformat := false
	inYamlMetadata := false
	for scanner.Scan() {
		linecount += 1
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if mdTitle == "" && strings.TrimSpace(line) != "" {
			if ContainsLetterRunes(line) {
				// Assume for markdown documents that the first non-blank line is the title until we reach the first line with a #
				mdTitle = strings.TrimSpace(line)
			}
		}
		if inPreformat {
			fmt.Fprintf(strippedTextBuilder, "%s\n", line)
			fmt.Fprintf(&preformattedTextBuilder, "%s\n", line)
			continue
		}

		if strings.HasPrefix(line, "```") {
			inPreformat = !inPreformat
		} else if strings.HasPrefix(line, "####") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "####")))
			if mdTitle == "" || lastTitleLevel > 4 {
				mdTitle = string(strings.TrimSpace(strings.TrimPrefix(line, "####")))
				lastTitleLevel = 4
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "###") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "###")))
			if mdTitle == "" || lastTitleLevel > 3 {
				mdTitle = string(strings.TrimSpace(strings.TrimPrefix(line, "###")))
				lastTitleLevel = 3
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "##") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "##")))
			if mdTitle == "" || lastTitleLevel > 2 {
				mdTitle = string(strings.TrimSpace(strings.TrimPrefix(line, "##")))
				lastTitleLevel = 2
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, "#") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(strings.TrimPrefix(line, "#")))
			if mdTitle == "" || lastTitleLevel > 1 {
				mdTitle = string(strings.TrimSpace(strings.TrimPrefix(line, "#")))
				lastTitleLevel = 1
			}
			fmt.Fprintf(&headingsBuilder, "%s\n", strings.TrimSpace(line))
		} else if strings.HasPrefix(line, ">") {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimPrefix(line, ">"))
		} else if strings.HasPrefix(line, "---") && (linecount == 1 || inYamlMetadata) {
			inYamlMetadata = !inYamlMetadata
		} else if inYamlMetadata {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "title:") {
				_, value, hasValue := strings.Cut(line, ":")
				if hasValue && value != "" && lastTitleLevel > 1 {
					mdTitle = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "\""), "\"")
					lastTitleLevel = 1
				}
			} /*else if strings.HasPrefix(line, "date:") {
			}*/
		} else {
			fmt.Fprintf(strippedTextBuilder, "%s\n", strings.TrimSpace(line))

			// Small paragraph parser for getting links
			var currentLinkTitleBuilder strings.Builder
			var currentLinkUrlBuilder strings.Builder

			inBracket := false
			expectingLinkUrlStart := false
			inLinkUrl := false
			escaped := false
			for _, r := range line {
				if r == '\\' {
					escaped = true
					continue
				}
				if inBracket {
					if r == '[' && !escaped {
						currentLinkTitleBuilder.Reset()
					} else if r == ']' && !escaped {
						inBracket = false
						expectingLinkUrlStart = true
					} else {
						if escaped {
							currentLinkTitleBuilder.WriteRune('\\')
						}
						// Add rune to current link title
						currentLinkTitleBuilder.WriteRune(r)
					}
				} else if expectingLinkUrlStart {
					if r == '(' && !escaped {
						inLinkUrl = true
						currentLinkUrlBuilder.Reset()
					} else {
						currentLinkTitleBuilder.Reset()
					}
					expectingLinkUrlStart = false
				} else if inLinkUrl {
					if r == ')' && !escaped {
						// End of entire link - add it to links list
						link_title := currentLinkTitleBuilder.String()
						url_without_fragment, _, _ := strings.Cut(currentLinkUrlBuilder.String(), "#")
						(*links) = append((*links), MarkdownLink{currentLinkTitleBuilder.String(), url_without_fragment})
						if isTimeDate(link_title) {
							isFeed = true
						}

						inLinkUrl = false
						currentLinkTitleBuilder.Reset()
						currentLinkUrlBuilder.Reset()
					} else {
						if escaped {
							currentLinkUrlBuilder.WriteRune('\\')
						}
						// Add rune to current line url
						currentLinkUrlBuilder.WriteRune(r)
					}
				} else if r == '[' && !escaped {
					inBracket = true
					currentLinkTitleBuilder.Reset()
				}
				if escaped {
					escaped = false
				}
			}
		}
	}

	return mdTitle, linecount, headingsBuilder.String(), preformattedTextBuilder.String(), size, isFeed
}
