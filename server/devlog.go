package main

import (
	"time"

	"gitlab.com/clseibold/auragem_sis/server/utils"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

func handleDevlog(s sis.VirtualServerHandle) {
	publishDate, _ := time.ParseInLocation(time.RFC3339, "2021-04-24T00:00:00", time.Local)
	s.AddRoute("/devlog/atom.xml", func(request *sis.Request) {
		atom, feedTitle, lastUpdate := utils.GenerateAtomFrom("SIS/auragem_gemini/devlog/index.gmi", "gemini://auragem.ddns.net", "gemini://auragem.ddns.net/devlog", "Christian Lee Seibold", "christian.seibold32@outlook.com")
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: publishDate, UpdateDate: lastUpdate, Language: "en", Abstract: "# " + feedTitle + "\n"})
		if request.ScrollMetadataRequested() {
			_ = request.SendAbstract("text/xml")
			return
		}
		_ = request.TextWithMimetype("text/xml", atom)
	})
}
