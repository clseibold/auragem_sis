package christianity

import (
	"encoding/json"
	"fmt"
	"net/http"
	// "net/url"
	"time"
	"io/ioutil"
	// "unicode/utf8"
	"runtime"
	// "strings"
)

type BibleLanguage struct {
	Id string `json:"id"`
	Name string `json:"name"`
	NameLocal string `json:"nameLocal"`
	Script string `json:"script"`
	ScriptDirection string `json:"scriptDirection"`
}

type BibleVersionWrapper struct {
	Data BibleVersion `json:"data"`
}
type BibleVersion struct {
	Language BibleLanguage `json:"language"`
	Name string `json:"name"`
	Id string `json:"id"`
	Abbreviation string `json:"abbreviation"`
	Description string `json:"description"`

	Copyright string `json:"copyright"`
	Info string `json:"info"`
	UpdatedAt string `json:"updatedAt"`
	Type string `json:"type"`

	// Countries
	// AudioBibles
}
func GetBibleVersion(versionId string, apiKey string) BibleVersion {
	apiUrl := "https://api.scripture.api.bible/v1/bibles/" + versionId

	fmt.Printf("Getting books from %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return BibleVersion {}
	}

	req.Header.Set("User-Agent", "AuraGem")
	req.Header.Set("api-key", apiKey)
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return BibleVersion {}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, readErr)
		return BibleVersion {}
	}

	bibleVersion := BibleVersionWrapper {}
	jsonErr := json.Unmarshal(body, &bibleVersion)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return BibleVersion {}
	}

	return bibleVersion.Data
}

type BibleBooksWrapper struct {
	Data []BibleBook `json:"data"`
}
type BibleBookWrapper struct {
	Data BibleBook `json:"data"`
}
type BibleBook struct {
	Id string `json:"id"`
	BibleId string `json:"bibleId"`
	Name string `json:"name"`
	NameLong string `json:"nameLong"`
	Abbreviation string `json:"abbreviation"`
	Chapters []BibleChapter `json:"chapters"`
}
func GetBooks(versionId string, apiKey string) []BibleBook {
	apiUrl := "https://api.scripture.api.bible/v1/bibles/" + versionId + "/books"

	fmt.Printf("Getting books from %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return []BibleBook {}
	}

	req.Header.Set("User-Agent", "AuraGem")
	req.Header.Set("api-key", apiKey)
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return []BibleBook {}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, readErr)
		return []BibleBook {}
	}

	bibleBooks := BibleBooksWrapper {}
	jsonErr := json.Unmarshal(body, &bibleBooks)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return []BibleBook {}
	}

	return bibleBooks.Data
}
func GetBook(versionId string, bookId, apiKey string, withChapters bool) BibleBook {
	apiUrl := "https://api.scripture.api.bible/v1/bibles/" + versionId + "/books/" + bookId
	if withChapters {
		apiUrl += "?include-chapters=true"
	}

	fmt.Printf("Getting book from %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return BibleBook {}
	}

	req.Header.Set("User-Agent", "AuraGem")
	req.Header.Set("api-key", apiKey)
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return BibleBook {}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, readErr)
		return BibleBook {}
	}

	bibleBook := BibleBookWrapper {}
	jsonErr := json.Unmarshal(body, &bibleBook)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return BibleBook {}
	}

	return bibleBook.Data
}

type BibleChaptersWrapper struct {
	Data []BibleChapter `json:"data"`
}
type BibleChapterWrapper struct {
	Data BibleChapter `json:"data"`
}
type BibleChapter struct {
	Id string `json:"id"`
	BibleId string `json:"bibleId"`
	Number string `json:"number"`
	Position int `json:"position"` // Only sometimes used
	BookId string `json:"bookId"`
	Reference string `json:"reference"` // Seems like this is the book name

	Content string `json:"content"`
	VerseCount int `json:"verseCount"`
	NextChapter BibleChapterReference `json:"next"`
	PreviousChapter BibleChapterReference `json:"previous"`
	Copyright string `json:"copyright"`
}
type BibleChapterReference struct {
	Id string `json:"id"`
	BookId string `json:"bookId"`
	Number string `json:"number"`
}
func GetChapters(versionId string, bookId string, apiKey string) []BibleChapter {
	apiUrl := "https://api.scripture.api.bible/v1/bibles/" + versionId + "/books/" + bookId + "/chapters"

	fmt.Printf("Getting chapters from %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return []BibleChapter {}
	}

	req.Header.Set("User-Agent", "AuraGem")
	req.Header.Set("api-key", apiKey)
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return []BibleChapter {}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, readErr)
		return []BibleChapter {}
	}

	bibleChapters := BibleChaptersWrapper {}
	jsonErr := json.Unmarshal(body, &bibleChapters)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return []BibleChapter {}
	}

	return bibleChapters.Data
}
func GetChapter(versionId string, chapterId string, apiKey string) BibleChapter {
	apiUrl := "https://api.scripture.api.bible/v1/bibles/" + versionId + "/chapters/" + chapterId + "?content-type=text&include-verse-numbers=true&include-notes=true&include-titles=true"

	fmt.Printf("Getting chapter from %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return BibleChapter {}
	}

	req.Header.Set("User-Agent", "AuraGem")
	req.Header.Set("api-key", apiKey)
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return BibleChapter {}
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	/*if res.Status != 200 {

	}*/

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, readErr)
		return BibleChapter {}
	}

	bibleChapter := BibleChapterWrapper {}
	jsonErr := json.Unmarshal(body, &bibleChapter)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return BibleChapter {}
	}

	return bibleChapter.Data
}