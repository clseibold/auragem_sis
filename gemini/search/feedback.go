package search

import (
	"strings"
	"unicode"
	"unicode/utf8"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

func handleSearchFeedback(s sis.ServerHandle) {
	s.AddRoute("/search/feedback.gmi", searchFeedbackPage)
	s.AddUploadRoute("/search/feedback.gmi", searchFeedbackPage)
}

func searchFeedbackPage(request *sis.Request) {
	token := "auragemsearchfeedback"
	/*feedbackPrefix := `# AuraGem Search Feedback

	This is the new AuraGem Search Feedback page! You can edit this page with the Titan protocol. Be sure to use the titan token "auragemsearchfeedback" for upload. If you do not have a browser that supports Titan, I suggest Lagrange, which provides an option to "Edit Page with Titan..." within the right-click menu.

	This page is used to provide feedback on the search engine or search results. If providing feedback on search results, please include the search query that you are commenting on.

	Any changes with profane words or slurs will be fully rejected by the server. Lastly, this page is append-only. Any edits made to already-existing content will be rejected.

	---
	`*/

	if request.Upload {
		if request.GetParam("token") != token {
			request.TemporaryFailure("A token is required.")
			return
		}

		if request.DataSize > 5*1024*1024 {
			request.TemporaryFailure("Size too large.")
			return
		}

		// TODO: Check that the mimetype is gemini or text file

		data, err := request.GetUploadData()
		if err != nil {
			return //err
		}

		if !utf8.ValidString(string(data)) {
			request.TemporaryFailure("Not a valid UTF-8 text file.")
			return
		}
		if ContainsCensorWords(string(data)) {
			request.TemporaryFailure("Profanity or slurs were detected. Your edit is rejected.")
			return
		}
		/*if !strings.HasPrefix(string(data), feedbackPrefix) {
					request.TemporaryFailure("You edited the start of the document above \"---\". Your edit is rejected.")
		return
				}*/

		fileBefore, _ := request.Server.FS().ReadFile("searchfeedback.gmi")
		//fileBefore, _ := os.ReadFile("searchfeedback.gmi")
		if !strings.HasPrefix(string(data), string(fileBefore)) {
			request.TemporaryFailure("You edited a portion of the document that already existed. Only appends are allowed. Your edit is rejected.")
			return
		}

		err = request.Server.FS().WriteFile("searchfeedback.gmi", data, 0600)
		if err != nil {
			return //err
		}
		request.Redirect("%s%s:%s/search/feedback.gmi", request.Server.Scheme(), request.Host.Hostname, request.Host.Port)
		return
	} else {
		fileData, err := request.Server.FS().ReadFile("searchfeedback.gmi")
		if err != nil {
			request.TemporaryFailure(err.Error())
			return
		}
		//request.FileMimetype("text/gemini", "auragem_gemini/searchfeedback.gmi")
		request.TextWithMimetype("text/gemini", string(fileData))
		return
	}
}

func CensorWords(str string) string {
	wordCensors := []string{"fuck", "kill", "die", "damn", "ass", "shit", "stupid", "faggot", "fag", "whore", "cock", "cunt", "motherfucker", "fucker", "asshole", "nigger", "abbie", "abe", "abie", "abid", "abeed", "ape", "armo", "nazi", "ashke-nazi", "אשכנאצי", "bamboula", "barbarian", "beaney", "beaner", "bohunk", "boerehater", "boer-hater", "burrhead", "burr-head", "chode", "chad", "penis", "vagina", "porn", "bbc", "stealthing", "bbw", "Hentai", "milf", "dilf", "tummysticks", "heeb", "hymie", "kike", "jidan", "sheeny", "shylock", "zhyd", "yid", "shyster", "smouch"}

	var result string = str
	for _, forbiddenWord := range wordCensors {
		replacement := strings.Repeat("*", len(forbiddenWord))
		result = strings.Replace(result, forbiddenWord, replacement, -1)
	}

	return result
}

func ContainsCensorWords(str string) bool {
	wordCensors := map[string]bool{"fuck": true, "f*ck": true, "kill": true, "k*ll": true, "die": true, "damn": true, "ass": true, "*ss": true, "shit": true, "sh*t": true, "stupid": true, "faggot": true, "fag": true, "f*g": true, "whore": true, "wh*re": true, "cock": true, "c*ck": true, "cunt": true, "c*nt": true, "motherfucker": true, "fucker": true, "f*cker": true, "asshole": true, "*sshole": true, "nigger": true, "n*gger": true, "n*gg*r": true, "abbie": true, "abe": true, "abie": true, "abid": true, "abeed": true, "ape": true, "armo": true, "nazi": true, "ashke-nazi": true, "אשכנאצי": true, "bamboula": true, "barbarian": true, "beaney": true, "beaner": true, "bohunk": true, "boerehater": true, "boer-hater": true, "burrhead": true, "burr-head": true, "chode": true, "chad": true, "penis": true, "vagina": true, "porn": true, "stealthing": true, "bbw": true, "Hentai": true, "milf": true, "dilf": true, "tummysticks": true, "heeb": true, "hymie": true, "kike": true, "k*ke": true, "jidan": true, "sheeny": true, "shylock": true, "zhyd": true, "yid": true, "shyster": true, "smouch": true}

	fields := strings.FieldsFunc(strings.ToLower(str), func(r rune) bool {
		if r == '*' {
			return false
		}
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsDigit(r) || !unicode.IsPrint(r)
	})

	for _, word := range fields {
		if _, ok := wordCensors[word]; ok {
			return true
		}
	}

	return false
}
