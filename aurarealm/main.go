package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

func main() {
	context, err := sis.InitConfiglessMode() // TODO
	if err != nil {
		panic(err)
	}
	err = context.SaveConfiguration()
	if err != nil {
		panic(err)
	}

	hostsConfig := []sis.HostConfig{
		{BindAddress: "0.0.0.0", BindPort: "7000", Hostname: "auragem.ddns.net", Port: "1965", Upload: false, SCGI: true},
		{BindAddress: "0.0.0.0", BindPort: "7000", Hostname: "auragem.ddns.net", Port: "1965", Upload: true, SCGI: true},
		{BindAddress: "0.0.0.0", BindPort: "7000", Hostname: "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", Port: "1965", Upload: false, SCGI: true},
		{BindAddress: "0.0.0.0", BindPort: "7000", Hostname: "auragemhkzsr5rowsaxauti6yhinsaa43wjtcqxhh7fw5tijdoqbreyd.onion", Port: "1965", Upload: true, SCGI: true},
	}
	geminiServer, _ := context.CreateServer(sis.ServerType_Gemini, "aurarealm_gemini", "en", hostsConfig...)
	geminiServer.AddRoute("/*", crossProtocolHandler)

	spartanServer, _ := context.CreateServer(sis.ServerType_Spartan, "aurarealm_spartan", "en", hostsConfig...)
	spartanServer.AddRoute("/*", crossProtocolHandler)
	spartanServer.AddProxyRoute("/about.gmi", "$aurarealm_gemini/about.gmi", '0')

	gopherServer, _ := context.CreateServer(sis.ServerType_Gopher, "aurarealm_gopher", "en", hostsConfig...)
	gopherServer.AddRoute("/*", crossProtocolHandler)
	gopherServer.AddProxyRoute("/about.txt", "$aurarealm_gemini/about.gmi", '0')

	internalHostsConfig := []sis.HostConfig{
		{BindAddress: "localhost", BindPort: "1965", Hostname: "localhost", Port: "1965", Upload: false, SCGI: false, CertPath: "internal.pem"},
		{BindAddress: "localhost", BindPort: "1965", Hostname: "localhost", Port: "1965", Upload: true, SCGI: false, CertPath: "internal.pem"},
	}
	internalGeminiServer, _ := context.CreateServer(sis.ServerType_Gemini, "aurarealm_internal_gemini", "en", internalHostsConfig...)
	internalGeminiServer.AddDirectory("/*", "/srv/internal")
	internalGeminiServer.AddProxyRoute("/about.gmi", "$aurarealm_gemini/about.gmi", '0')
	internalGeminiServer.AddProxyRoute("/public/*", "$aurarealm_gemini/*", '1')

	context.Start()
}

func crossProtocolHandler(request *sis.Request) {
	subDirectoryName := "gemini"
	aboutPath := "/about.gmi"
	switch request.Type {
	case sis.ProtocolType_Gemini:
		subDirectoryName = "gemini"
		aboutPath = "/about.gmi"
	case sis.ProtocolType_Spartan:
		subDirectoryName = "spartan"
		aboutPath = "/about.gmi"
	case sis.ProtocolType_Gopher:
		subDirectoryName = "gopher"
		aboutPath = "/about.txt"
	}

	if strings.HasPrefix(request.GlobString, "~") || strings.HasPrefix(request.GlobString, "/~") {
		request.GlobString = strings.TrimPrefix(strings.TrimPrefix(request.GlobString, "/"), "~") // TODO: Hack

		// Check that user has directory for given protocol and serve it.
		parts := strings.Split(request.GlobString, "/")
		request.GlobString = path.Join(parts[1:]...)
		if info, err := os.Stat(path.Join("/home/", parts[0], subDirectoryName)); err == nil && info.IsDir() {
			request.ServeDirectory(filepath.Join("/home/", parts[0], subDirectoryName))
			return
		} else {
			request.NotFound("Gemini directory not found.")
			return
		}
	} else if request.GlobString == "" {
		// Homepage - list all users
		request.Heading(1, "AuraRealm")
		request.PlainText("Welcome to AuraGem's pubnix, AuraRealm! You can find more information below.\n")
		request.Link(aboutPath, "About AuraRealm")
		request.Heading(2, "Users")
		homeDirEntries, _ := os.ReadDir("/home/")
		for _, entry := range homeDirEntries {
			// Make sure user's directory has a directory for the given protocol
			if info, err := os.Stat(path.Join("/home/", entry.Name(), subDirectoryName)); err == nil && info.IsDir() {
				request.Link("/~"+entry.Name(), entry.Name())
			}
		}
		return
	} else if strings.HasPrefix(request.GlobString, aboutPath[1:]) {
		// TODO: Always assume about.gmi is available and just convert it to the desired format for given protocol?
		//request.EnableConvertMode()
		request.File(filepath.Join("/srv/gemini/", "about.gmi"))
	} else {
		request.ServeDirectory(filepath.Join("/srv/", subDirectoryName))
		return
	}
}
