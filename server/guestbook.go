package main

import (
	// "io"

	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
	/*"gitlab.com/clseibold/auragem_sis/twitch"*/)

func handleGuestbook(request *sis.Request) {
	guestbookPrefix := `# AuraGem Guestbook

This is the new guestbook! You can edit this page with the Titan protocol. Be sure to use the token "auragemguestbook" for upload.

Any changes with profane words or slurs will be fully rejected by the server.

Lastly, this guestbook is append-only. Any edits made to already-existing content will be rejected.

---
`
	if request.GetParam("token") != "auragemguestbook" {
		_ = request.TemporaryFailure("A token is required.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "A token is required.")
	} else if request.DataSize > 5*1024*1024 { // 5 MB max size
		_ = request.TemporaryFailure("Size too large.")
		return
	} else if request.DataMime != "text/plain" && request.DataMime != "text/gemini" {
		_ = request.TemporaryFailure("Wrong mime type.")
		return
	}

	data, read_err := request.GetUploadData()
	if read_err != nil {
		return
	}
	if !utf8.ValidString(string(data)) {
		_ = request.TemporaryFailure("Not a valid UTF-8 text file.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "Not a valid UTF-8 text file.")
	}
	if ContainsCensorWords(string(data)) {
		_ = request.TemporaryFailure("Profanity or slurs were detected. Your edit is rejected.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "Profanity or slurs were detected. Your edit is rejected.")
	}
	if !strings.HasPrefix(string(data), guestbookPrefix) {
		_ = request.TemporaryFailure("You edited the start of the document above \"---\". Your edit is rejected.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "You edited the start of the document above \"---\". Your edit is rejected.")
	}

	fileBefore, _ := request.Server.FS().ReadFile("guestbook.gmi")
	//fileBefore, _ := os.ReadFile(filepath.Join(request.Server.Directory, "gemini", "guestbook.gmi"))
	if !strings.HasPrefix(string(data), string(fileBefore)) {
		_ = request.TemporaryFailure("You edited a portion of the document that already existed. Only appends are allowed. Your edit is rejected.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "You edited a portion of the document that already existed. Only appends are allowed. Your edit is rejected.")
	}

	err := request.Server.FS().WriteFile("guestbook.gmi", data, 0600)
	//err := os.WriteFile(filepath.Join(request.Server.Directory, "gemini", "guestbook.gmi"), data, 0600)
	if err != nil {
		fmt.Printf("Write failed: %s\n", err.Error())
		return
		//return err
	}
	request.Redirect("gemini://%s:%s/guestbook.gmi", request.Host.Hostname, request.Host.Port)
}

func CensorWords(str string) string {
	wordCensors := []string{"fuck", "kill", "die", "damn", "ass", "shit", "stupid", "faggot", "fag", "whore", "cock", "cunt", "motherfucker", "fucker", "asshole", "nigger", "abbie", "abe", "abie", "abid", "abeed", "ape", "armo", "nazi", "ashke-nazi", "אשכנאצי", "bamboula", "barbarian", "beaney", "beaner", "bohunk", "boerehater", "boer-hater", "burrhead", "burr-head", "chode", "chad", "penis", "vagina", "porn", "bbc", "stealthing", "bbw", "Hentai", "milf", "dilf", "tummysticks", "heeb", "hymie", "kike", "jidan", "sheeny", "shylock", "zhyd", "yid", "shyster", "smouch"}

	var result = str
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
