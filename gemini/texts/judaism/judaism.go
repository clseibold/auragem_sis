package judaism

import (
	// "context"
	// "database/sql"
	// "time"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	// "math"

	// "github.com/krixano/ponixserver/src/db"

	"runtime"

	sis "gitlab.com/clseibold/smallnetinformationservices"
	"golang.org/x/net/html"
	// "github.com/microcosm-cc/bluemonday"
)

func HandleJewishTexts(g sis.ServerHandle) {
	//stripTags := bluemonday.StripTagsPolicy()

	index := GetFullIndex()
	//indexMap := make(map[string]int)

	g.AddRoute("/scriptures/jewish/", func(request sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure(err.Error())
			return
		} else if query == "" {
			handleIndex(index, request)
		} else {
			handleCategory(index, query, request)
		}
	})

	g.AddRoute("/scriptures/jewish/t/:ref", func(request sis.Request) {
		ref := request.GetParam("ref")
		handleText(ref, request)
	})
}

func handleIndex(index []SefariaIndexCategoryOrText, request sis.Request) {
	var builder strings.Builder
	for _, category := range index {
		fmt.Fprintf(&builder, "=> /scriptures/jewish/?%s %s\n", url.QueryEscape(category.Category), category.Category)
	}

	calendars := GetCalendars()
	fmt.Fprintf(&builder, "\n## Learning Schedules for %s\n", calendars.Date)
	for _, calendar := range calendars.CalendarItems {
		// Special handling of parashot
		if calendar.Title.English == "Parashat Hashavua" {
			fmt.Fprintf(&builder, "### Parashat %s (%s)\n", calendar.DisplayValue.English, calendar.Title.Hebrew+" "+calendar.DisplayValue.Hebrew)
			fmt.Fprintf(&builder, "=> /scriptures/jewish/t/%s %s\n", EncodeTextReference(calendar.Ref), calendar.Ref)
			if calendar.Description.English != "" {
				fmt.Fprintf(&builder, "%s\n\n", calendar.Description.English)
			}
		} else if calendar.Title.English == "Tanakh Yomi" {
			fmt.Fprintf(&builder, "### %s (%s)\n", calendar.Title.English, calendar.Title.Hebrew)
			fmt.Fprintf(&builder, "=> /scriptures/jewish/t/%s %s (%s)\n", EncodeTextReference(calendar.Ref), calendar.DisplayValue.English, calendar.Ref)
			if calendar.Description.English != "" {
				fmt.Fprintf(&builder, "%s\n\n", calendar.Description.English)
			}
		} else {
			fmt.Fprintf(&builder, "### %s (%s)\n", calendar.Title.English, calendar.Title.Hebrew)
			fmt.Fprintf(&builder, "=> /scriptures/jewish/t/%s %s\n", EncodeTextReference(calendar.Ref), calendar.DisplayValue.English)
			if calendar.Description.English != "" {
				fmt.Fprintf(&builder, "%s\n\n", calendar.Description.English)
			}
		}
	}

	template := `# Jewish Texts (Sefaria Proxy)

=> https://sefaria.org Powered by Sefaria.org

%s
`
	request.Gemini(fmt.Sprintf(template, builder.String()))
}

func handleCategory(index []SefariaIndexCategoryOrText, query string, request sis.Request) {
	categories := strings.Split(query, "/")
	var categoryStringBuilder strings.Builder
	for i, c := range categories {
		if i == 0 {
			fmt.Fprintf(&categoryStringBuilder, "%s ", c)
		} else {
			fmt.Fprintf(&categoryStringBuilder, "> %s ", c)
		}
	}

	categoryOrText := findInIndex(index, categories)
	/*if categoryOrText.Title == "" && categoryOrText.Category == "" {
		// TODO: Could not find text/category
	}*/

	var builder strings.Builder
	for _, category := range categoryOrText.Contents {
		if category.Title != "" {
			// A text
			fmt.Fprintf(&builder, "=> /scriptures/jewish/t/%s %s\n", url.PathEscape(category.Title), category.Title)
		} else {
			fmt.Fprintf(&builder, "=> /scriptures/jewish/?%s %s\n", url.QueryEscape(query+"/"+category.Category), category.Category)
		}
	}

	request.Gemini(fmt.Sprintf(`# %s

=> https://sefaria.org Powered by Sefaria.org

%s
`, categoryStringBuilder.String(), builder.String()))
}

func handleText(ref string, request sis.Request) {
	text := GetText(ref, "", "" /*"Tanakh: The Holy Scriptures, published by JPS"*/)

	var startingChapterInRef int = 1
	if text.TextDepth-1 >= 0 && text.TextDepth-1 <= len(text.Sections) {
		if utf8.Valid(text.Sections[text.TextDepth-2]) {
			i, err := strconv.Atoi(string(text.Sections[text.TextDepth-2]))
			if err == nil {
				startingChapterInRef = i
				fmt.Printf("%s\n", err)
			}
		}
	}

	var startingVerseInRef int = 1
	/*if text.TextDepth <= len(text.Sections) {
		if utf8.Valid(text.Sections[text.TextDepth - 1]) {
			i, err := strconv.Atoi(string(text.Sections[text.TextDepth - 1]))
			if err == nil {
				startingVerseInRef = i
				fmt.Printf("%s\n", err)
			}
		}
	}*/

	//endingVerseInRef := text.ToSections[text.TextDepth]

	// TODO: For the tanakh, add in paragraph divisions somehow (The 929 website seems to be able to do this)
	var builder strings.Builder
	for spanningInt, spanningSection := range text.Text {
		if spanningInt != 0 {
			fmt.Fprintf(&builder, "\n\n")
		}
		fmt.Fprintf(&builder, "%d ", spanningInt+startingChapterInRef)
		for verse, t := range spanningSection {
			verseNumber := verse + 1
			if spanningInt == 0 {
				verseNumber += startingVerseInRef - 1
			}
			finalText := strings.Replace(t, "יהוה", "the LORD", -1) // TODO: Make this a setting?
			finalText = strings.Replace(finalText, "<br>", "\n", -1)
			finalText = parseHtmlFromString(finalText)
			if text.PrimaryCategory == "Tanakh" {
				fmt.Fprintf(&builder, "\u200E(%d) %s ", verseNumber, finalText)
			} else {
				fmt.Fprintf(&builder, "\u200E[%d] %s ", verseNumber, finalText)
			}
			if text.PrimaryCategory == "Talmud" || text.PrimaryCategory == "Mishnah" {
				fmt.Fprintf(&builder, "\n\n")
			} else if text.PrimaryCategory == "Commentary" {
				fmt.Fprintf(&builder, "\n")
			}
		}
	}

	if text.Next != "" || text.Prev != "" {
		fmt.Fprintf(&builder, "\n\n")
	}
	if text.Prev != "" {
		fmt.Fprintf(&builder, "=> /scriptures/jewish/t/%s Previous\n", url.PathEscape(text.Prev))
	}
	if text.Next != "" {
		fmt.Fprintf(&builder, "=> /scriptures/jewish/t/%s Next", url.PathEscape(text.Next))
	}

	// Commentary links to display on the text's page. The other commentaries are listed on a "More Commentaries" page
	// TODO: Add Abarbanel (Which is considered in the category of "Quoting Commentary")
	// TODO: Add Gersonides/Ralbag
	allowedCommentaries := map[string]bool{
		"Rashi":                          true,
		"Rashbam":                        true,
		"Ramban":                         true,
		"Sforno":                         true,
		"Ibn Ezra":                       true,
		"Radak":                          true,
		"Or HaChaim":                     true,
		"Tosefta":                        true,
		"Tosafot":                        true,
		"Tur":                            true,
		"Sefer Mitzvot Gadol":            true,
		"Contemporary Halakhic Problems": true,
		"Steinsaltz":                     true,

		// Specific to Jerusalem Talmud
		"Penei Moshe":                    true,
		"Mareh HaPanim":                  true,
		"Notes by Heinrich Guggenheimer": true,
	}

	if text.PrimaryCategory == "Tanakh" || text.PrimaryCategory == "Talmud" || text.PrimaryCategory == "Mishnah" {
		links := GetLinks(text.Ref, "", "") // Get Commentaries
		dict := make(map[string]bool)

		fmt.Fprintf(&builder, "\n\n##Commentaries\n\n")
		if text.PrimaryCategory == "Mishnah" {
			fmt.Fprintf(&builder, "=> /scriptures/jewish/t/%s Tosefta\n", url.PathEscape("Tosefta "+text.Ref))
		}
		for _, link := range links {
			if (link.Category != "Commentary" && link.Category != "Halakhah" && link.Category != "Targum") || !link.SourceHasEn {
				continue
			}
			if _, ok := allowedCommentaries[link.CollectiveTitle.English]; !ok && !strings.HasPrefix(link.CollectiveTitle.English, "Mishneh Torah") && !strings.HasPrefix(link.CollectiveTitle.English, "Shulchan Arukh") && !strings.HasPrefix(link.CollectiveTitle.English, "Mishnah") && !strings.HasPrefix(link.CollectiveTitle.English, "Targum") && !strings.HasPrefix(link.CollectiveTitle.English, "Onkelos") /*&& link.Category != "Talmud"*/ {
				continue
			}
			if _, ok := dict[link.IndexTitle]; !ok {
				// Add to map so we can check whether it repeats, then print the first reference.
				dict[link.IndexTitle] = true
				fmt.Fprintf(&builder, "=> /scriptures/jewish/t/%s %s\n", url.PathEscape(link.Ref), link.IndexTitle)
			}
		}
	}

	request.Gemini(fmt.Sprintf(`# %s

=> /scriptures/jewish Home
=> /scriptures/jewish?%s %s

%s

## Version Info

Version: %s
=> %s Source: %s
License: %s

=> /scriptures/jewish Jewish Texts
=> https://sefaria.org Powered by Sefaria.org`, text.Ref, url.QueryEscape(strings.Join(text.Categories, "/")), text.Categories[len(text.Categories)-1] /*stripTags.Sanitize(*/, builder.String() /*)*/, text.VersionTitle, text.VersionSource, text.VersionSource, text.License))
}

func parseHtmlFromString(text string) string {
	root, err := html.Parse(strings.NewReader(text))
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return ""
	}

	var builder strings.Builder
	var footNotes []string
	doTraverse(root, &builder, &footNotes)
	//fmt.Printf("Text: %s", builder.String())
	return builder.String()
}

func doTraverse(root *html.Node, builder *strings.Builder, footNotes *[]string) {
	var traverse func(n *html.Node) *html.Node
	traverse = func(n *html.Node) *html.Node {
		isFootnote := false
		for _, attribute := range n.Attr {
			if attribute.Key == "class" && attribute.Val == "footnote" {
				isFootnote = true
			}
		}

		if isFootnote {
			return nil
		}

		if n.Type == html.ElementNode {
			if n.Data == "b" || n.Data == "strong" {
				fmt.Fprintf(builder, "**")
			} else if n.Data == "i" {
				fmt.Fprintf(builder, "*")
			}
		}

		// Go through every descendant, depth-first
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode && n.Data != "sup" && n.Data != "a" {
				fmt.Fprintf(builder, "%s", c.Data)
			}

			res := traverse(c)
			if res != nil {
				return res
			}
		}

		if n.Type == html.ElementNode {
			if n.Data == "b" || n.Data == "strong" {
				fmt.Fprintf(builder, "**")
			} else if n.Data == "i" {
				fmt.Fprintf(builder, "*")
			}
		}

		return nil
	}

	traverse(root)
}
