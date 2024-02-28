package main

import (
	"gitlab.com/clseibold/auragem_sis/gemini"
	_ "gitlab.com/clseibold/auragem_sis/migration"
	//"io"
	//"io/ioutil"
	//"os"
	//"database/sql"
	//"github.com/pitr/gig"
	//"github.com/nakagami/firebirdsql"
	//_ "gitlab.com/clseibold/auragem_sis/migration"
	//"github.com/spf13/cobra"
)

func main() {
	/*conn, _ := sql.Open("firebirdsql", firebirdConnectionString)
	defer conn.Close()*/

	gemini.GeminiCommand.Execute()
}
