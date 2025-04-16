package main

import (
	"fmt"
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
	/*hostsConfig2 := []sis.HostConfig{
	{BindAddress: "0.0.0.0", Hostname: "auragem.ddns.net", Upload: false, CertPath: "auragem.pem"},
	{BindAddress: "0.0.0.0", Hostname: "auragem.ddns.net", Upload: true, CertPath: "auragem.pem"},
	}*/
	geminiServer, _ := context.AddServer(sis.VirtualServer{Type: sis.ServerType_Gemini, Name: "aurarealm_gemini", DefaultLanguage: "en"}, hostsConfig...)
	geminiServer.AddRoute("/*", func(request *sis.Request) {
		fmt.Printf("Glob: %s\n", request.GlobString)
		if strings.HasPrefix(request.GlobString, "~") || strings.HasPrefix(request.GlobString, "/~") {
			request.GlobString = strings.TrimPrefix(strings.TrimPrefix(request.GlobString, "/"), "~") // TODO: Hack

			parts := strings.Split(request.GlobString, "/")
			request.GlobString = path.Join(parts[1:]...)
			if info, err := os.Stat(path.Join("/home/", parts[0], "gemini")); err == nil && info.IsDir() {
				request.ServeDirectory(filepath.Join("/home/", parts[0], "/gemini"))
				return
			} else {
				request.NotFound("Gemini directory not found.")
				return
			}
		} else if request.GlobString == "" {
			// Homepage - list all users
			request.Heading(1, "AuraRealm")
			request.PlainText("Welcome to AuraGem's pubnix, AuraRealm! You can find more information below.\n")
			request.Link("/about.gmi", "About AuraRealm")
			request.Heading(2, "Users")
			homeDirEntries, _ := os.ReadDir("/home/")
			for _, entry := range homeDirEntries {
				// Make sure user's directory has a gemini directory
				if info, err := os.Stat(path.Join("/home/", entry.Name(), "gemini")); err == nil && info.IsDir() {
					request.Link("/~"+entry.Name(), entry.Name())
				}
			}
			//request.ServeDirectory("/home/")
			return
		} else {
			request.ServeDirectory("/srv/gemini/")
			return
		}
	})

	context.Start()
}
