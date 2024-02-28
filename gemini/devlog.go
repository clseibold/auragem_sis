package gemini

import (
	utils "gitlab.com/clseibold/auragem_sis/gemini/utils"
	sis "gitlab.com/clseibold/smallnetinformationservices"
)

func handleDevlog(g sis.ServerHandle) {
	/*g.AddRoute("/~krixano/gemlog/atom.xml", func(request sis.Request) {
		// c.NoContent(gig.StatusRedirectTemporary, "/devlog/atom.xml")
		request.TextWithMimetype("text/xml", generateAtomFrom("SIS/gemini/devlog/index.gmi", "gemini://auragem.letz.dev", "gemini://auragem.letz.dev/devlog", "Christian \"Krixano\" Seibold", "christian.seibold32@outlook.com"))
	})*/

	g.AddRoute("/devlog/atom.xml", func(request sis.Request) {
		request.TextWithMimetype("text/xml", utils.GenerateAtomFrom("SIS/gemini/devlog/index.gmi", "gemini://auragem.letz.dev", "gemini://auragem.letz.dev/devlog", "Christian \"Krixano\" Seibold", "krixano@protonmail.com"))
	})
}
