package music

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"math"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gammazero/deque"
	sis "gitlab.com/clseibold/smallnetinformationservices"
)

func NopSeekCloser(r io.ReadSeeker) io.ReadSeekCloser {
	/*if _, ok := r.(io.WriterTo); ok {
		return nopCloserWriterTo{r}
	}*/
	return nopCloser{r}
}

type nopCloser struct {
	io.ReadSeeker
}

func (nopCloser) Close() error { return nil }

type RadioBufInterface interface {
	io.WriteCloser
	NewReader() (io.ReadSeekCloser, error)
}

type RadioBuf struct {
	currentMusicFile    MusicFile
	fileChangeIndex     int64
	currentFileLocation int64
	bitrate             int64
	clientCount         int64
	//*os.File
	sync.RWMutex
	readCond     *sync.Cond
	nextSongCond *sync.Cond
}

func (rb *RadioBuf) NewSong(conn *sql.DB, songsPlayed []int64, station *RadioStation, firstTime bool, lastAnnouncerTime *time.Time) (MusicFile, bool, string, error) {
	// Lock to change the main file. Will wait until all RLocks are complete
	rb.Lock()
	if !firstTime {
		// Wait until prompted to get next song
		rb.nextSongCond.Wait()
	}
	//fmt.Printf("%s Station: Getting next song file.\n", station.Name)

	// Determine if should play announcer, and set new announcer time
	t := time.Now()
	var announcer bool = false
	if t.Sub(*lastAnnouncerTime) >= (time.Minute * 30) {
		announcer = true
		min := 0
		if t.Minute() >= 30 {
			min = 30
		}
		*lastAnnouncerTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), min, 0, 0, t.Location())
	}

	// Get New Song from DB based on schedule and current time.
	var file MusicFile
	var currentRadioGenre string = "Announcer"
	var announcer_success bool = false
	if announcer {
		file, announcer_success = GetRandomPublicDomainFileInLibrary_RadioStation_Announcer(conn, station)
	}
	if !announcer_success {
		var success bool
		file, currentRadioGenre, success = GetRandomPublicDomainFileInLibrary_RadioStation(conn, songsPlayed, station)
		if !success {
			// Try getting from Any genre of the station
			fmt.Printf("Couldn't find any file for radio station %s, genre %s. Getting file for radio station from 'Any' genre.\n", station.Name, currentRadioGenre)
			file, success = GetRandomPublicDomainFileInLibrary_RadioStation_Any(conn, songsPlayed, station)
			if !success {
				// Get any file from library instead
				fmt.Printf("Couldn't find any file for radio station %s, genre %s. Getting Any public domain file.\n", station.Name, currentRadioGenre)
				file, success = GetRandomPublicDomainFileInLibrary(conn, songsPlayed)
				if !success {
					fmt.Printf("Couldn't get random public domain file.\n")
					return MusicFile{}, true, "", nil // TODO: This locks forever
				}
			}
		}
	}

	// Open the file.
	/*f, err := os.Open(filepath.Join(musicDirectory, file.Filename))
	if err != nil {
		fmt.Printf("Failed to open.\n")
		return MusicFile{}, true, "", err // TODO: This locks forever
	}*/
	//rb.File = f

	// Skip ID3v2 Tags at start of file
	/*skip_err := tag.SkipID3v2Tags(f)
	if skip_err != nil {
		fmt.Printf("Failed to skip ID3 Headers\n")
	}*/

	// Set starting location to after tags, set the bitrate, update the fileChangeIndex, unlock the lock, and broadcast that the new song was selected
	//currentLocation, _ := f.Seek(0, io.SeekCurrent)
	rb.currentFileLocation = 0

	rb.currentMusicFile = file
	rb.bitrate = file.CbrKbps
	rb.fileChangeIndex += 1
	rb.Unlock()
	//fmt.Printf("%s Station: Unlocked\n", station.Name)
	rb.readCond.Broadcast() // Broadcast that there's been a change in file
	//f.Close()
	return file, false, currentRadioGenre, nil
}

func (rb *RadioBuf) NewReader(old_fileChangeIndex int64, station *RadioStation) (io.ReadSeekCloser, int64, int64, error) {
	rb.RLock()
	//fmt.Printf("%s Station: Checking (%d==%d)\n", station.Name, old_fileChangeIndex, rb.fileChangeIndex)
	for old_fileChangeIndex == rb.fileChangeIndex {
		//fmt.Printf("%s Station: Waiting (%d==%d)\n", station.Name, old_fileChangeIndex, rb.fileChangeIndex)
		rb.readCond.Wait()
		//fmt.Printf("%s Station: Received Broadcast (%d==%d)\n", station.Name, old_fileChangeIndex, rb.fileChangeIndex)
	}

	// Use ffmpeg to get file data without container tags
	path := filepath.Join(musicDirectory, rb.currentMusicFile.Filename)
	noTags := exec.Command("ffmpeg", "-v", "quiet", "-i", path, "-map", "0:a", "-c:a", "copy", "-map_metadata", "-1", "-f", "mp3", "-")
	buffer, err := noTags.Output()
	if err != nil {
		fmt.Printf("FFMPEG Error: %s\n", err.Error())
		// TODO
	}
	f := NopSeekCloser(bytes.NewReader(buffer))
	f.Seek(0, io.SeekStart)
	bitrate := rb.bitrate

	/*f, err := os.Open()
	rb.RUnlock()
	f.Seek(rb.currentFileLocation, io.SeekStart)
	if err != nil {
		return nil, 0, 0, err
	}*/

	//bufferedReader := bufio.NewReader(f)
	return f, rb.fileChangeIndex, bitrate, nil
}

func NewRadioBuffer() (*RadioBuf, error) {
	/*f, err := os.Open(name)
	if err != nil {
		return nil, err
	}*/

	radioBuffer := new(RadioBuf)
	radioBuffer.fileChangeIndex = 0
	radioBuffer.currentFileLocation = 0
	//radioBuffer.File = nil
	radioBuffer.RWMutex = sync.RWMutex{}
	radioBuffer.readCond = sync.NewCond(radioBuffer.RWMutex.RLocker())
	radioBuffer.nextSongCond = sync.NewCond(&radioBuffer.RWMutex) // TODO
	return radioBuffer, nil
}

func handleRadioService(s sis.ServerHandle, conn *sql.DB) {
	//var file_buffer_mutex sync.RWMutex

	/*radioBuffer, _ := NewRadioBuffer()
	go radioService(conn, radioBuffer)
	go fakeClient(radioBuffer)*/

	var totalClientsConnected int64 = 0
	setupStation(s, conn, &RadioStation_Diverse, &totalClientsConnected)
	setupStation(s, conn, &RadioStation_Mainstream, &totalClientsConnected)
	setupStation(s, conn, &RadioStation_Classical, &totalClientsConnected)
	setupStation(s, conn, &RadioStation_Nonmainstream, &totalClientsConnected)
	setupStation(s, conn, &RadioStation_OTR, &totalClientsConnected)
	setupStation(s, conn, &RadioStation_Piano, &totalClientsConnected)
	setupStation(s, conn, &RadioStation_Religious, &totalClientsConnected)

	s.AddRoute("/music/public_radio/schedule", func(request sis.Request) {
		request.Redirect("/music/public_radio/Diverse")
	})
	s.AddRoute("/music/public_radio/", func(request sis.Request) {
		request.Redirect("/music/public_radio")
	})
	s.AddRoute("/music/public_radio", func(request sis.Request) {
		creationDate, _ := time.ParseInLocation(time.RFC3339, "2023-09-19T00:00:00", time.Local)
		updateDate, _ := time.ParseInLocation(time.RFC3339, "2023-11-30T00:00:00", time.Local)
		abstract := `# AuraGem Music: Public Radio

This is AuraGem Music's public radio that plays public domain and royalty free music. All music is collected from sources like the Free Music Archive, archive.org, and Chosic, and stored on my server. This radio does not proxy from the web, unlike other radios over on Gopherspace.
`
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Classification: sis.ScrollResponseUDC_Music, Author: "Christian Lee Seibold", PublishDate: creationDate.UTC(), UpdateDate: updateDate.UTC(), Language: "en", Abstract: abstract})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		var builder strings.Builder
		for _, station := range radioStations {
			fmt.Fprintf(&builder, "=> /music/public_radio/%s/ %s Station\n", url.PathEscape(station.Name), station.Name)
		}
		request.Gemini(fmt.Sprintf(`# AuraGem Music: Public Radio
This is AuraGem Music's public radio that plays public domain and royalty free music. All music is collected from sources like the Free Music Archive, archive.org, and Chosic, and stored on my server. This radio does not proxy from the web, unlike other radios over on Gopherspace.

## Radio Stations
%s
Total Clients Currently Connected: %d

## Other Gemini Radios

=> gemini://gemini.thebackupbox.net/~epoch/blog/radio epoch's Radio (Icecast)
=> gemini://gemini.stillstream.com/ Still Stream Internet Radio
=> gemini://hashnix.club/radio/ Hashnix.Club Radio (with Icecast Support)

## Client Support

Some spec-compliant clients that support playing mp3 files should be able to stream, particularly Lagrange on Desktop. Clients that can pipe into another program will also work. This includes diohsc, which can pipe the data into mpv or vlc.

Currently, Rosy Crow does not seem to support streaming mp3 files. Additionally, Lagrange has several bugs that prevent it from streaming or playing audio on Android and iOS. There is an issue open for this on Github and Bubble, and it seems to apply to all audio playback within the app from all protocols that lagrange supports:
=> https://github.com/skyjake/lagrange/issues/611 Lagrange Github: spartan mp3 links won't play on android
=> gemini://bbs.geminispace.org/s/Lagrange-Issues/14 [#14] Lagrange android doesn't play any sound files

Finally, Lagrange via the AppImage also seems to be bugged for some users and does not stream audio files. You can try to build lagrange yourself instead.

Supported Clients: diohsc (pipe into vlc or mpv), Lagrange on Desktop (Windows, macOS, Flatpak, & self-built w/ mpg123)
Unsuported Clients: GemiNaut, Lagrange on iOS & Android, Kristall, Rosy Crow
Buggy Clients: Lagrange AppImage

## Client Problems

If your client starts playing but then times out, then you can look in settings to disable this timeout, or request that the developer change the implementation so that timeouts only apply when no data has been received for a given period of time.

Some clients, like GemiNaut, will timeout if the connection hasn't been closed by the server within a given time (aka. if the file hasn't downloaded within a given time). This assumes that all files take a specific period of time to download, which is a faulty assumption, especially for users with very slow connections that can't even download a text file within that given time. This is a naive implementation that developers should fix by only timing out based on the period of time that the client has not been given any data. This allows both very slow connections and the downloading of binary and large text files to work.

If your client tries to download infinitely without playing, that means the client is waiting for the connection to close before displaying/playing the file. You can try to look in settings for a streaming option, or request that the developer change this so that data is displayed/played as it comes in. This change is compliant with the spec, as per Section 1.1:

> Note that clients are not obligated to wait until the server closes the connection to begin handling the response. This is shown above only for simplicity/clarity, to emphasise that responsibility for closing the connection under typical conditions lies with the server and that the connection should be closed immediately after the completion of the response body.

=> gemini://geminiprotocol.net/docs/specification.gmi Gemini Specification

In fact, this addition to the spec was made specifically to allow for streaming:

=> gemini://gemini.circumlunar.space/users/solderpunk/gemlog/a-vision-for-gemini-applications.gmi 2020-06-16 A vision for Gemini applications (Solderpunk)

If your client doesn't start playing the music and also times out, then both problems above apply to this client.

## Gemini-Supported Media Player Project

I am also announcing that I will be working on a Media Player that uses VLC (libvlc) as the backend and which will support gemini urls and streams. I have not started the project yet, but I intend to very soon. I hope to make it cross-platform on the desktop, but I plan to support Linux first since it does not seem to have a graphical browser that supports audio streams atm (since Lagrange seems to be buggy with this). It would be nice to also get mobile apps, that that requires much more work and I have to pay to get the app in official appstores, so I won't be able to do any of that for a while, unless Skyjake fixes lagrange's audio streaming on mobile.
`, builder.String(), totalClientsConnected))
	})

	/*
		g.Handle("/music/stream/public_radio", func(c gig.Context) error {
			// Add to client count
			radioBuffer.clientCount += 1

			// Seek to current location in file, then start playing the file
			var old_fileChangeIndex int64 = 0
			for {
				file_reader, fileChangeIndex, bitrate, err := radioBuffer.NewReader(old_fileChangeIndex)
				old_fileChangeIndex = fileChangeIndex

				var rate float64 = float64(bitrate) * 1000 / 8
				rate_reader := RateReader(file_reader, ratelimit.NewBucketWithRate(rate, bitrate*1000/8*2)) // 96 kbps
				err2 := c.Stream("audio/mpeg", rate_reader)
				file_reader.Close()
				if err2 != nil {
					// Remove client from client count
					radioBuffer.clientCount -= 1
					return err
				}
			}
		})
	*/
}

func radioService(conn *sql.DB, radioBuffer *RadioBuf, station *RadioStation) {
	//songsPlayed := make([]int64, 0, 10) // Ids of songs played within the hour, so we don't get repeats for the whole hour
	songsPlayed_genre := "Any" // Used to detect genre changes. If changed, clear songsPlayed

	var songsPlayed *deque.Deque[int64] = new(deque.Deque[int64])
	songsPlayed.SetBaseCap(19)
	songsPlayed.Grow(50)
	//var songsPlayed *deque.Deque[int64] = deque.New[int64](50, 19)

	initialTime := time.Now()
	min := 0
	if initialTime.Minute() >= 30 {
		min = 30
	}

	firstTime := true
	lastAnnouncerTime := time.Date(initialTime.Year(), initialTime.Month(), initialTime.Day(), initialTime.Hour(), min, 0, 0, initialTime.Location())
	for {
		var songsPlayed_slice []int64 = make([]int64, songsPlayed.Len())
		for i := 0; i < songsPlayed.Len(); i++ {
			songsPlayed_slice[i] = songsPlayed.At(i)
		}

		music_file, cont, genre, _ := radioBuffer.NewSong(conn, songsPlayed_slice, station, firstTime, &lastAnnouncerTime)
		if cont {
			continue
		}
		firstTime = false

		// Genre didn't switch, and not in a program (disable exlude_ids stuff for OTR programs)
		if genre == songsPlayed_genre && genre != "OTR-Program" && genre != "OTR-Program-Rerun" {
			max := 1
			switch genre {
			case "Calm Piano":
				max = 40 - 1
			case "Christian Chorale":
				max = 40 - 1
			case "Rock":
				max = 25 - 1
			case "Pop":
				max = 23 - 1
			case "Classical":
				max = 19 - 1
			case "Electronic":
				max = 19 - 1
			case "Blues":
				max = 11 - 1
			case "Jazz":
				max = 10 - 1
			case "Cinematic":
				max = 10 - 1
			case "BeOS":
				max = 9 - 1
			case "Ambient":
				max = 8 - 1
			case "Lofi":
				max = 6 - 1
			case "Acoustic":
				max = 5 - 1
			case "World":
				max = 5 - 1

			case "OTR-Pop":
				max = 37 - 1
			case "OTR-Jazz":
				max = 14 - 1
			case "OTR-Book":
				max = 2
			case "OTR-Acoustic":
				max = 2
			case "OTR-Blues":
				max = 26 - 1
			case "OTR-Country":
				max = 18 - 1

			case "Any":
				max = 50
			default:
				max = 5
			}

			// Pop a quarter of the list when length is greater than or equal to max for the current genre
			// NOTE: We pop off a quarter of the list so that there's more than 1 song to choose from for randomization; We use Ceil so there's always at least 1 song poped off in cases of decimals below 1.
			if songsPlayed.Len() >= max {
				quarter := int(math.Ceil(float64(songsPlayed.Len()) / 4))
				for i := 0; i < quarter; i++ {
					songsPlayed.PopBack()
				}
			}
		} else if genre == songsPlayed_genre && (genre == "OTR-Program" || genre == "OTR-Program-Rerun") {
			// When in OTR-Program, never pop anything off
		} else {
			// Clear all but three songs from the songsPlayed queue, and reset the genre to the new genre
			num := int(math.Max(float64(songsPlayed.Len()-3), 0))
			for i := 0; i < num; i++ {
				songsPlayed.PopBack()
			}
			songsPlayed_genre = genre
		}

		// Push the song id so it is not replayed for the rest of the hour/timeslot
		songsPlayed.PushFront(music_file.Id)
	}
}

func fakeClient(radioBuffer *RadioBuf, station *RadioStation) {
	limiter := time.NewTicker(time.Second * 1)
	radioBuffer.nextSongCond.Broadcast()
	fmt.Printf("Starting Fake Client for %s Station.\n", station.Name)
	//time.NewTicker(time.Millisecond * 125) // TODO: Time for each frame of mp3 file

	// Seek to current location in file, then start playing the file
	var old_fileChangeIndex int64 = 0
	for {
		file_reader, fileChangeIndex, bitrate, _ := radioBuffer.NewReader(old_fileChangeIndex, station)
		old_fileChangeIndex = fileChangeIndex

		tmpBuffer := make([]byte, bitrate*(1000/8))
		for {
			n, r_err := file_reader.Read(tmpBuffer)
			radioBuffer.currentFileLocation += int64(n)
			if r_err == io.EOF {
				// End of file
				//fmt.Printf("End of file.\n")
				//time.Sleep(5 * time.Second) // NOTE: Hacky delay so that clients don't get too far behind
				break
			}
			<-limiter.C // Wait 1 second after every read
		}
		file_reader.Close()
		radioBuffer.nextSongCond.Broadcast()
	}
	//limiter.Stop()
}
