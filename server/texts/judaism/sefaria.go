package judaism

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	// "net/url"

	"runtime"
	"strings"
	"time"
	"unicode/utf8"
)

func EncodeTextReference(s string) string {
	result := strings.Replace(s, " ", "_", -1)
	return result
}

/*
	type SefariaIndexCategory struct {
		Contents []SefariaIndexCategoryOrText `json:"contents"`
		Order float64 `json:"order"`
		EnComplete bool `json:"enComplete"`
		HeComplete bool `json:"heComplete"`
		EnDesc string `json:"enDesc"`
		HeDesc string `json:"heDesc"`
		EnShortDesc string `json:"enShortDesc"`
		HeShortDesc string `json:"heShortDesc"`
		SearchRoot string `json:"searchRoot"`
		HeCategory string `json:"heCategory"`
		Category string `json:"category"` // Title of Category
	}
*/
type SefariaIndexCategoryOrText struct {
	// Category Info
	Contents    []SefariaIndexCategoryOrText `json:"contents"`
	Order       float64                      `json:"order"` // Also in Text Info
	EnComplete  bool                         `json:"enComplete"`
	HeComplete  bool                         `json:"heComplete"`
	EnDesc      string                       `json:"enDesc"` // Also in Text Info
	HeDesc      string                       `json:"heDesc"` // Also in Text Info
	EnShortDesc string                       `json:"enShortDesc"`
	HeShortDesc string                       `json:"heShortDesc"`
	SearchRoot  string                       `json:"searchRoot"`
	HeCategory  string                       `json:"heCategory"`
	Category    string                       `json:"category"` // Title of Category

	// Text Info
	Categories        []string `json:"categories"`
	Dependence        string   `json:"dependence"`
	PrimaryCategory   string   `json:"primary_category"`
	CollectiveTitle   string   `json:"collective_title"`
	HeCollectiveTitle string   `json:"heCollectiveTitle"`
	Commentator       string   `json:"commentator"`
	HeCommentator     string   `json:"heCommentator"`
	BaseTextOrder     float64  `json:"base_text_order"`
	Corpus            string   `json:"corpus"`
	HeTitle           string   `json:"heTitle"`
	Title             string   `json:"title"`
}

func GetFullIndex() []SefariaIndexCategoryOrText {
	url := "https://sefaria.org/api/index"

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
		return []SefariaIndexCategoryOrText{}
	}

	req.Header.Set("User-Agent", "ScholasticDiversity")
	req.Header.Set("accept", "application/json")
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		fmt.Println(err)
		return []SefariaIndexCategoryOrText{}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println(err)
		return []SefariaIndexCategoryOrText{}
	}

	var sefariaIndex []SefariaIndexCategoryOrText
	jsonErr := json.Unmarshal(body, &sefariaIndex)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return []SefariaIndexCategoryOrText{}
	}

	return sefariaIndex
}

// Give array of category titles to go down
// TODO: Should not continue trying to find something when the exact ordering given in titles is not followed
func findInIndex(index []SefariaIndexCategoryOrText, titles []string) SefariaIndexCategoryOrText {
	for _, i := range index {
		if titles[0] == i.Category {
			if len(titles) == 1 {
				// This was the last thing to find, return it
				return i
			} else {
				// There's still more to find
				newTitles := titles[1:]
				return findInIndex(i.Contents, newTitles)
			}
		} else if len(i.Categories) != 0 && i.Title == titles[0] {
			// It is a text
			return i
		}
	}

	return SefariaIndexCategoryOrText{}
}

type SefariaText struct {
	Ref                      string          `json:"ref"`
	HeRef                    string          `json:"heRef"`
	IsComplex                bool            `json:"isComplex"`
	JsonText                 json.RawMessage `json:"text"` // This is used by the GetText function to fill in Text [][]string (normalizing the json to something consistent)
	Text                     [][]string      // This is what users will use
	JsonHeText               json.RawMessage `json:"he"`
	HeText                   [][]string
	Versions                 []SefariaTextVersion `json:"versions"` // Array of Versions
	TextDepth                int                  `json:"textDepth"`
	SectionNames             []string             `json:"sectionNames"` // Names of all section levels (ex: ["Chapter", "Verse"])
	AddressTypes             []string             `json:"addressTypes"`
	Lengths                  []int                `json:"lengths"` // Lengths for all section levels (from SectionNames field)
	Length                   int                  `json:"length"`  // Length in highest level section
	HeTitle                  string               `json:"heTitle"`
	TitleVariants            []string             `json:"titleVariants"`
	HeTitleVariants          []string             `json:"heTitleVariants"`
	Type                     string               `json:"type"` // Same as Categories[0]
	PrimaryCategory          string               `json:"primary_category"`
	Book                     string               `json:"book"`
	Categories               []string             `json:"categories"`
	Order                    interface{}          `json:"order"`      // NOTE: Can be []int or string
	Sections                 []json.RawMessage    `json:"sections"`   // TODO: Can be array of ints or strings
	ToSections               []json.RawMessage    `json:"toSections"` // Parallel to Sections filed, specifies end range (where the section is ending *to*)
	IsDependant              bool                 `json:"isDependant"`
	IndexTitle               string               `json:"indexTitle"`
	HeIndexTitle             string               `json:"heIndexTitle"`
	SectionRef               string               `json:"sectionRef"`
	FirstAvailableSectionRef string               `json:"firstAvailableSectionRef"`
	HeSectionRef             string               `json:"heSectionRef"`
	IsSpanning               bool                 `json:"isSpanning"`
	SpanningRefs             []string             `json:"spanningRefs"`
	Sources                  []string             `json:"sources"`

	// Version Info of Text field
	VersionTitle              string `json:"versionTitle"`
	VersionTitleInHebrew      string `json:"versionTitleInHebrew"`
	ShortVersionTitle         string `json:"shortVersionTitle"`
	ShortVersionTitleInHebrew string `json:"shortVersionTitleInHebrew"`
	VersionSource             string `json:"versionSource"`
	VersionStatus             string `json:"versionStatus"`
	VersionNotes              string `json:"versionNotes"`
	ExtendedNotes             string `json:"extendedNotes"`
	ExtendedNotesHebrew       string `json:"extendedNotesHebrew"`
	VersionNotesInHebrew      string `json:"versionNotesInHebrew"`
	DigitizedBySefaria        bool   `json:"digitizedBySefaria"`
	License                   string `json:"license"`
	FormatEnAsPoetry          bool   `json:"formatEnAsPoetry"`

	// Version Info of HeText field
	HeVersionTitle              string `json:"heVersionTitle"`
	HeVersionTitleInHebrew      string `json:"heVersionTitleInHebrew"`
	HeShortVersionTitle         string `json:"heShortVersionTitle"`
	HeShortVersionTitleInHebrew string `json:"heShortVersionTitleInHebrew"`
	HeVersionSource             string `json:"heVersionSource"`
	HeVersionStatus             string `json:"heVersionStatus"`
	HeVersionNotes              string `json:"heVersionNotes"`
	HeExtendedNotes             string `json:"heExtendedNotes"`
	HeExtendedNotesHebrew       string `json:"heExtendedNotesHebrew"`
	HeVersionNotesInHebrew      string `json:"heVersionNotesInHebrew"`
	HeDigitizedBySefaria        bool   `json:"heDigitizedBySefaria"`
	HeLicense                   string `json:"heLicense"`
	HeFormatHeAsPoetry          bool   `json:"formatHeAsPoetry"`

	JsonAlts json.RawMessage `json:"alts"`
	Alts     [][]SefariaTextAlt
	Next     string `json:"next"` // Ref of next section
	Prev     string `json:"prev"` // Ref of previous section

	// TODO: Don't know the types of these yet
	Commentary []SefariaTextLink `json:"commentary"`
	Sheets     []string          `json:"sheets"`
	Layer      []string          `json:"layer"`
}

type SefariaTextVersion struct {
	Title                     string      `json:"title"` // Title of text, not version title
	VersionTitle              string      `json:"versionTitle"`
	VersionSource             string      `json:"versionSource"`
	Language                  string      `json:"language"`
	Status                    string      `json:"status"`
	License                   string      `json:"license"`
	VersionNotes              string      `json:"versionNotes"`
	DigitizedBySefaria        interface{} `json:"digitizedBySefaria"` // NOTE: can be bool or empty string
	Priority                  interface{} `json:"priority"`           // NOTE: can be float64 or empty string
	VersionTitleInHebrew      string      `json:"versionTitleInHebrew"`
	VersionNotesInHebrew      string      `json:"versionNotesInHebrew"`
	ExtendedNotes             string      `json:"extendedNotes"`
	ExtendedNotesHebrew       string      `json:"extendedNotesHebrew"`
	PurchaseInformationImage  string      `json:"purchaseInformationImage"`
	PurchaseInformationURL    string      `json:"purchaseInformationURL"`
	ShortVersionTitle         string      `json:"shortVersionTitle"`
	ShortVersionTitleInHebrew string      `json:"shortVersionTitleInHebrew"`
}

type SefariaTextAlt struct {
	En    []string `json:"en"`
	He    []string `json:"he"`
	Whole bool     `json:"whole"`
}

// TODO: pad parameter will return entire books instead of going directly to the first section (Genesis will return entire book, instead of Genesis 1)
// context - specifying a verse will return it with its context
// commentary - whether to include a list of commentaries
func GetText(ref string, lang string, version string) SefariaText {
	apiUrl := "https://sefaria.org/api/texts/" + EncodeTextReference(ref)
	if lang != "" && version != "" {
		apiUrl += "/" + lang + "/" + version
	}

	fmt.Printf("Getting text for %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return SefariaText{}
	}

	req.Header.Set("User-Agent", "ScholasticDiversity")
	req.Header.Set("accept", "application/json")
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return SefariaText{}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, readErr)
		return SefariaText{}
	}

	sefariaText := SefariaText{}
	jsonErr := json.Unmarshal(body, &sefariaText)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return SefariaText{}
	}

	if utf8.Valid(sefariaText.JsonText) {
		if sefariaText.IsSpanning {
			// sefariaText.Text is [][]string, rather than []string
			jsonErr2 := json.Unmarshal(sefariaText.JsonText, &sefariaText.Text)
			if jsonErr2 != nil {
				_, filename, line, _ := runtime.Caller(0)
				fmt.Printf("%s:%d %s (`%s`)\n", filename, line, jsonErr2, string(sefariaText.JsonText))
				return SefariaText{}
			}
		} else {
			var t []string
			jsonErr2 := json.Unmarshal(sefariaText.JsonText, &t)
			if jsonErr2 != nil {
				_, filename, line, _ := runtime.Caller(0)
				fmt.Printf("%s:%d %s `%s`\n", filename, line, jsonErr2, string(sefariaText.JsonText))
				fmt.Printf("%s\n", body)
				return SefariaText{}
			}
			sefariaText.Text = append(sefariaText.Text, t)
		}
	}

	if utf8.Valid(sefariaText.JsonHeText) {
		if sefariaText.IsSpanning {
			// sefariaText.HeText is [][]string, rather than []string
			jsonErr2 := json.Unmarshal(sefariaText.JsonHeText, &sefariaText.HeText)
			if jsonErr2 != nil {
				_, filename, line, _ := runtime.Caller(0)
				fmt.Printf("%s:%d %s\n", filename, line, jsonErr2)
				return SefariaText{}
			}
		} else {
			var t []string
			jsonErr2 := json.Unmarshal(sefariaText.JsonHeText, &t)
			if jsonErr2 != nil {
				_, filename, line, _ := runtime.Caller(0)
				fmt.Printf("%s:%d %s\n", filename, line, jsonErr2)
				return SefariaText{}
			}
			sefariaText.HeText = append(sefariaText.HeText, t)
		}
	}

	if utf8.Valid(sefariaText.JsonAlts) {
		if sefariaText.IsSpanning {
			// sefariaText.Alts is [][]SefariaTextAlt, rather than []SefariaTextAlt
			jsonErr2 := json.Unmarshal(sefariaText.JsonAlts, &sefariaText.Alts)
			if jsonErr2 != nil {
				_, filename, line, _ := runtime.Caller(0)
				fmt.Printf("%s:%d %s\n", filename, line, jsonErr2)
				return SefariaText{}
			}
		} else {
			var t []SefariaTextAlt
			jsonErr2 := json.Unmarshal(sefariaText.JsonAlts, &t)
			if jsonErr2 != nil {
				_, filename, line, _ := runtime.Caller(0)
				fmt.Printf("%s:%d %s\n", filename, line, jsonErr2)
				return SefariaText{}
			}
			sefariaText.Alts = append(sefariaText.Alts, t)
		}
	}

	return sefariaText
}

// /api/v2/index/:title
type SefariaTextIndexRecord struct {
	Title      string   `json:"title"`
	Categories []string `json:"categories"`
	// Schema []SefariaTextIndexRecordSchema `json:"schema"`
	Order           []int    `json:"order"`
	Authors         []string `json:"authors"`
	EnDesc          string   `json:"enDesc"`
	HeDesc          string   `json:"heDesc"`
	EnShortDesc     string   `json:"enShortDesc"`
	HeShortDesc     string   `json:"heShortDesc"`
	PubDate         string   `json:"pubDate"`
	CompDate        string   `json:"compDate"`
	CompPlace       string   `json:"comPlace"`
	PubPlace        string   `json:"PubPlace"`
	ErrorMargin     string   `json:"errorMargin"`
	IsCited         bool     `json:"is_cited"`
	Corpora         []string `json:"corpora"`
	HeTitle         string   `json:"heTitle"`
	TitleVariants   []string `json:"titleVariants"`
	HeTitleVariants []string `json:"heTitleVariants"`
	// Alts SomeObjectHere `json:"alts"`
	SectionNames []string `json:"sectionNames"`
	Depth        int      `json:"depth"`
	HeCategories []string `json:"heCategories"`
	// CompDateString SomeObjHere `json:"compDateString"`
	// CompPlaceString SomeObjHere `json:"compPlaceString"`
}

func GetTextIndexRecord() SefariaTextIndexRecord {

	return SefariaTextIndexRecord{}
}

type SefariaTextLink struct {
	Id                string   `json:"_id"`
	IndexTitle        string   `json:"index_title"`
	Category          string   `json:"category"`
	Type              string   `json:"type"`
	Ref               string   `json:"ref"` // NOTE: Looks like this is the same as SourceRef
	AnchorRef         string   `json:"anchorRef"`
	AnchorRefExpanded []string `json:"anchorRefExpanded"`
	SourceRef         string   `json:"sourceRef"`
	SourceHeRef       string   `json:"sourceHeRef"`
	AnchorVerse       int      `json:"anchorVerse"`
	SourceHasEn       bool     `json:"sourceHasEn"`
	CompDate          []int    `json:"compDate"`
	ErrorMargin       int      `json:"errorMargin"`
	// AnchorVersion
	//sourceVersion
	DisplayedText   SefariaHebrewEnglishText `json:"displayedText"`
	CommentaryNum   float64                  `json:"commentaryNum"` // Will have decimal if the commentary covers only a verse (a part of the given ref); will be whole number for number of commentaries on whole given ref
	CollectiveTitle SefariaHebrewEnglishText `json:"collectiveTitle"`
	HeTitle         string                   `json:"heTitle"`
}

func GetLinks(ref string, lang string, version string) []SefariaTextLink {
	apiUrl := "https://sefaria.org/api/links/" + EncodeTextReference(ref)
	apiUrl += "?with_text=0"

	fmt.Printf("Getting links for %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return []SefariaTextLink{}
	}

	req.Header.Set("User-Agent", "ScholasticDiversity")
	req.Header.Set("accept", "application/json")
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return []SefariaTextLink{}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, readErr)
		return []SefariaTextLink{}
	}

	sefariaLinks := []SefariaTextLink{}
	jsonErr := json.Unmarshal(body, &sefariaLinks)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return []SefariaTextLink{}
	}

	return sefariaLinks
}

type SefariaCalendarResponse struct {
	Date          string                `json:"date"`
	Timezone      string                `json:"timezone"`
	CalendarItems []SefariaCalendarItem `json:"calendar_items"`
}

type SefariaCalendarItem struct {
	Title        SefariaHebrewEnglishText `json:"title"`
	DisplayValue SefariaHebrewEnglishText `json:"displayValue"`
	Url          string                   `json:"url"`
	Ref          string                   `json:"ref"`
	HeRef        string                   `json:"heRef"`
	Order        int                      `json:"order"`
	Category     string                   `json:"category"`
	// ExtraDetails SomeObjectHere `json:"extraDetails"`
	Description SefariaHebrewEnglishText `json:"description"`
}

type SefariaHebrewEnglishText struct {
	English string `json:"en"`
	Hebrew  string `json:"he"`
}

// TODO: Add parameters for day, month, year, timezone, diaspora, and custom (ashkenazi vs. sephardi)
func GetCalendars() SefariaCalendarResponse {
	url := "https://sefaria.org/api/calendars"

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
		return SefariaCalendarResponse{}
	}

	req.Header.Set("User-Agent", "ScholasticDiversity")
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		fmt.Println(err)
		return SefariaCalendarResponse{}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println(err)
		return SefariaCalendarResponse{}
	}

	sefariaCalendarResponse := SefariaCalendarResponse{}
	jsonErr := json.Unmarshal(body, &sefariaCalendarResponse)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return SefariaCalendarResponse{}
	}

	return sefariaCalendarResponse
}
