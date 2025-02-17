package textola

import (
	context_pkg "context"
	_ "embed"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sis "gitlab.com/sis-suite/smallnetinformationservices"
	"golang.org/x/time/rate"
)

type TextolaContentType int

var TextolaContentType_Fiction TextolaContentType = 0
var TextolaContentType_Song TextolaContentType = 1
var TextolaContentType_Nonfiction TextolaContentType = 2
var TextolaContentType_Poetry TextolaContentType = 3
var TextolaContentType_OldClassicsFiction TextolaContentType = 4
var TextolaContentType_LearningNonfiction TextolaContentType = 5
var TextolaContentType_Technical TextolaContentType = 6
var TextolaContentType_Legal TextolaContentType = 7

func getWPMForContent(contentType TextolaContentType) int {
	switch contentType {
	case TextolaContentType_Fiction:
		return 260
	case TextolaContentType_Song:
		return 270
	case TextolaContentType_Nonfiction:
		return 238
	case TextolaContentType_Poetry:
		return 225
	case TextolaContentType_OldClassicsFiction:
		return 213 - 13
	case TextolaContentType_LearningNonfiction:
		return 200
	case TextolaContentType_Technical:
		return 178
	case TextolaContentType_Legal:
		return 162
	}

	return 260
}

type TextolaContext struct {
	currentText    TextolaText
	currentLine    int64
	currentSeconds int64
	readCond       *sync.Cond
	changeInt      int64
	mutex          sync.RWMutex
	newText        bool
	startTime      time.Time
}

type TextolaText struct {
	contentType  TextolaContentType
	lines        []TextolaLine
	totalSeconds int64
}

type TextolaLine struct {
	text    string
	seconds int64 // How many milliseconds to show the line
}

func makeTextFromString(announcer string, text string, contentType TextolaContentType) TextolaText {
	lines := strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n'
	})

	textolaText := TextolaText{
		contentType:  contentType,
		lines:        make([]TextolaLine, len(lines)+1),
		totalSeconds: 0,
	}

	// Handle announcer line
	announcer_words := strings.Fields(announcer)
	announcer_wpmForContent := getWPMForContent(contentType)
	announcer_seconds := int64(math.Ceil(float64(len(announcer_words)) / float64(announcer_wpmForContent) * 60))
	textolaText.lines[0] = TextolaLine{text: announcer, seconds: announcer_seconds}
	textolaText.totalSeconds += announcer_seconds

	// Handle rest of lines
	for index := 0; index < len(lines); index++ {
		line := strings.TrimSpace(lines[index])
		if strings.HasPrefix(line, "####") {
			line = "\n" + strings.TrimPrefix(line, "####")
		} else if strings.HasPrefix(line, "###") {
			line = "\n" + strings.TrimPrefix(line, "###")
		} else if strings.HasPrefix(line, "##") {
			line = "\n" + strings.TrimPrefix(line, "##")
		} else if strings.HasPrefix(line, "#") {
			line = "\n" + strings.TrimPrefix(line, "#")
		}
		words := strings.Fields(line)
		wpmForContent := getWPMForContent(contentType)
		seconds := int64(math.Ceil(float64(len(words)) / float64(wpmForContent) * 60))
		textolaText.lines[index+1] = TextolaLine{text: line, seconds: seconds}
		textolaText.totalSeconds += seconds
	}

	return textolaText
}

/*
func makeTextFromFile(announcer string, filepath string, contentType TextolaContentType) TextolaText {
	text, _ := os.ReadFile(filepath)
	return makeTextFromString(announcer, string(text), contentType)
}
*/

var hourAnnouncer TextolaText = makeTextFromString("Welcome.", "This is AuraGem Textola, the text equivalent of radio!\n$time\n", TextolaContentType_Fiction)

var schedule = map[int]string{
	0:  "",
	1:  "",
	2:  "",
	3:  "",
	4:  "",
	5:  "JewishPrayerMorning",
	6:  "",
	7:  "",
	8:  "",
	9:  "",
	10: "",
	11: "",
	12: "JewishPrayerAfternoon",
	13: "",
	14: "",
	15: "",
	16: "JewishPrayerAfternoon",
	17: "",
	18: "",
	19: "JewishPrayerEvening",
	20: "Fiction",
	21: "",
	22: "",
	23: "",
}

func getTextFromSchedule(s sis.ServerHandle) TextolaText {
	currentTime := time.Now()
	scheduleString := schedule[currentTime.Hour()]
	switch scheduleString {
	case "JewishPrayerMorning":
		filedata, _ := s.GetServer().FS.ReadFile("textolatexts/prayers/jewishprayermorning.txt")
		return makeTextFromString("Presenting Jewish Morning Prayers, Weekdays", string(filedata), TextolaContentType_Poetry)
	case "JewishPrayerAfternoon":
		filedata, _ := s.GetServer().FS.ReadFile("textolatexts/prayers/jewishprayerafternoon.txt")
		return makeTextFromString("Presenting Jewish Afternoon Prayers, Weekdays", string(filedata), TextolaContentType_Poetry)
	case "JewishPrayerEvening":
		filedata, _ := s.GetServer().FS.ReadFile("textolatexts/prayers/jewishprayerevening.txt")
		return makeTextFromString("Presenting Jewish Evening Prayers, Weekdays", string(filedata), TextolaContentType_Poetry)
	case "Fiction":
		filedata, _ := s.GetServer().FS.ReadFile("textolatexts/fiction/the_cask_of_amontillado.txt")
		return makeTextFromString("Presenting Edgar Allan Poe's The Cask of Amontillado", string(filedata), TextolaContentType_OldClassicsFiction)
	case "": // Anything (random)
		filedata, _ := s.GetServer().FS.ReadFile("textolatexts/fiction/the_cask_of_amontillado.txt")
		return makeTextFromString("Presenting Edgar Allan Poe's The Cask of Amontillado", string(filedata), TextolaContentType_OldClassicsFiction)
	}

	filedata, _ := s.GetServer().FS.ReadFile("the_cask_of_amontillado.txt")
	return makeTextFromString("Presenting Edgar Allan Poe's The Cask of Amontillado", string(filedata), TextolaContentType_OldClassicsFiction)
}

func HandleTextola(s sis.ServerHandle) {
	publishDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-19T13:51:00", time.Local)
	updateDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-19T13:51:00", time.Local)
	var context *TextolaContext = &TextolaContext{
		currentText: getTextFromSchedule(s),
		mutex:       sync.RWMutex{},
	}
	context.readCond = sync.NewCond(context.mutex.RLocker())

	//fmt.Printf("Textola Text: (Lines %d) %#v\n", len(context.currentText.lines), context.currentText.lines)

	var connectedClients atomic.Int64
	s.AddRoute("/textola/", func(request *sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: publishDate, UpdateDate: updateDate, Language: "en", Abstract: "# Textola\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}
		request.Gemini("# Textola\n\n")
		limiter := rate.NewLimiter(rate.Every(time.Second), 1)

		connectedClients.Add(1)

		currentChangeInt := context.changeInt
		currentLine := context.currentText.lines[context.currentLine]
		currentCountDown := currentLine.seconds - context.currentSeconds

		// Print out the current Line
		if context.newText {
			err := request.Gemini("\n\n------- " + currentLine.text + fmt.Sprintf(" [%d WPM] -------\n", getWPMForContent(context.currentText.contentType)))
			if err != nil {
				connectedClients.Add(-1)
				return
				//return err
			}
		} else {
			err := request.Gemini(currentLine.text + "\n")
			if err != nil {
				connectedClients.Add(-1)
				return
				//return err
			}
		}

		for {
			//<-ticker.C
			limiter.Wait(context_pkg.Background())

			currentCountDown -= 1
			if currentCountDown <= 0 {
				context.readCond.L.Lock() // RLock
				for currentChangeInt == context.changeInt {
					//fmt.Printf("Waiting.\n")
					context.readCond.Wait()
				}
				//fmt.Printf("Done Waiting.\n")

				// Switch to the next line/text and print it
				currentChangeInt = context.changeInt
				currentLine = context.currentText.lines[context.currentLine]
				currentCountDown = currentLine.seconds - context.currentSeconds
				context.readCond.L.Unlock()

				text := currentLine.text
				var err error
				if context.newText {
					err = request.Gemini("\n\n------- " + text + fmt.Sprintf(" [%d WPM] -------\n", getWPMForContent(context.currentText.contentType)))
				} else if strings.HasPrefix(text, "$time") {
					err = request.Gemini(fmt.Sprintf("The current time in UTC is %s.\n", time.Now().UTC().Format("03:04 PM")))
				} else {
					err = request.Gemini(text + "\n")
				}
				if err != nil {
					break
				}
			}
		}

		connectedClients.Add(-1)
	})
	go fakeClient(s, context)
}

func fakeClient(s sis.ServerHandle, context *TextolaContext) {
	// Seconds Ticker
	//ticker := time.NewTicker(time.Millisecond / 100)
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	currentLine := context.currentLine
	context.newText = true
	context.startTime = time.Now()
	previousAnnouncerTime := time.Now()
	for {
		//<-ticker.C
		limiter.Wait(context_pkg.Background())
		context.currentSeconds += 1

		if context.currentSeconds >= context.currentText.lines[currentLine].seconds {
			context.mutex.Lock()
			if currentLine+1 >= int64(len(context.currentText.lines)) {
				// Switch to next text and reset currentMS
				context.currentSeconds = 0

				currentLine = 0
				context.currentLine = 0
				//prevStartTime := context.startTime
				context.startTime = time.Now()

				if context.startTime.Sub(previousAnnouncerTime) >= (time.Minute * 15) {
					// If Hour changed or 30 minutes after, switch to hourAnnouncer text
					context.currentText = hourAnnouncer
					previousAnnouncerTime = context.startTime
				} else {
					// Else, switch to the next text on the schedule
					context.currentText = getTextFromSchedule(s)
				}

				context.newText = true
				context.changeInt += 1
				context.readCond.Broadcast()
			} else {
				// Switch to the next line if there is one next
				context.currentSeconds = 0

				currentLine += 1
				context.currentLine += 1

				context.newText = false
				context.changeInt += 1
				context.readCond.Broadcast()
			}
			context.mutex.Unlock()
		}
	}
}
