package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/spf13/cobra"
	"gitlab.com/clseibold/auragem_sis/migration"
	_ "gitlab.com/clseibold/auragem_sis/migration"
	"gitlab.com/clseibold/auragem_sis/server/ask"
	"gitlab.com/clseibold/auragem_sis/server/chat"
	"gitlab.com/clseibold/auragem_sis/server/music"
	"gitlab.com/clseibold/auragem_sis/server/search"
	"gitlab.com/clseibold/auragem_sis/server/starwars"
	"gitlab.com/clseibold/auragem_sis/server/textgame"
	"gitlab.com/clseibold/auragem_sis/server/textgame2"
	"gitlab.com/clseibold/auragem_sis/server/textola"
	"gitlab.com/clseibold/auragem_sis/server/texts"
	"gitlab.com/clseibold/auragem_sis/server/youtube"
	"gitlab.com/sis-suite/aurarepo"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

const cpuprofile = "./cpu.prof"
const memprofile = "./mem.prof"

var Command = &cobra.Command{
	Short: "Start SIS",
	Run:   RunServer,
}

func init() {
	migration.InitMigrationCommands(Command)
}

func main() {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	err := Command.Execute()
	if err != nil {
		panic(err)
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

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

var OnionId string = ""

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
	geminiServer.AddSCGIRoute("/realm/*", "scgi://10.42.0.2:7000")

	textGame2Context := textgame2.NewContext()
	textGame2Context.Attach(geminiServer.Group("/textgame2/"))
	textGame2Context.Start()

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
	nexServer.AddSCGIRoute("/realm/*", "scgi://10.42.0.2:7000")

	spartanServer, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Spartan, Name: "spartan", DefaultLanguage: "en"}, hostsConfig...)
	// context.AddServerRoute("0.0.0.0", "300", sis.ProtocolType_Spartan, "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", spartanServer)
	spartanServer.AddFile("/", "./index.gmi")
	spartanServer.AddProxyRoute("/*", "$auragem_gemini/*", '1')
	spartanServer.AddSCGIRoute("/realm/*", "scgi://10.42.0.2:7001")

	gopherServer, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Gopher, Name: "gopher", DefaultLanguage: "en"}, hostsConfig2...)
	//context.AddServerRoute("0.0.0.0", "70", sis.ProtocolType_Gopher, "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", gopherServer)
	gopherServer.AddDirectory("/*", "./")
	gopherServer.AddProxyRoute("/g/*", "$auragem_gemini/*", '1')
	gopherServer.AddProxyRoute("/scholasticdiversity/*", "$scholasticdiversity_gemini/*", '1')
	gopherServer.AddProxyRoute("/scrollprotocol/*", "$scrollprotocol_scroll/*", '1')
	gopherServer.AddSCGIRoute("/realm/*", "scgi://10.42.0.2:7002")
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
