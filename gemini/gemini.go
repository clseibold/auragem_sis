package gemini

import (
	// "io"

	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"gitlab.com/clseibold/auragem_sis/gemini/ask"
	"gitlab.com/clseibold/auragem_sis/gemini/chat"
	"gitlab.com/clseibold/auragem_sis/gemini/music"
	"gitlab.com/clseibold/auragem_sis/gemini/search"
	"gitlab.com/clseibold/auragem_sis/gemini/starwars"
	"gitlab.com/clseibold/auragem_sis/gemini/textgame"
	"gitlab.com/clseibold/auragem_sis/gemini/textola"
	"gitlab.com/clseibold/auragem_sis/gemini/texts"
	"gitlab.com/clseibold/auragem_sis/gemini/youtube"
	"gitlab.com/sis-suite/aurarepo"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
	/*"gitlab.com/clseibold/auragem_sis/twitch"*/)

var Command = &cobra.Command{
	Short: "Start SIS",
	Run:   RunServer,
}

var OnionId string = ""

func RunServer(cmd *cobra.Command, args []string) {
	context, err := sis.InitSIS("./SIS/")
	if err != nil {
		panic(err)
	}
	err = context.SaveConfiguration()
	if err != nil {
		panic(err)
	}

	chatContext := chat.NewChatContext()

	// Setup AuraRepo Server
	aurarepoContext := aurarepo.NewAuraRepoContext("AuraRepo", "/home/clseibold/repos", "https://auragem.ddns.net/~aurarepo")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "smallnetinformationservices", "Smallnet Information Services", "./smallnetinformationservices/", "Server software suite for smallnet internet ecosystem, managed with a Gemini admin dashboard.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "aurarepo", "AuraRepo", "./aurarepo/", "A Git repository hosting forge SCGI application server for the Gemini Protocol, built using Smallnet Information Services and go-git.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "auramuse-lite", "AuraMuse Lite", "./auramuse-lite/", "An SCGI application server that provides radio over Gemini.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "auragem", "AuraGem Servers", "./auragem/", "The code for AuraGem and related servers.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Prism, "prism", "Prism VCS", "./prism/", "Distributed version control through a clearer lens. A new VCS that builds on the good ideas of git, fossil, subversion, and mercurial, while simplifying the UI, and introducing unique concepts to Distributed VCSs.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "linux-mirror", "Linux Stable Mirror", "./linux-mirror/", "Mirror of kernel.org's kernel/git/stable/linux.git")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "git-mirror", "Git Mirror", "./git-mirror/", "Mirror of kernel.org's git.git")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "aocl", "AOCL", "./aocl/", "AuraGem Opensource Copyleft License")

	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "deconflicter", "Deconflicter", "./deconflicter/", "VCS conflicts terminal viewer, written in C")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "derevel", "Derevel", "./derevel/", "A build system using scripts that are written in C")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "edimcoder", "EdimCoder", "./edimcoder/", "A terminal line text-editor, in a similar vein to Ed, written in C.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "paled", "Paled", "./paled/", "Linux shell written in C.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "textadventure", "TextAdventure", "./textadventure/", "An unfinished sandbox text-based game written in C.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "multilect", "Multilect", "./multilect/", "A terminal markdown file reader written in C.")
	aurarepoContext.AddRepo(aurarepo.AuraRepoType_Git, "ncure", "Ncure", "./ncure/", "A cross-platform terminal manipulation library written in C.")

	// Setup BinTree server
	bintreeContext := aurarepo.NewAuraRepoContext("BinTree", "/home/clseibold/bintree-repos", "")
	bintreeContext.AddRepo(aurarepo.AuraRepoType_Git, "profectus.bin", "Profectus", "./profectus.bin/", "Smallnet GUI browser.")
	bintreeContext.AddRepo(aurarepo.AuraRepoType_Git, "golang.bin", "Golang", "./golang.bin/", "The Go Programming Language")

	setupWebServer(aurarepoContext, bintreeContext)
	go startTorOnlyWebServer()
	setupTorOnly(context)

	setupAuraGem(context, chatContext, aurarepoContext, bintreeContext)
	setupScholasticDiversity(context)
	setupScrollProtocol(context)
	setupNewsfin(context)

	context.Start()
}

func setupTor() {
	/*t, err := tor.Start(nil, &tor.StartConf{DataDir: "./TorData/", EnableNetwork: true})
	if err != nil {
		fmt.Printf("Failed to start tor: %v", err)
	}*/

	//go startTorWebServer(t)
}

// Tor - 9400
/*
func startTorWebServer(t *tor.Tor) {
	// Wait at most a few minutes to publish the service
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Create an onion v3 service to listen on any port but show as 80
	onion, err := t.Listen(ctx, &tor.ListenConf{RemotePorts: []int{80}, Version3: true})
	if err != nil {
		log.Panicf("Failed to create onion service: %v", err)
	}
	defer onion.Close()
	OnionId = onion.ID

	fmt.Printf("Onion Service: %v.onion\n", onion.ID)

	//http.Handle("/*", )
	http.Handle("/", http.FileServer(http.Dir("/home/clseibold/ServerData/auragem_sis/SIS/auragem_tor_http")))
	http.Handle("/scrollprotocol/", http.FileServer(http.Dir("/home/clseibold/ServerData/auragem_sis/SIS/scrollprotocol_http")))
	http.Serve(onion, nil)
}
*/

func setupWebServer(aurarepoContext *aurarepo.AuraRepoContext, bintreeContext *aurarepo.AuraRepoContext) {
	httpMuxer := http.NewServeMux()
	httpMuxer.Handle("/", http.FileServer(http.Dir("/home/clseibold/ServerData/auragem_sis/SIS/auragem_http")))
	httpMuxer.Handle("scrollprotocol.us.to/", http.FileServer(http.Dir("/home/clseibold/ServerData/auragem_sis/SIS/scrollprotocol_http")))
	httpMuxer.Handle("auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion/", http.FileServer(http.Dir("/home/clseibold/ServerData/auragem_sis/SIS/auragem_tor_http")))

	aurarepoMuxer := http.NewServeMux()
	aurarepoContext.AttachHTTPSmart(aurarepoMuxer)
	httpMuxer.Handle("/~aurarepo/", http.StripPrefix("/~aurarepo", aurarepoMuxer))

	bintreeMuxer := http.NewServeMux()
	bintreeContext.AttachHTTPSmart(bintreeMuxer)
	httpMuxer.Handle("/~bintree/", http.StripPrefix("/~bintree", bintreeMuxer))

	go func() {
		err := http.ListenAndServe("0.0.0.0:80", httpMuxer)
		if err != nil {
			fmt.Printf("Failed to start web server on 0.0.0.0:80. %s\n", err.Error())
		}
	}()
	go func() {
		err2 := http.ListenAndServeTLS("0.0.0.0:443", "/etc/letsencrypt/live/auragem.ddns.net/fullchain.pem", "/etc/letsencrypt/live/auragem.ddns.net/privkey.pem", httpMuxer)
		if err2 != nil {
			fmt.Printf("Failed to start web server on 0.0.0.0:443. %s\n", err2.Error())
		}
	}()
}

// Tor-only server
func startTorOnlyWebServer() {
	err := http.ListenAndServe("0.0.0.0:8080", http.FileServer(http.Dir("/home/clseibold/ServerData/auragem_sis/SIS/varilib_http")))
	if err != nil {
		fmt.Printf("Failed to start web server on 0.0.0.0:80. %s\n", err.Error())
	}
}

func setupTorOnly(context *sis.SISContext) {
	// geminiServer := context.AddServer(sis.Server{Type: sis.ServerType_Gemini, Name: "varilib_gemini", DefaultLanguage: "en"}, sis.HostConfig{BindAddress: "0.0.0.0", Hostname: "varilibcoo4yblrqhx43y5kvryy6htzoaa2vcmguxer4yti2r4ffyfyd.onion", Upload: false, CertPath: "varilib.pem"}, sis.HostConfig{BindAddress: "0.0.0.0", Hostname: "varilibcoo4yblrqhx43y5kvryy6htzoaa2vcmguxer4yti2r4ffyfyd.onion", Upload: true, CertPath: "varilib.pem"})
	//context.GetPortListener("0.0.0.0", "1965").AddCertificate("varilibcoo4yblrqhx43y5kvryy6htzoaa2vcmguxer4yti2r4ffyfyd.onion", "varilib.pem")
	// geminiServer.AddDirectory("/*", "./")

	// scrollServer := context.AddServer(sis.Server{Type: sis.ServerType_Scroll, Name: "varilib_scroll", DefaultLanguage: "en"}, sis.HostConfig{BindAddress: "0.0.0.0", Hostname: "varilibcoo4yblrqhx43y5kvryy6htzoaa2vcmguxer4yti2r4ffyfyd.onion", Upload: false, CertPath: "varilib.pem"}, sis.HostConfig{BindAddress: "0.0.0.0", Hostname: "varilibcoo4yblrqhx43y5kvryy6htzoaa2vcmguxer4yti2r4ffyfyd.onion", Upload: true, CertPath: "varilib.pem"})
	//context.GetPortListener("0.0.0.0", "5699").AddCertificate("varilibcoo4yblrqhx43y5kvryy6htzoaa2vcmguxer4yti2r4ffyfyd.onion", "varilib.pem")
	// scrollServer.AddProxyRoute("/*", "$varilib_gemini/*", '1')

	// spartanServer := context.AddServer(sis.Server{Type: sis.ServerType_Spartan, Name: "varilib_spartan", DefaultLanguage: "en"}, sis.HostConfig{BindAddress: "0.0.0.0", Hostname: "varilibcoo4yblrqhx43y5kvryy6htzoaa2vcmguxer4yti2r4ffyfyd.onion", Upload: false, CertPath: "varilib.pem"}, sis.HostConfig{BindAddress: "0.0.0.0", Hostname: "varilibcoo4yblrqhx43y5kvryy6htzoaa2vcmguxer4yti2r4ffyfyd.onion", Upload: true, CertPath: "varilib.pem"})
	// spartanServer.AddProxyRoute("/*", "$varilib_gemini/*", '1')
}

func setupAuraGem(context *sis.SISContext, chatContext *chat.ChatContext, aurarepoContext *aurarepo.AuraRepoContext, bintreeContext *aurarepo.AuraRepoContext) {
	hostsConfig := []sis.HostConfig{
		{BindAddress: "0.0.0.0", Hostname: "auragem.ddns.net", Upload: false, CertPath: "auragem.pem"},
		{BindAddress: "0.0.0.0", Hostname: "auragem.ddns.net", Upload: true, CertPath: "auragem.pem"},
		{BindAddress: "0.0.0.0", Hostname: "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", Upload: false, CertPath: "auragem.pem"},
		{BindAddress: "0.0.0.0", Hostname: "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", Upload: true, CertPath: "auragem.pem"},
	}
	hostsConfig2 := []sis.HostConfig{
		{BindAddress: "0.0.0.0", Hostname: "auragem.ddns.net", Upload: false, CertPath: "auragem.pem"},
		{BindAddress: "0.0.0.0", Hostname: "auragem.ddns.net", Upload: true, CertPath: "auragem.pem"},
	}
	geminiServer, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Gemini, Name: "auragem_gemini", DefaultLanguage: "en"}, hostsConfig...)
	// context.GetPortListener("0.0.0.0", "1965").AddCertificate("auragem.ddns.net", "auragem.pem")

	// Add Onion Address handling for server
	// context.AddServerRoute("0.0.0.0", "1965", sis.ProtocolType_Gemini, "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", geminiServer)
	// context.GetPortListener("0.0.0.0", "1965").AddCertificate("auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", "auragem.pem")

	geminiServer.AddDirectory("/*", "./")
	geminiServer.AddFile("/.well-known/security.txt", "./security.txt")
	geminiServer.AddProxyRoute("/nex/*", "$auragem_nex/*", '1')
	geminiServer.AddUploadRoute("/guestbook.gmi", handleGuestbook)
	geminiServer.AddCGIRoute("/debug/*", "/home/clseibold/ServerData/cgi_test")

	// Proxies
	youtube.HandleYoutube(geminiServer)
	handleGithub(geminiServer)
	// twitch.HandleTwitch(geminiServer)

	// Services
	handleDevlog(geminiServer)
	handleWeather(geminiServer)

	chatContext.Attach(geminiServer)
	aurarepoContext.Attach(geminiServer.Group("/~aurarepo/"))
	bintreeContext.Attach(geminiServer.Group("/~bintree/"))

	textgame.HandleTextGame(geminiServer)
	textola.HandleTextola(geminiServer)
	music.HandleMusic(geminiServer)
	search.HandleSearchEngine(geminiServer)
	starwars.HandleStarWars(geminiServer)
	ask.HandleAsk(geminiServer)

	// Add "/texts/" redirect from AuraGem gemini server to scholastic diversity gemini server
	geminiServer.AddRoute("/texts/*", func(request *sis.Request) {
		unescaped, err := url.PathUnescape(request.GlobString)
		if err != nil {
			_ = request.TemporaryFailure("%s", err.Error())
			return
		}
		request.Redirect("gemini://scholasticdiversity.us.to/scriptures/%s", unescaped)
	})

	scrollServer, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Scroll, Name: "auragem_scroll", DefaultLanguage: "en"}, hostsConfig...)
	// context.GetPortListener("0.0.0.0", "5699").AddCertificate("auragem.ddns.net", "auragem.pem")
	// context.AddServerRoute("0.0.0.0", "5699", sis.ProtocolType_Scroll, "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", scrollServer)
	// context.GetPortListener("0.0.0.0", "5699").AddCertificate("auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", "auragem.pem")
	scrollServer.AddProxyRoute("/*", "$auragem_gemini/*", '1')

	nexServer, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Nex, Name: "auragem_nex", DefaultLanguage: "en"}, hostsConfig2...)
	//context.AddServerRoute("0.0.0.0", "1900", sis.ProtocolType_Nex, "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", nexServer)
	nexServer.AddDirectory("/*", "./")
	nexServer.AddProxyRoute("/gemini/*", "$auragem_gemini/*", '1')
	nexServer.AddProxyRoute("/scholasticdiversity/*", "$scholasticdiversity_gemini/*", '1')
	nexServer.AddProxyRoute("/scrollprotocol/*", "$scrollprotocol_gemini/*", '1')

	spartanServer, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Spartan, Name: "spartan", DefaultLanguage: "en"}, hostsConfig...)
	// context.AddServerRoute("0.0.0.0", "300", sis.ProtocolType_Spartan, "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", spartanServer)
	spartanServer.AddFile("/", "./index.gmi")
	spartanServer.AddProxyRoute("/*", "$auragem_gemini/*", '1')

	gopherServer, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Gopher, Name: "gopher", DefaultLanguage: "en"}, hostsConfig2...)
	//context.AddServerRoute("0.0.0.0", "70", sis.ProtocolType_Gopher, "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", gopherServer)
	gopherServer.AddDirectory("/*", "./")
	gopherServer.AddProxyRoute("/g/*", "$auragem_gemini/*", '1')
	gopherServer.AddProxyRoute("/scholasticdiversity/*", "$scholasticdiversity_gemini/*", '1')
	gopherServer.AddProxyRoute("/scrollprotocol/*", "$scrollprotocol_scroll/*", '1')
}

func setupScholasticDiversity(context *sis.SISContext) {
	hostsConfig := []sis.HostConfig{
		{BindAddress: "0.0.0.0", Hostname: "scholasticdiversity.us.to", Upload: false, CertPath: "scholasticdiversity.pem"},
		{BindAddress: "0.0.0.0", Hostname: "scholasticdiversity.us.to", Upload: true, CertPath: "scholasticdiversity.pem"},
	}
	scholasticdiversity_gemini, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Gemini, Name: "scholasticdiversity_gemini", DefaultLanguage: "en"}, hostsConfig...)
	// context.GetPortListener("0.0.0.0", "1965").AddCertificate("scholasticdiversity.us.to", "scholasticdiversity.pem")
	scholasticdiversity_gemini.AddDirectory("/*", "./")

	texts.HandleTexts(scholasticdiversity_gemini)

	scholasticdiversity_scroll, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Scroll, Name: "scholasticdiversity_scroll", DefaultLanguage: "en"}, hostsConfig...)
	// context.GetPortListener("0.0.0.0", "5699").AddCertificate("scholasticdiversity.us.to", "scholasticdiversity.pem")
	scholasticdiversity_scroll.AddProxyRoute("/*", "$scholasticdiversity_gemini/*", '1')
}

func setupScrollProtocol(context *sis.SISContext) {
	hostsConfig := []sis.HostConfig{
		{BindAddress: "0.0.0.0", Hostname: "scrollprotocol.us.to", Upload: false, CertPath: "scrollprotocol.pem"},
		{BindAddress: "0.0.0.0", Hostname: "scrollprotocol.us.to", Upload: true, CertPath: "scrollprotocol.pem"},
	}
	scrollProtocol_scroll, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Scroll, Name: "scrollprotocol_scroll", DefaultLanguage: "en"}, hostsConfig...)
	// context.GetPortListener("0.0.0.0", "5699").AddCertificate("scrollprotocol.us.to", "scrollprotocol.pem")
	scrollProtocol_scroll.AddDirectory("/*", "./")

	scrollProtocol_gemini, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Gemini, Name: "scrollprotocol_gemini", DefaultLanguage: "en"}, hostsConfig...)
	// context.GetPortListener("0.0.0.0", "1965").AddCertificate("scrollprotocol.us.to", "scrollprotocol.pem")
	scrollProtocol_gemini.AddProxyRoute("/*", "$scrollprotocol_scroll/*", '1')
}

func setupNewsfin(context *sis.SISContext) {
	hostsConfig := []sis.HostConfig{
		{BindAddress: "0.0.0.0", Hostname: "newsfin.us.to", Upload: false, CertPath: "newsfin.pem"},
		{BindAddress: "0.0.0.0", Hostname: "newsfin.us.to", Upload: true, CertPath: "newsfin.pem"},
	}

	/*scrollProtocol_scroll := context.AddServer(sis.Server{Type: sis.ServerType_Scroll, Name: "scrollprotocol_scroll", Hostname: "scrollprotocol.us.to", DefaultLanguage: "en"})
	context.GetPortListener("0.0.0.0", "5699").AddCertificate("scrollprotocol.us.to", "scrollprotocol.pem")
	scrollProtocol_scroll.AddDirectory("/*", "./")*/

	newsfin_gemini, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Gemini, Name: "newsfin_gemini", DefaultLanguage: "en"}, hostsConfig...)
	// context.GetPortListener("0.0.0.0", "1965").AddCertificate("newsfin.us.to", "newsfin.pem")
	newsfin_gemini.AddDirectory("/*", "./")
}

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
