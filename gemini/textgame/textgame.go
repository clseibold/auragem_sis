package textgame

import (
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

func HandleTextGame(g sis.VirtualServerHandle) {
	g.AddRoute("/textgame/", Homepage)
}

func Homepage(request *sis.Request) {
	request.Gemini("# Text Game\n")
}
