package christianity

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"gitlab.com/clseibold/auragem_sis/config"
)

// https://api.esv.org/v3/passage/text/?q=

type PassageResponse struct {
	Query       string        `json:"query"`
	Canonical   string        `json:"canonical"`
	Parsed      [][]int       `json:"parsed"`
	PassageMeta []PassageMeta `json:"passage_meta"`
	Passages    []string      `json:"passages"`
}
type PassageMeta struct {
	Canonical    string `json:"canonical"`
	ChapterStart []int  `json:"chapter_start"`
	ChapterEnd   []int  `json:"chapter_end"`
	PrevVerse    int    `json:"prev_verse"`
	NextVerse    int    `json:"next_verse"`
	PrevChapter  []int  `json:"prev_chapter"`
	NextChapter  []int  `json:"next_chapter"`
}

func GetPassages(query string) PassageResponse {
	apiUrl := "https://api.esv.org/v3/passage/text/?q=" + url.QueryEscape(query) + "&include-short-copyright=false&include-copyright=true"

	fmt.Printf("Getting passage from ESV: %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return PassageResponse{}
	}

	req.Header.Set("User-Agent", "AuraGem")
	req.Header.Set("Authorization", config.ESVAPIKey)
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return PassageResponse{}
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
		return PassageResponse{}
	}

	response := PassageResponse{}
	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return PassageResponse{}
	}

	return response
}
