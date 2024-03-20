package main

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"

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

const cpuprofile = "./cpu.prof"
const memprofile = "./mem.prof"

func main() {
	/*conn, _ := sql.Open("firebirdsql", firebirdConnectionString)
	defer conn.Close()*/

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

	gemini.GeminiCommand.Execute()

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
