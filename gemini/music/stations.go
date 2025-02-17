package music

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/juju/ratelimit"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

func GetRadioGenre(currentTime time.Time, station *RadioStation) string {
	var currentHour int64 = int64(currentTime.Hour())
	weekend := currentTime.Weekday() == time.Saturday || currentTime.Weekday() == time.Sunday

	if !weekend {
		return station.Weekday_Schedule[currentHour]
	} else {
		return station.Weekend_Schedule[currentHour]
	}
}

func setupStation(s sis.ServerHandle, conn *sql.DB, station *RadioStation, totalClientsConnected *int64) {
	radioBuffer, _ := NewRadioBuffer()

	go radioService(conn, radioBuffer, station)
	go fakeClient(radioBuffer, station)

	s.AddRoute("/music/public_radio/"+url.PathEscape(station.Name), func(request *sis.Request) {
		creationDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-14T18:07:00", time.Local)
		creationDate = creationDate.UTC()
		abstract := fmt.Sprintf("# AuraGem Public Radio - %s Station\n\n%s\nClients Connected: %d\n", station.Name, station.Description, radioBuffer.clientCount)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Music, Author: "Christian Lee Seibold", PublishDate: creationDate, UpdateDate: creationDate, Language: "en", Abstract: abstract})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		currentTime := time.Now()
		radioGenre := GetRadioGenre(currentTime, station)

		attribution := ""
		if radioBuffer.currentMusicFile.Attribution != "" {
			attribution = "\n" + radioBuffer.currentMusicFile.Attribution
		}

		var scheduleBuilder strings.Builder
		fmt.Fprintf(&scheduleBuilder, "```\n")
		fmt.Fprintf(&scheduleBuilder, "%-10s | %-10s | %-20s | %-20s\n", "PT", "CT", "Weekdays", "Weekends")
		fmt.Fprintf(&scheduleBuilder, "-----------|------------|----------------------|---------------------\n")

		prev_ampm_pst := ""
		prev_ampm_cst := ""
		for i := int64(0); i < 24; i++ {
			weekday_genre := station.Weekday_Schedule[i]
			weekend_genre := station.Weekend_Schedule[i]

			hour_start_cst := i
			hour_start_pm_cst := false
			hour_end_cst := i + 1

			hour_start_pst := hour_start_cst - 2
			hour_start_pm_pst := false
			hour_end_pst := hour_end_cst - 2

			if hour_start_pst < 0 {
				hour_start_pst += 24
				hour_start_pm_pst = true
			}
			if hour_end_pst < 0 {
				hour_end_pst += 24
			}

			if hour_start_cst > 12 {
				hour_start_cst -= 12
				hour_start_pm_cst = true
			} else if hour_start_cst == 12 {
				hour_start_pm_cst = true
			}
			if hour_end_cst > 12 {
				hour_end_cst -= 12
			}
			if hour_start_pst > 12 {
				hour_start_pst -= 12
				hour_start_pm_pst = true
			} else if hour_start_pst == 12 {
				hour_start_pm_pst = true
			}
			if hour_end_pst > 12 {
				hour_end_pst -= 12
			}

			if hour_start_cst == 0 {
				hour_start_cst = 12
			}
			if hour_end_cst == 0 {
				hour_end_cst = 12
			}
			if hour_start_pst == 0 {
				hour_start_pst = 12
			}
			if hour_end_pst == 0 {
				hour_end_pst = 12
			}

			ampm_cst := "AM"
			if hour_start_pm_cst {
				ampm_cst = "PM"
			}
			ampm_pst := "AM"
			if hour_start_pm_pst {
				ampm_pst = "PM"
			}

			if prev_ampm_cst == ampm_cst {
				ampm_cst = "  "
			} else {
				prev_ampm_cst = ampm_cst
			}
			if prev_ampm_pst == ampm_pst {
				ampm_pst = "  "
			} else {
				prev_ampm_pst = ampm_pst
			}

			fmt.Fprintf(&scheduleBuilder, "%2d - %2d %s | %2d - %2d %s | %-20s | %-20s\n", hour_start_pst, hour_end_pst, ampm_pst, hour_start_cst, hour_end_cst, ampm_cst, weekday_genre, weekend_genre)
		}
		fmt.Fprintf(&scheduleBuilder, "```\n")

		if station.Name == "Old-Time-Radio" {
			fmt.Fprintf(&scheduleBuilder, "\n## Program Schedule\n```\n")
			for wd := time.Sunday; wd <= time.Saturday; wd++ {
				program := station.ProgramInfo[wd]
				fmt.Fprintf(&scheduleBuilder, "%-20s %s\n", wd.String()+":", program)
			}
			fmt.Fprintf(&scheduleBuilder, "```\n")

			fmt.Fprintf(&scheduleBuilder, "\nAdditional Notes: Episode 1 of The Adventures of Philip Marlowe will begin at 12:00 PM CST on Friday, December 1 2023, and Episode 1 of Yours Truly, Johnny Dollar will begin at 12:00 PM CST on Tuesday, November 28, 2023. Those listed as 'Program-Rerun' on the schedule always play the previous episode for those who might have missed it.\n")

			// Add schedules of each program here:
			//programs := []string { "The Adventures of Philip Marlowe",  }
		}

		// Station homepage here
		template := `# AuraGem Music Public Radio - %s Station

%s

=> /music/public_radio/ Public Radio Home
=> /music/public_radio/%s/schedule_feed/ Schedule Gemsub Feed
=> /music/stream/public_radio/%s.mp3 Stream Station

Clients Currently Connected to Station: %d
Current Time and Genre: %s CST (%s)
Current song playing: %s by %s
%s

## Schedule
%s
`
		request.Gemini(fmt.Sprintf(template, station.Name, station.Description, url.PathEscape(station.Name), url.PathEscape(station.Name), radioBuffer.clientCount, currentTime.Format("03:04 PM"), radioGenre, radioBuffer.currentMusicFile.Title, radioBuffer.currentMusicFile.Artist, attribution, scheduleBuilder.String()))
	})

	s.AddRoute("/music/public_radio/"+url.PathEscape(station.Name)+"/schedule_feed", func(request *sis.Request) {
		creationDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-14T18:07:00", time.Local)
		creationDate = creationDate.UTC()
		abstract := fmt.Sprintf("# AuraGem Public Radio - %s Station Schedule\n", station.Name)
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Music, Author: "Christian Lee Seibold", PublishDate: creationDate, UpdateDate: time.Now(), Language: "en", Abstract: abstract})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		currentTime := time.Now()
		current_wd := currentTime.Weekday()
		program := station.ProgramInfo[current_wd]
		episode := station.CurrentEpisode[program]
		var hour int64 = 0

		if current_wd != time.Saturday && current_wd != time.Sunday {
			for key, val := range station.Weekday_Schedule {
				if val == "OTR-Program" || val == "Program" {
					hour = key
					break
				}
			}
		} else {
			for key, val := range station.Weekend_Schedule {
				if val == "OTR-Program" || val == "Program" {
					hour = key
					break
				}
			}
		}

		year, month, day := currentTime.Date()
		timeOfProgram := time.Date(year, month, day, int(hour), 0, 0, 0, currentTime.Location())

		template := `# AuraGem Public Radio - %s Station Schedule

%s

=> /music/public_radio/ Public Radio Home
=> /music/public_radio/%s Station Homepage

=> /music/stream/public_radio/%s.mp3 %s %s UTC %s: Ep. %d

`
		request.Gemini(fmt.Sprintf(template, station.Name, station.Description, url.PathEscape(station.Name), url.PathEscape(station.Name), timeOfProgram.UTC().Format("2006-01-02"), timeOfProgram.UTC().Format("03:04 PM"), program, episode))
	})

	s.AddRoute("/music/stream/public_radio", func(request *sis.Request) {
		request.Redirect("/music/stream/public_radio/" + url.PathEscape(RadioStation_Diverse.Name) + ".mp3")
	})
	s.AddRoute("/music/stream/public_radio/"+url.PathEscape(station.Name), func(request *sis.Request) {
		request.Redirect("/music/stream/public_radio/" + url.PathEscape(station.Name) + ".mp3")
	})
	s.AddRoute("/music/stream/public_radio/"+url.PathEscape(station.Name)+".mp3", func(request *sis.Request) {
		creationDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-14T18:07:00", time.Local)
		creationDate = creationDate.UTC()
		abstract := ""
		if request.ScrollMetadataRequested {
			currentTime := time.Now()
			radioGenre := GetRadioGenre(currentTime, station)
			attribution := ""
			if radioBuffer.currentMusicFile.Attribution != "" {
				attribution = "\n" + radioBuffer.currentMusicFile.Attribution
			}

			abstract = fmt.Sprintf("# AuraGem Public Radio - %s Station\n\n%s\nClients Currently Connected to Station: %d\nCurrent Time and Genre: %s CST (%s)\nCurrent song playing: %s by %s\n%s", station.Name, station.Description, radioBuffer.clientCount, currentTime.Format("03:04 PM"), radioGenre, radioBuffer.currentMusicFile.Title, radioBuffer.currentMusicFile.Artist, attribution)
		}

		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Music, Author: "Christian Lee Seibold", PublishDate: creationDate, UpdateDate: time.Now(), Language: "en", Abstract: abstract})
		if request.ScrollMetadataRequested {
			request.SendAbstract("audio/mpeg")
			return
		}

		// Station streaming here
		// Add to client count
		radioBuffer.clientCount += 1
		(*totalClientsConnected) += 1

		// Seek to current location in file, then start playing the file
		var old_fileChangeIndex int64 = 0
		for {
			file_reader, fileChangeIndex, bitrate, _ := radioBuffer.NewReader(old_fileChangeIndex, station)
			old_fileChangeIndex = fileChangeIndex

			var rate float64 = float64(bitrate) * 1000 / 8 // Turn kbps to KB/s
			rate_reader := RateReader(file_reader, ratelimit.NewBucketWithRate(rate, bitrate*1000/8*2))
			err2 := request.StreamBuffer("audio/mpeg", rate_reader, make([]byte, bitrate*1000/8)) // Match buffer size
			file_reader.Close()
			if err2 != nil {
				// Remove client from client count
				radioBuffer.clientCount -= 1
				(*totalClientsConnected) -= 1
				return //err
			}
			//time.Sleep(1 * time.Second)
		}
	})
}

type RadioSchedule map[int64]string
type RadioStation struct {
	Name             string
	Description      string
	Weekday_Schedule RadioSchedule
	Weekend_Schedule RadioSchedule
	AnyCategory      []string                // What Genres are allowed in "Any"
	ProgramInfo      map[time.Weekday]string // What programs to play on a specific weekday
	CurrentEpisode   map[string]int          // Key = album (as program name), Value = track # (as episode number)
}

var radioStations = []*RadioStation{
	&RadioStation_Diverse,
	&RadioStation_Mainstream,
	&RadioStation_Classical,
	&RadioStation_Nonmainstream,
	&RadioStation_OTR,
	&RadioStation_Piano,
	&RadioStation_Religious,
}

var RadioStation_Diverse = RadioStation{
	"Diverse",
	"Plays a diverse set of genres at specific time slots.",
	map[int64]string{
		0:  "Ambient",
		1:  "Lofi",
		2:  "Cinematic",
		3:  "Acoustic",
		4:  "BeOS",
		5:  "World",
		6:  "Classical",
		7:  "Classical",
		8:  "Jazz",
		9:  "Pop",
		10: "Rock",
		11: "Pop",
		12: "Rock",
		13: "Any",
		14: "Any",
		15: "Any",
		16: "Any",
		17: "Any",
		18: "Blues",
		19: "Pop",
		20: "Rock",
		21: "Electronic",
		22: "Calm Piano", // Normal Sleep
		23: "Classical",
	},
	map[int64]string{
		0:  "Electronic",
		1:  "Electronic",
		2:  "Calm Piano",
		3:  "Ambient",
		4:  "Lofi",
		5:  "World",
		6:  "Classical",
		7:  "Classical",
		8:  "Jazz",
		9:  "Cinematic",
		10: "Acoustic",
		11: "Blues",
		12: "Pop",
		13: "Rock",
		14: "Pop",
		15: "Rock",
		16: "Any",
		17: "Any", // Work ends
		18: "Blues",
		19: "Pop",
		20: "Any",
		21: "Classical",
		22: "Rock", // Normal Sleep
		23: "Electronic",
	},
	[]string{"Ambient", "Lofi", "Cinematic", "Acoustic", "World", "Jazz", "Pop", "Rock", "Blues", "Electronic", "Calm Piano"},
	map[time.Weekday]string{},
	map[string]int{},
}

var RadioStation_Mainstream = RadioStation{
	"Mainstream",
	"Plays mainstream genres, including Pop, Rock, Acoustic, Cinematic, Calm Piano, and World",
	map[int64]string{
		0:  "Calm Piano",
		1:  "Cinematic",
		2:  "Acoustic",
		3:  "World",
		4:  "World",
		5:  "Calm Piano", // Early Wake up
		6:  "Calm Piano",
		7:  "Acoustic",
		8:  "Pop",  // Normal Wake Up
		9:  "Rock", // Work starts
		10: "Pop",
		11: "Rock",
		12: "Pop", // Late Wake Up / Lunch time
		13: "Rock",
		14: "Any",
		15: "Any",
		16: "Any",
		17: "Any", // Work ends
		18: "Pop",
		19: "Rock",
		20: "Cinematic",
		21: "Cinematic",
		22: "Calm Piano", // Normal Sleep
		23: "Calm Piano",
	},
	map[int64]string{
		0:  "Calm Piano", // Late Sleep (Weekends)
		1:  "Calm Piano",
		2:  "Calm Piano",
		3:  "Calm Piano",
		4:  "World",
		5:  "Calm Piano", // Early Wake up
		6:  "Calm Piano",
		7:  "Cinematic",
		8:  "Acoustic", // Normal Wake up
		9:  "Any",
		10: "Any",
		11: "Any",
		12: "Pop", // Late wake up / Lunch
		13: "Rock",
		14: "Pop",
		15: "Rock",
		16: "Acoustic",
		17: "Any",
		18: "Pop",
		19: "Rock",
		20: "Any",
		21: "Any",
		22: "Cinematic", // Normal Sleep
		23: "Cinematic",
	},
	[]string{"Cinematic", "Acoustic", "World", "Pop", "Rock", "Calm Piano"},
	map[time.Weekday]string{},
	map[string]int{},
}

var RadioStation_Classical = RadioStation{
	"Classical",
	"Plays Classical and Classical-adjacent genres, including Classical, Jazz, and Blues.",
	map[int64]string{
		0:  "Blues",
		1:  "Blues",
		2:  "Jazz",
		3:  "Blues",
		4:  "Blues",
		5:  "Classical", // Early Wake up
		6:  "Classical",
		7:  "Jazz",
		8:  "Classical", // Normal Wake Up
		9:  "Jazz",      // Work starts
		10: "Jazz",
		11: "Calm Piano",
		12: "Jazz", // Late wake up / Lunch
		13: "Jazz",
		14: "Classical",
		15: "Any",
		16: "Any",
		17: "Jazz", // Work ends
		18: "Jazz",
		19: "Blues",
		20: "Blues", // TODO
		21: "Classical",
		22: "Classical", // Normal Sleep
		23: "Classical",
	},
	map[int64]string{
		0:  "Classical", // Late Sleep (Weekends)
		1:  "Classical",
		2:  "Classical",
		3:  "Blues",
		4:  "Blues",
		5:  "Classical", // Early Wake up
		6:  "Classical",
		7:  "Jazz",
		8:  "Classical", // Normal Wake Up
		9:  "Jazz",
		10: "Jazz",
		11: "Any",
		12: "Classical", // Late wake up / Lunch
		13: "Classical",
		14: "Jazz",
		15: "Calm Piano",
		16: "Any",
		17: "Any",
		18: "Jazz",
		19: "Jazz",
		20: "Blues",
		21: "Blues",
		22: "Blues", // Normal Sleep
		23: "Blues",
	},
	[]string{"Classical", "Jazz", "Blues", "Calm Piano"},
	map[time.Weekday]string{},
	map[string]int{},
}

var RadioStation_Nonmainstream = RadioStation{
	"Non-mainstream",
	"Plays non-mainstream non-classical music, including Electronic, Lofi, Ambient, and Rock.",
	map[int64]string{
		0:  "Ambient", // Late Sleep (Weekends)
		1:  "Lofi",
		2:  "Lofi",
		3:  "Any",
		4:  "BeOS",
		5:  "Ambient", // Early Wake up
		6:  "Lofi",
		7:  "Electronic",
		8:  "Ambient", // Normal Wake Up
		9:  "Lofi",    // Work starts
		10: "Electronic",
		11: "Any",
		12: "Rock", // Late wake up / Lunch
		13: "Electronic",
		14: "Any",
		15: "Any",
		16: "Any",
		17: "Any", // Work ends
		18: "Rock",
		19: "Rock",
		20: "Electronic",
		21: "Electronic",
		22: "Ambient", // Normal Sleep
		23: "Ambient",
	},
	map[int64]string{
		0:  "Ambient", // Late Sleep (Weekends)
		1:  "Ambient",
		2:  "Ambient",
		3:  "Lofi",
		4:  "Lofi",
		5:  "Ambient", // Early Wake up
		6:  "Lofi",
		7:  "Electronic",
		8:  "Ambient", // Normal Wake Up
		9:  "Lofi",
		10: "Electronic",
		11: "Any",
		12: "Ambient", // Late wake up
		13: "Electronic",
		14: "Any",
		15: "Any",
		16: "Any",
		17: "Any",
		18: "Rock",
		19: "Electronic",
		20: "Rock",
		21: "Rock",
		22: "Electronic", // Normal Sleep
		23: "Electronic",
	},
	[]string{"Electronic", "Lofi", "Ambient", "BeOS", "Rock"},
	map[time.Weekday]string{},
	map[string]int{},
}

var RadioStation_OTR = RadioStation{
	"Old-Time-Radio",
	"Plays old time radio shows, radio dramatizations, and old public-domain music gathered from the Internet Archive. Note that Pop in older music included Rock and Rock n' Roll, and so this station does the same. Rock became a distinct genre/style in the late 1960s (source: Oxford Companion to Music).",
	map[int64]string{
		0:  "OTR-Jazz", // Late Sleep (Weekends)
		1:  "OTR-Program-Rerun",
		2:  "Any",
		3:  "OTR-Blues",
		4:  "OTR-Acoustic",
		5:  "OTR-Jazz", // Early Wake up
		6:  "OTR-Book",
		7:  "OTR-Country",
		8:  "OTR-Pop", // Normal Wake Up
		9:  "Any",     // Work starts
		10: "Any",
		11: "OTR-Book",
		12: "Any", // Late wake up / Lunch
		13: "OTR-Program",
		14: "OTR-Pop",
		15: "OTR-Jazz",
		16: "OTR-Acoustic",
		17: "Any", // Work ends
		18: "Any",
		19: "Any",
		20: "OTR-Book",
		21: "OTR-Blues",
		22: "OTR-Pop", // Normal Sleep
		23: "OTR-Pop",
	},
	map[int64]string{
		0:  "OTR-Jazz", // Late Sleep (Weekends)
		1:  "OTR-Program-Rerun",
		2:  "Any",
		3:  "OTR-Blues",
		4:  "OTR-Acoustic",
		5:  "OTR-Jazz", // Early Wake up
		6:  "OTR-Book",
		7:  "OTR-Country",
		8:  "OTR-Pop", // Normal Wake Up
		9:  "Any",
		10: "OTR-Book",
		11: "Any",
		12: "OTR-Program", // Late wake up
		13: "OTR-Pop",
		14: "OTR-Jazz",
		15: "OTR-Acoustic",
		16: "Any",
		17: "Any",
		18: "Any",
		19: "Any",
		20: "OTR-Blues",
		21: "OTR-Book",
		22: "OTR-Pop", // Normal Sleep
		23: "OTR-Pop",
	},
	[]string{"OTR-Jazz", "OTR-Pop", "OTR-Acoustic", "OTR-Blues", "OTR-Country"},
	map[time.Weekday]string{
		time.Sunday:    "Suspense: The Radio Show",
		time.Monday:    "The Adventures of Philip Marlowe",
		time.Tuesday:   "Yours Truly, Johnny Dollar",
		time.Wednesday: "McLevy Series",
		time.Thursday:  "Suspense: The Radio Show",
		time.Friday:    "The Adventures of Philip Marlowe",
		time.Saturday:  "Yours Truly, Johnny Dollar",
	},
	map[string]int{
		"The Adventures of Philip Marlowe": 1,
		"Yours Truly, Johnny Dollar":       2,
		"Suspense: The Radio Show":         2,
	},
}

var RadioStation_Piano = RadioStation{
	"Piano",
	"Plays Piano music.",
	map[int64]string{
		0:  "Calm Piano",
		1:  "Calm Piano",
		2:  "Calm Piano",
		3:  "Calm Piano",
		4:  "Calm Piano",
		5:  "Calm Piano", // Early Wake up
		6:  "Calm Piano",
		7:  "Calm Piano",
		8:  "Calm Piano", // Normal Wake Up
		9:  "Calm Piano", // Work starts
		10: "Calm Piano",
		11: "Calm Piano",
		12: "Calm Piano", // Late Wake Up / Lunch time
		13: "Calm Piano",
		14: "Calm Piano",
		15: "Calm Piano",
		16: "Calm Piano",
		17: "Calm Piano", // Work ends
		18: "Calm Piano",
		19: "Calm Piano",
		20: "Calm Piano",
		21: "Calm Piano",
		22: "Calm Piano", // Normal Sleep
		23: "Calm Piano",
	},
	map[int64]string{
		0:  "Calm Piano", // Late Sleep (Weekends)
		1:  "Calm Piano",
		2:  "Calm Piano",
		3:  "Calm Piano",
		4:  "Calm Piano",
		5:  "Calm Piano", // Early Wake up
		6:  "Calm Piano",
		7:  "Calm Piano",
		8:  "Calm Piano", // Normal Wake up
		9:  "Calm Piano",
		10: "Calm Piano",
		11: "Calm Piano",
		12: "Calm Piano", // Late wake up / Lunch
		13: "Calm Piano",
		14: "Calm Piano",
		15: "Calm Piano",
		16: "Calm Piano",
		17: "Calm Piano",
		18: "Calm Piano",
		19: "Calm Piano",
		20: "Calm Piano",
		21: "Calm Piano",
		22: "Calm Piano", // Normal Sleep
		23: "Calm Piano",
	},
	[]string{"Cinematic", "Calm Piano", "Classical"},
	map[time.Weekday]string{},
	map[string]int{},
}

var RadioStation_Religious = RadioStation{
	"Religious",
	"Plays a diverse set of religious music. Currently Christian-only until I can find music of other religions.",
	map[int64]string{
		0:  "Christian Chorale",
		1:  "Christian Chorale",
		2:  "Christian Chorale",
		3:  "Christian Chorale",
		4:  "Christian Chorale",
		5:  "Christian Chorale",
		6:  "Christian Chorale",
		7:  "Christian Chorale",
		8:  "Christian Chorale",
		9:  "Christian Chorale",
		10: "Christian Chorale",
		11: "Christian Chorale",
		12: "Christian Chorale",
		13: "Christian Chorale",
		14: "Christian Chorale",
		15: "Christian Chorale",
		16: "Christian Chorale",
		17: "Christian Chorale",
		18: "Christian Chorale",
		19: "Christian Chorale",
		20: "Christian Chorale",
		21: "Christian Chorale",
		22: "Christian Chorale", // Normal Sleep
		23: "Christian Chorale",
	},
	map[int64]string{
		0:  "Christian Chorale",
		1:  "Christian Chorale",
		2:  "Christian Chorale",
		3:  "Christian Chorale",
		4:  "Christian Chorale",
		5:  "Christian Chorale",
		6:  "Christian Chorale",
		7:  "Christian Chorale",
		8:  "Christian Chorale",
		9:  "Christian Chorale",
		10: "Christian Chorale",
		11: "Christian Chorale",
		12: "Christian Chorale",
		13: "Christian Chorale",
		14: "Christian Chorale",
		15: "Christian Chorale",
		16: "Christian Chorale",
		17: "Christian Chorale", // Work ends
		18: "Christian Chorale",
		19: "Christian Chorale",
		20: "Christian Chorale",
		21: "Christian Chorale",
		22: "Christian Chorale", // Normal Sleep
		23: "Christian Chorale",
	},
	[]string{"Classical", "Christian Chorale"},
	map[time.Weekday]string{},
	map[string]int{},
}
