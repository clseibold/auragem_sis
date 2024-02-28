package texts

import (
	// "context"
	// "database/sql"
	// "time"
	// "fmt"
	// "strings"
	// "strconv"

	// "github.com/krixano/ponixserver/src/db"

	"gitlab.com/clseibold/auragem_sis/gemini/texts/christianity"
	"gitlab.com/clseibold/auragem_sis/gemini/texts/islam"
	"gitlab.com/clseibold/auragem_sis/gemini/texts/judaism"
	sis "gitlab.com/clseibold/smallnetinformationservices"
)

func HandleTexts(g sis.ServerHandle) {
	g.AddRoute("/texts/", func(request sis.Request) {
		request.Gemini(`# Religious Texts

=> /texts/jewish/ ✡ Jewish Texts
=> /texts/christian/ ✝ Christian Texts
=> /texts/islam/ ☪ Islamic Texts
`)
	})

	judaism.HandleJewishTexts(g)
	christianity.HandleChristianTexts(g)
	islam.HandleIslamicTexts(g)
}
