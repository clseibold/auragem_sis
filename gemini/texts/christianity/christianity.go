package christianity

import (
	// "context"
	// "database/sql"
	// "time"
	"fmt"
	"net/url"
	"strings"

	// "strconv"
	// "unicode/utf8"

	// "github.com/krixano/ponixserver/src/db"

	"gitlab.com/clseibold/auragem_sis/config"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
	// "runtime"
	// "golang.org/x/net/html"
	// "github.com/microcosm-cc/bluemonday"
)

// KJV -> Revised Version (RV) -> ASV -> RSV  -> NRSV
//                                          \--> ESV

type BibleVersionListItem struct {
	Name string
	Id   string
}

var apiKey = config.APIBibleApiKey

func HandleChristianTexts(g sis.ServerHandle) {
	//apiKey := "df9786562778d4ff76d7a9ac6bcb149f"

	englishBibleVersions := []BibleVersionListItem{
		// Popular (ASV, Douay-Rheims, and Brenton English Septuagint)
		{"The Holy Bible, American Standard Version", "06125adad2d5898a-01"},
		{"American Standard Version (Byzantine Text with Apocrypha)", "685d1470fe4d5c3b-01"},
		{"Douay-Rheims American 1899", "179568874c45066f-01"},
		{"Brenton English Septuagint (Updated Spelling and Formatting)", "6bab4d6c61b31b80-01"},
		{"", ""},

		// KJV
		{"Cambridge Paragraph Bible of the KJV", "55212e3cf5d04d49-01"},
		{"King James (Authorised) Version", "de4e12af7f28f599-02"},
		{"Revised Version 1885", "40072c4a5aba4022-01"},
		{"JPS TaNaKH 1917 (aka. OJPS)", "bf8f1c7f3f9045a5-01"},
		{"", ""},

		// Others
		{"Geneva Bible", "c315fa9f71d4af3a-01"},
		{"World English Bible", "9879dbb7cfe39e4d-04"},
		{"World English Bible British Edition", "7142879509583d59-04"},
		{"World Messianic Bible (aka. Hebrew Names Version)", "f72b840c855f362c-04"},
		{"World Messianic Bible British Edition", "04da588535d2f823-04"},
		{"Targum Onkelos Etheridge", "ec290b5045ff54a5-01"},
	}

	spanishBibleVersions := []BibleVersionListItem{
		{"Reina Valera 1909", "592420522e16049f-01"},
	}

	germanBibleVersions := []BibleVersionListItem{
		{"Elberfelder Translation (Version of bibelkommentare.de)", "f492a38d0e52db0f-01"},
		{"German Luther Bible 1912 with Strong's numbers", "926aa5efbc5e04e2-01"},
	}

	arabicBibleVersions := []BibleVersionListItem{
		{"Biblica® Open New Arabic Version 2012", "b17e246951402e50-01"},
	}

	hebrewBibleVersions := []BibleVersionListItem{
		{"Biblica® Open Hebrew Living New Testament 2009", "a8a97eebae3c98e4-01"},
	}

	greekBibleVersions := []BibleVersionListItem{
		{"Brenton Greek Septuagint", "c114c33098c4fef1-01"},
		{"Greek Textus Receptus", "3aefb10641485092-01"},
	}

	italianBibleVersions := []BibleVersionListItem{
		{"Diodati Bible 1885", "41f25b97f468e10b-01"},
	}

	// Dutch Bible 1939 ead7b4cc5007389c-01
	// Swedish Core Bible - expanded fa4317c59f0825e0-01
	// Serbian Bible 06995ce9cd23361b-01
	// Ukrainian, Biblica® Open New Ukrainian Translation 2022 6c696cd1d82e2723-04
	// Hindi, Indian Revised Version(IRV) Hindi - 2019 1e8ab327edbce67f-01
	// Hindi Contemporary Version 2019 2133003bb8b5e62b-01

	// Cache the books from the ASV version of the bible. These should be the same for the ESV bible as well. Note: This does not include the apocrypha.
	asvBooks := GetBooks(englishBibleVersions[0].Id, apiKey)

	g.AddRoute("/scriptures/christian/", func(request sis.Request) {
		request.SetNoLanguage()
		request.SetClassification(sis.ScrollResponseUDC_Scripture)
		var builder strings.Builder
		fmt.Fprintf(&builder, "## Bible Versions\n")
		fmt.Fprintf(&builder, "### English\n")
		fmt.Fprintf(&builder, "=> /scriptures/christian/bible/esv/ ESV Bible\n")
		for _, version := range englishBibleVersions {
			if version.Id != "" {
				fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/ %s\n", version.Id, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### Spanish\n")
		for _, version := range spanishBibleVersions {
			if version.Id != "" {
				fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/ %s\n", version.Id, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### German\n")
		for _, version := range germanBibleVersions {
			if version.Id != "" {
				fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/ %s\n", version.Id, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### Arabic\n")
		for _, version := range arabicBibleVersions {
			if version.Id != "" {
				fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/ %s\n", version.Id, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### Italian\n")
		for _, version := range italianBibleVersions {
			if version.Id != "" {
				fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/ %s\n", version.Id, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### Modern Hebrew\n")
		for _, version := range hebrewBibleVersions {
			if version.Id != "" {
				fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/ %s\n", version.Id, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### Greek\n")
		for _, version := range greekBibleVersions {
			if version.Id != "" {
				fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/ %s\n", version.Id, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}

		request.Gemini(fmt.Sprintf(`# Christian Texts

=> https://scripture.api.bible Bibles Powered by API.Bible

%s

Tags: #bible #new #old #testament #septuagint #pentateuch
`, builder.String()))
	})

	g.AddRoute("/scriptures/christian/bible/esv/", func(request sis.Request) {
		request.SetLanguage("en-US")
		request.SetClassification(sis.ScrollResponseUDC_Scripture)
		var builder strings.Builder
		for _, book := range asvBooks {
			fmt.Fprintf(&builder, "=> /scriptures/christian/bible/esv/%s/ %s\n", url.PathEscape(book.Name+" 1"), book.Name)
		}

		request.Gemini(fmt.Sprintf(`# ESV Bible

=> https://api.esv.org/ Powered by Crossway's ESV API

%s

## ESV Copyright Notice

Scripture quotations marked “ESV” are from the ESV® Bible (The Holy Bible, English Standard Version®), copyright © 2001 by Crossway, a publishing ministry of Good News Publishers. Used by permission. All rights reserved. The ESV text may not be quoted in any publication made available to the public by a Creative Commons license. The ESV may not be translated into any other language.

Users may not copy or download more than 500 verses of the ESV Bible or more than one half of any book of the ESV Bible.
`, builder.String()))
	})

	g.AddRoute("/scriptures/christian/bible/esv/:text", func(request sis.Request) {
		request.SetLanguage("en-US")
		request.SetClassification(sis.ScrollResponseUDC_Scripture)
		text := request.GetParam("text")
		resp := GetPassages(text)
		var builder strings.Builder
		for _, s := range resp.Passages {
			fmt.Fprintf(&builder, "%s", s)
		}

		request.Gemini(fmt.Sprintf(`# ESV: %s

%s
`, resp.Canonical, builder.String()))
	})

	g.AddRoute("/scriptures/christian/bible/:id", func(request sis.Request) {
		versionId := request.GetParam("id")
		version := GetBibleVersion(versionId, apiKey)
		request.SetLanguage(version.Language.Id)
		request.SetClassification(sis.ScrollResponseUDC_Scripture)
		books := GetBooks(versionId, apiKey)
		var builder strings.Builder
		for _, book := range books {
			fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/%s/ %s\n", versionId, book.Id, book.Name)
		}

		request.Gemini(fmt.Sprintf(`# %s

=> /scriptures/christian/ Bible Versions

%s

Description: %s
Copyright: %s

=> https://scripture.api.bible Powered by API.Bible`, version.Name, builder.String(), version.Description, version.Copyright))
	})

	g.AddRoute("/scriptures/christian/bible/:id/:book", func(request sis.Request) {
		versionId := request.GetParam("id")
		bookId := request.GetParam("book")
		version := GetBibleVersion(versionId, apiKey)
		request.SetLanguage(version.Language.Id)
		request.SetClassification(sis.ScrollResponseUDC_Scripture)
		book := GetBook(versionId, bookId, apiKey, true)
		var builder strings.Builder
		for _, chapter := range book.Chapters {
			fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/chapter/%s/ Chapter %s\n", versionId, chapter.Id, chapter.Number)
		}

		request.Gemini(fmt.Sprintf(`# %s: %s

=> /scriptures/christian/bible/%s/ Books

%s

%s

=> https://scripture.api.bible Powered by API.Bible`, version.Abbreviation, book.Name, versionId, builder.String(), version.Copyright))
	})

	g.AddRoute("/scriptures/christian/bible/:id/chapter/:chapter", func(request sis.Request) {
		versionId := request.GetParam("id")
		chapterId := request.GetParam("chapter")
		version := GetBibleVersion(versionId, apiKey)
		request.SetLanguage(version.Language.Id)
		request.SetClassification(sis.ScrollResponseUDC_Scripture)
		chapter := GetChapter(versionId, chapterId, apiKey)
		var builder strings.Builder
		fmt.Fprintf(&builder, "%s", chapter.Content)
		/*for _, chapter := range book.Chapters {
			fmt.Fprintf(&builder, "=> /scriptures/christian/bible/%s/%s/%s Chapter %s\n", versionId, book.Id, chapter.Id, chapter.Number)
		}*/

		request.Gemini(fmt.Sprintf(`# %s: %s

=> /scriptures/christian/bible/%s/%s/ Chapters

%s

=> /scriptures/christian/bible/%s/chapter/%s/ Previous
=> /scriptures/christian/bible/%s/chapter/%s/ Next

%s

=> https://scripture.api.bible Powered by API.Bible`, version.Abbreviation, chapter.Reference, versionId, chapter.BookId, builder.String(), versionId, chapter.PreviousChapter.Id, versionId, chapter.NextChapter.Id, version.Copyright))
	})
}
