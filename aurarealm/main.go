package main

import (
	"os"
	"path"
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
		if strings.HasPrefix(request.GlobString, "~") {
			request.GlobString = strings.TrimPrefix(request.GlobString, "~") // TODO: Hack

			parts := strings.Split(request.GlobString, "/")
			if info, err := os.Stat(path.Join(parts[0], "gemini")); err == nil && info.IsDir() {
				request.ServeDirectory("/home/")
				return
			} else {
				request.NotFound("Not found.")
				return
			}
		} else if request.GlobString == "" {
			// Homepage - list all users
			request.ServeDirectory("/home/")
		} else {
			request.NotFound("Not found.")
			return
		}
	})

	context.Start()
}
