package gemini

import (
	// "io"

	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"gitlab.com/clseibold/auragem_sis/gemini/chat"
	"gitlab.com/clseibold/auragem_sis/gemini/music"
	"gitlab.com/clseibold/auragem_sis/gemini/search"
	"gitlab.com/clseibold/auragem_sis/gemini/textgame"
	"gitlab.com/clseibold/auragem_sis/gemini/textola"
	"gitlab.com/clseibold/auragem_sis/gemini/texts"
	youtube "gitlab.com/clseibold/auragem_sis/gemini/youtube"
	sis "gitlab.com/clseibold/smallnetinformationservices"
	// "gitlab.com/clseibold/auragem_sis/lifekept"
	/*"gitlab.com/clseibold/auragem_sis/ask"
	"gitlab.com/clseibold/auragem_sis/music"
	"gitlab.com/clseibold/auragem_sis/search"
	"gitlab.com/clseibold/auragem_sis/starwars"
	"gitlab.com/clseibold/auragem_sis/twitch"*/)

var GeminiCommand = &cobra.Command{
	Short: "Start SIS",
	Run:   RunServer,
}

func RunServer(cmd *cobra.Command, args []string) {
	//f, _ := os.Create("access.log")
	//gig.DefaultWriter = io.MultiWriter(f, os.Stdout)

	context, err := sis.InitSIS("./SIS/")
	context.AdminServer().BindAddress = "0.0.0.0"
	context.AdminServer().Hostname = "auragem.letz.dev"
	context.AdminServer().AddCertificate("auragem.pem")
	context.SaveConfiguration()
	context.GetPortListener("0.0.0.0", "1995").AddCertificate("auragem.letz.dev", "auragem.pem")
	if err != nil {
		panic(err)
	}

	// ----- AuraGem Servers -----

	// TODO: Pointers are not stable! Use a new ServerID struct with methods instead.
	geminiServer := context.AddServer(sis.Server{Type: sis.ServerType_Gemini, Name: "auragem_gemini", Hostname: "auragem.letz.dev"})
	geminiServer.AddCertificate("auragem.pem")
	context.GetPortListener("0.0.0.0", "1965").AddCertificate("auragem.letz.dev", "auragem.pem")

	geminiServer.AddDirectory("/*", "./")
	geminiServer.AddFile("/.well-known/security.txt", "./security.txt")
	geminiServer.AddProxyRoute("/nex/*", "$auragem_nex/*", '1')

	// Guestbook via Titan
	geminiServer.AddUploadRoute("/guestbook.gmi", handleGuestbook)

	handleDevlog(geminiServer)
	youtube.HandleYoutube(geminiServer)
	handleWeather(geminiServer)
	handleGithub(geminiServer)
	textgame.HandleTextGame(geminiServer)
	chat.HandleChat(geminiServer)
	textola.HandleTextola(geminiServer)
	music.HandleMusic(geminiServer)
	search.HandleSearchEngine(geminiServer)
	// twitch.HandleTwitch(geminiServer)
	// ask.HandleAsk(geminiServer)

	nexServer := context.AddServer(sis.Server{Type: sis.ServerType_Nex, Name: "auragem_nex", Hostname: "auragem.letz.dev"})
	nexServer.AddDirectory("/*", "./")
	nexServer.AddProxyRoute("/gemini/*", "$auragem_gemini/*", '1')
	nexServer.AddProxyRoute("/scholasticdiversity/*", "$scholasticdiversity_gemini/*", '1')

	// ----- Scholastic Diversity stuff -----
	scholasticdiversity_gemini := context.AddServer(sis.Server{Type: sis.ServerType_Gemini, Name: "scholasticdiversity_gemini", Hostname: "scholasticdiversity.us.to"})
	scholasticdiversity_gemini.AddCertificate("scholasticdiversity.pem")
	context.GetPortListener("0.0.0.0", "1965").AddCertificate("scholasticdiversity.us.to", "scholasticdiversity.pem")
	scholasticdiversity_gemini.AddDirectory("/*", "./")

	texts.HandleTexts(scholasticdiversity_gemini)
	// Add "/texts/" redirect from auragem gemini server to scholastic diversity gemini server
	geminiServer.AddRoute("/texts/*", func(request sis.Request) {
		unescaped, err := url.PathUnescape(request.GlobString)
		if err != nil {
			request.TemporaryFailure(err.Error())
			return
		}
		request.Redirect("gemini://scholasticdiversity.us.to/scriptures/%s", unescaped)
	})

	gopherServer := context.AddServer(sis.Server{Type: sis.ServerType_Gopher, Name: "gopher", Hostname: "auragem.letz.dev"})
	gopherServer.AddRoute("/", func(request sis.Request) {
		request.GophermapLine("i", "                             AuraGem Gopher Server", "/", "", "")
		request.GophermapLine("i", "", "/", "", "")
		request.GophermapLine("0", "About this server", "/about.txt", "", "")
		request.GophermapLine("i", "", "/", "", "")
		request.GophermapLine("7", "Search Geminispace", "/g/search/s/", "", "")
		request.GophermapLine("1", "Devlog", "/g/devlog/", "", "")
		request.GophermapLine("1", "Personal Log", "/g/~clseibold/", "", "")
		request.GophermapLine("0", "My Experience Within the Bitreich IRC", "/on_bitreich.txt", "", "")
		request.GophermapLine("0", "Freedom of Protocols Initiative", "/freedom_of_protocols_initiative.txt", "", "")
		request.GophermapLine("i", "", "/", "", "")

		request.GophermapLine("i", "Services/Info", "/", "", "")
		request.GophermapLine("1", "AuraGem Public Radio", "/g/music/public_radio/", "", "")
		request.GophermapLine("1", "Search Engine Homepage", "/g/search/", "", "")
		request.GophermapLine("1", "YouTube Proxy", "/g/youtube/", "", "")
		request.GophermapLine("1", "Scholastic Diversity", "/scholasticdiversity/", "", "")
		request.GophermapLine("i", "", "/", "", "")

		request.GophermapLine("i", "Software", "/", "", "")
		request.GophermapLine("1", "Misfin-Server", "/g/misfin-server/", "", "")
		request.GophermapLine("i", "", "/", "", "")

		request.GophermapLine("i", "Links", "/", "", "")
		request.GophermapLine("1", "Gopher Starting Point", "/iOS/gopher", "forthworks.com", "70")
		request.GophermapLine("7", "Search Via Veronica-2", "/v2/vs", "gopher.floodgap.com", "70")
		request.GophermapLine("7", "Search Via Quarry", "/quarry", "gopher.icu", "70")
		request.GophermapLine("1", "Gopherpedia", "/", "gopherpedia.com", "70")
		request.GophermapLine("1", "Bongusta Phlog Aggregator", "/bongusta", "i-logout.cz", "70")
		request.GophermapLine("1", "Moku Pona Phlog Aggregator", "/moku-pona", "gopher.black", "70")
		request.GophermapLine("1", "Mare Tranquillitatis People's Circumlunar Zaibatsu", "/", "zaibatsu.circumlunar.space", "70")
		request.GophermapLine("1", "Cosmic Voyage", "/", "cosmic.voyage", "70")
		request.GophermapLine("1", "Mozz.Us", "/", "mozz.us", "70")
		request.GophermapLine("1", "Quux", "/", "gopher.quux.org", "70")
		request.GophermapLine("1", "Mateusz' gophre lair", "/", "gopher.viste.fr", "70")
		request.GophermapLine("i", "", "/", "", "")

		request.GophermapLine("i", "Sister Sites", "/", "", "")
		request.GophermapLine("h", "AuraGem Gemini Server", "URL:gemini://auragem.letz.dev", "", "")
		request.GophermapLine("h", "AuraGem Nex Server", "URL:nex://auragem.letz.dev", "", "")
		request.GophermapLine("h", "Scholastic Diversity Gemini Server", "URL:gemini://scholasticdiversity.us.to", "", "")
		request.GophermapLine("i", "", "/", "", "")

		request.GophermapLine("i", "Ways to Contact Me:", "", "", "")
		request.GophermapLine("i", "IRC: ##misfin on libera.chat", "/", "", "")
		request.GophermapLine("h", "Email", "URL:mailto:christian.seibold32@outlook.com", "", "")
		request.GophermapLine("h", "Misfin Mail", "URL:misfin://clseibold@auragem.letz.dev", "", "")
		request.GophermapLine("i", "", "/", "", "")

		request.GophermapLine("i", "Powered By", "/", "", "")
		request.GophermapLine("i", "This server is powered by Smallnet Information Services (SIS):", "/", "", "")
		request.GophermapLine("h", "SIS Project", "URL:https://gitlab.com/clseibold/smallnetinformationservices/", "", "")
		request.GophermapLine("i", "Note that while SIS docs use the term \"proxying\" to describe requests of one server being handed off to another server of a different protocol, this is not proxying proper. The default document format of the protocol (index gemtext files, gophermaps, and Nex Listings) is translated when needed, but that and links are the only conversions that happen. This form of \"proxying\" all happens internally in the server software and *not* over the network or sockets. It is functionally equivalent to protocol proxying, but works slightly differently.", "/", "", "")
		request.GophermapLine("i", "", "/", "", "")
	})
	gopherServer.AddDirectory("/*", "./")
	gopherServer.AddProxyRoute("/g/*", "$auragem_gemini/*", '1')
	gopherServer.AddProxyRoute("/scholasticdiversity/*", "$scholasticdiversity_gemini/*", '1')

	spartanServer := context.AddServer(sis.Server{Type: sis.ServerType_Spartan, Name: "spartan", Hostname: "auragem.letz.dev"})
	spartanServer.AddRoute("/", func(request sis.Request) {
		request.Gemini(`# AuraGem Spartan Server

=: /g/search/s/ 🔍 Search
=> /g/devlog/ Devlog
=> /g/~clseibold/ Personal Log
=> /g/music/public_radio/ AuraGem Public Radio

## Software
=> /g/misfin-server/ Misfin-Server

## Other
=> /g/search/ Search Engine Homepage
=> /g/youtube/ YouTube Proxy

## Sister Sites
=> gemini://auragem.letz.dev/ AuraGem Gemini Server
=> nex://auragem.letz.dev/ AuraGem Nex Server
=> gemini://scholasticdiversity.us.to/ Scholastic Diversity

## Powered By
This server is powered by Smallnet Information Services (SIS):
=> https://gitlab.com/clseibold/smallnetinformationservices/ SIS Project
`)
	})
	spartanServer.AddProxyRoute("/g/*", "$auragem_gemini/*", '1')

	//gopherServer.AddProxyRoute("/devlog/*", "$auragem_gemini/devlog/*", '1')
	//gopherServer.AddProxyRoute("/~clseibold/*", "$auragem_gemini/~clseibold/*", '1')
	//gopherServer.AddProxyRoute("/misfin-server/*", "$auragem_gemini/misfin-server/*", '1')

	// Proxy public_radio stuff to gopherserver
	//gopherServer.AddProxyRoute("/public_radio/", "$auragem_gemini/music/public_radio/", '1')
	//gopherServer.AddProxyRoute("/public_radio/:station_name", "$auragem_gemini/music/public_radio/$station_name", '1')
	//gopherServer.AddProxyRoute("/public_radio/:station_name/schedule_feed", "$auragem_gemini/public_radio/$station_name/schedule_feed", '1')
	//gopherServer.AddProxyRoute("/stream/public_radio/:station_name", "$auragem_gemini/stream/public_radio/$station_name", 's')

	context.Start()
}

func handleGuestbook(request sis.Request) {
	guestbookPrefix := `# AuraGem Guestbook

This is the new guestbook! You can edit this page with the Titan protocol. Be sure to use the token "auragemguestbook" for upload.

Any changes with profane words or slurs will be fully rejected by the server.

Lastly, this guestbook is append-only. Any edits made to already-existing content will be rejected.

---
`
	if request.GetParam("token") != "auragemguestbook" {
		request.TemporaryFailure("A token is required.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "A token is required.")
	} else if request.DataSize > 5*1024*1024 { // 5 MB max size
		request.TemporaryFailure("Size too large.")
		return
	} else if request.DataMime != "text/plain" && request.DataMime != "text/gemini" {
		request.TemporaryFailure("Wrong mime type.")
		return
	}

	data, read_err := request.GetUploadData()
	if read_err != nil {
		return
	}
	if !utf8.ValidString(string(data)) {
		request.TemporaryFailure("Not a valid UTF-8 text file.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "Not a valid UTF-8 text file.")
	}
	if ContainsCensorWords(string(data)) {
		request.TemporaryFailure("Profanity or slurs were detected. Your edit is rejected.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "Profanity or slurs were detected. Your edit is rejected.")
	}
	if !strings.HasPrefix(string(data), guestbookPrefix) {
		request.TemporaryFailure("You edited the start of the document above \"---\". Your edit is rejected.")
		return
		//return c.NoContent(gig.StatusPermanentFailure, "You edited the start of the document above \"---\". Your edit is rejected.")
	}

	fileBefore, _ := request.Server.FS().ReadFile("guestbook.gmi")
	//fileBefore, _ := os.ReadFile(filepath.Join(request.Server.Directory, "gemini", "guestbook.gmi"))
	if !strings.HasPrefix(string(data), string(fileBefore)) {
		request.TemporaryFailure("You edited a portion of the document that already existed. Only appends are allowed. Your edit is rejected.")
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
	request.Redirect("gemini://%s:%s/guestbook.gmi", request.Server.Hostname(), request.Server.Port())
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
