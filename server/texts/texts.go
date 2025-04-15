package texts

import (
	// "context"
	// "database/sql"
	// "time"
	// "fmt"
	// "strings"
	// "strconv"

	// "github.com/krixano/ponixserver/src/db"

	"gitlab.com/clseibold/auragem_sis/server/texts/christianity"
	"gitlab.com/clseibold/auragem_sis/server/texts/islam"
	"gitlab.com/clseibold/auragem_sis/server/texts/judaism"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

func HandleTexts(g sis.VirtualServerHandle) {
	g.AddRoute("/scriptures/", func(request *sis.Request) {
		request.SetClassification(sis.ScrollResponseUDC_Scripture)
		request.Gemini(`# Religious Texts

=> /scriptures/jewish/ ✡ Jewish Texts
=> /scriptures/christian/ ✝ Christian Texts
=> /scriptures/islam/ ☪ Islamic Texts
`)
	})

	judaism.HandleJewishTexts(g)
	christianity.HandleChristianTexts(g)
	islam.HandleIslamicTexts(g)
}
