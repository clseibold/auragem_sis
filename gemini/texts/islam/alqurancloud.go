package islam

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

type SurahsWrapper struct {
	Code int `json:"code"`
	Status string `json:"status"`
	Data []Surah `json:"data"`
}
type SurahWrapper struct {
	Code int `json:"code"`
	Status string `json:"status"`
	Data Surah `json:"data"`
}
type Surah struct {
	Number int `json:"number"`
	Name string `json:"name"`
	EnglishName string `json:"englishName"`
	EnglishNameTranslation string `json:"englishNameTranslation"`
	NumberOfAyahs int `json:"numberOfAyahs"`
	RevelationType string `json:"revelationType"`

	Ayahs []Ayah `json:"ayahs"`
	Edition SurahEdition `json:"edition"`
}
type Ayah struct {
	Number int `json:"number"`
	Text string `json:"text"`
	NumberInSurah int `json:"numberInSurah"`
	Juz int `json:"juz"`
	Manzil int `json:"manzil"`
	Page int `json:"page"`
	Ruku int `json:"ruku"`
	HizbQuarter int `json:"hizbQuarter"`
	Sajda json.RawMessage `json:"sajda"` // TODO: Can be bool or object
}
type SurahEdition struct {
	Identifier string `json:"identifier"`
	Language string `json:"language"`
	Name string `json:"name"`
	EnglishName string `json:"englishName"`
	Format string `json:"format"`
	Type string `json:"type"`
	Direction string `json:"direction"`
}

func GetSurahs() []Surah {
	apiUrl := "https://api.alquran.cloud/v1/surah"

	fmt.Printf("Getting surahs from %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return []Surah {}
	}

	req.Header.Set("User-Agent", "AuraGem")
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return []Surah {}
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
		return []Surah {}
	}

	quranSurahs := SurahsWrapper {}
	jsonErr := json.Unmarshal(body, &quranSurahs)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return []Surah {}
	}

	return quranSurahs.Data
}

func GetSurah(versionId string, number string) Surah {
	apiUrl := "https://api.alquran.cloud/v1/surah/" + number
	if versionId != "arabic" {
		apiUrl += "/" + versionId
	}

	fmt.Printf("Getting surah from %s\n", apiUrl)

	spaceClient := http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, err)
		return Surah {}
	}

	req.Header.Set("User-Agent", "AuraGem")
	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, getErr)
		return Surah {}
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
		return Surah {}
	}

	quranSurah := SurahWrapper {}
	jsonErr := json.Unmarshal(body, &quranSurah)
	if jsonErr != nil {
		_, filename, line, _ := runtime.Caller(0)
		fmt.Printf("%s:%d %s\n", filename, line, jsonErr)
		return Surah {}
	}

	return quranSurah.Data
}
