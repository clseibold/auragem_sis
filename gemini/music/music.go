package music

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/dhowden/tag"
	"gitlab.com/clseibold/auragem_sis/config"
	"gitlab.com/clseibold/auragem_sis/db"
	sis "gitlab.com/clseibold/smallnetinformationservices"
	// "github.com/giorgisio/goav/avformat"
	// "github.com/giorgisio/goav/avcodec"
	// "github.com/giorgisio/goav/avutil"
)

// TODO: Jukebox - infinite stream of music picked randomely from user's library
// TODO: Store filesize and length in Db

// fpcalc 0084b5fd2f6f7c47065257203a34af4908f5b6e5.mp3 -plain

// var musicDirectory_SMB = "music/" // Outdated
var musicDirectory = config.MusicDirectory
var userSongQuota = config.MusicSongQuota
var globalSongQuota float64 = config.GlobalSongQuota // 38666.67

var registerNotification = `# AuraGem Music

You have selected a certificate that has not been registered yet. Please register here:

=> /music/register Register Page
=> /music/quota How the Quota System Works
`

//go:embed music_index.gmi
var index_gmi string

//go:embed music_index_scroll.scroll
var index_scroll string

func HandleMusic(s sis.ServerHandle) {
	publishDate, _ := time.ParseInLocation(time.RFC3339, "2022-07-15T00:00:00", time.Local)
	updateDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-19T13:51:00", time.Local)
	// ffmpeg (goav) Stuff: Register all formats and codecs
	// avformat.AvRegisterAll()
	// avcodec.AvcodecRegisterAll()

	// Database Connection
	conn := db.NewConn(db.MusicDB)
	conn.SetMaxOpenConns(500)
	conn.SetMaxIdleConns(3)
	conn.SetConnMaxLifetime(time.Hour * 4)
	//defer conn.Close() // TODO

	// Music File Streaming Throttling
	//throttlePool := iothrottler.NewIOThrottlerPool(iothrottler.BytesPerSecond * (1024.0 * 40.0 * 1.15))        // (320 Kbps = 40 KB/s) * 1.15 seconds = 368 Kbps
	//throttlePool_no_buffer := iothrottler.NewIOThrottlerPool(iothrottler.BytesPerSecond * (1024.0 * 40.0 * 1)) // (320 Kbps = 40 KB/s) * 1.15 seconds = 368 Kbps
	//defer throttlePool.ReleasePool()

	handleRadioService(s, conn)

	s.AddRoute("/music/", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: publishDate, UpdateDate: updateDate, Language: "en", Abstract: "# AuraGem Music\nA music service where you can upload a limited number of mp3s over Titan and listen to your private music library over Scroll/Gemini/Spartan. Stream individual songs or full albums, or use the \"Shuffled Stream\" feature that acts like a private radio of random songs from your library.\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		cert := request.UserCert
		if cert == nil {
			if request.Type == sis.ServerType_Gemini {
				request.Gemini(index_gmi)
			} else if request.Type == sis.ServerType_Scroll {
				request.Scroll(index_scroll)
			} else if request.Type == sis.ServerType_Spartan {
				request.TemporaryFailure("Service not available over Spartan. Please visit over Gemini or Scroll.")
			} else if request.Type == sis.ServerType_Nex {
				request.TemporaryFailure("Service not available over Nex. Please visit over Gemini or Scroll.")
			} else if request.Type == sis.ServerType_Gopher {
				request.TemporaryFailure("Service not available over Gopher. Please visit over Gemini or Scroll.")
			}
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: publishDate, UpdateDate: updateDate, Language: "en", Abstract: "# AuraGem Music - " + user.Username + "\n"})
				if request.ScrollMetadataRequested {
					request.SendAbstract("")
					return
				}
				getUserDashboard(request, conn, user)
				return
			}
		}
	})

	s.AddRoute("/music/quota", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: publishDate, UpdateDate: updateDate, Language: "en", Abstract: "# AuraGem Music - How the Quota System Works\nDescribes the quota system.\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}
		template := `# AuraGem Music - How the Quota System Works

Each song adds to your quota 1 divided by the number of people who have uploaded that same song. If 3 people have uploaded the 3 same songs, only 1 song gets added to each person's quota (3 songs / 3 uploaders). However, if you are the only person who has uploaded a song, then 1 will be added to your quota (1 song / 1 uploader). The maximum quota that each user has is currently set to %d.

Note that the below calculations assume an average mp3 file size of 7.78 MB per song:
If 38 people uploaded 1000 *unique* songs, that would fill up ~288.8 GB. But if 38 people uploaded the same 1000 songs, that would only be as if one person uploaded those 1000 songs, and that'd only be ~7.6 GB. And the quota added to each person is 1000 songs / 38 uploaders = 26.3
Which means that leaves each user with 973.7 more songs. If each user uploaded 973.7 *unique* songs, that would take up ~281 GB. 281 GB + the 7.6 GB for the 1000 non-unique songs would equal ~288.6 GB

The idea behind this system is to take into account duplicate uploads as being considered only 1 upload, because I use deduplication of song files to save space. By dividing an upload across the number of users that have uploaded that file, each user gets the same slice of quota, and the sum adds up to 1.

=> /music/ Back
`
		request.Gemini(fmt.Sprintf(template, userSongQuota))
	})

	s.AddRoute("/music/about", func(request sis.Request) {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: publishDate, UpdateDate: updateDate, Language: "en", Abstract: "# About AuraGem Music\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}
		template := `# About AuraGem Music

This is a gemini capsule that allows users to upload their own mp3s (or oggs) to thier own private library (via Titan) and stream/download them via Gemini. A user's library is completely private. Nobody else can see the library, and songs are only streamable by the user that uploaded that song.

In order to save space, AuraGem Music deduplicates songs by taking the hash of the audio contents. This is only done when the songs of multiple users are the *exact* same, and is done on upload of a song. Deduplication also has the benefit of lowering a user's quota. If the exact same song is in multiple users' libraries, the sum of the quotas for that song for each user adds up to 1. This is because the song is only stored once on the server. The quota is spread evenly between each of the users that have uploaded the song. The more users, the less quota each user has for that one song. You can find out more about how the quota system works below:

=> /music/quota How the Quota System Works
`
		request.Gemini(template)
	})

	s.AddRoute("/music/random", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate.")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				file, exists := GetRandomFileInUserLibray(conn, user.Id)
				if !exists {
					request.NotFound("File not found. Your library may be empty.")
					return
				}
				openFile, err := os.Open(filepath.Join(musicDirectory, file.Filename))
				if err != nil {
					panic(err)
				}
				request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# Random Music File\n"})
				request.SetNoLanguage()
				if request.ScrollMetadataRequested {
					request.SendAbstract("audio/mpeg")
					return
				}
				request.Stream("audio/mpeg", openFile) // TODO: Use mimetype from db
				openFile.Close()
				return
			}
		}
	})

	s.AddRoute("/music/upload", func(request sis.Request) {
		if request.UserCert == nil {
			request.RequestClientCert("Please enable a certificate.")
			return
		}
		uploadLink := ""
		uploadMethod := ""
		if request.Type == sis.ServerType_Gemini || request.Type == sis.ServerType_Scroll {
			titanHost := "titan://auragem.letz.dev/"
			if request.Hostname() == "192.168.0.60" {
				titanHost = "titan://192.168.0.60/"
			} else if request.Hostname() == "auragem.ddns.net" {
				titanHost = "titan://auragem.ddns.net/"
			}
			uploadLink = "=> " + titanHost + "/music/upload"
			uploadMethod = "Titan"
		}

		request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: publishDate, UpdateDate: updateDate, Abstract: "# Upload File with " + uploadMethod + "\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		request.Gemini(fmt.Sprintf(`# Upload File with %s

Upload an mp3 music file to this page with %s. It will then be automatically added to your library. Please make sure that the metadata tags on the mp3 are correct and filled in before uploading, especially the Title, AlbumArtist, and Album tags.

%s Upload

=> gemini://transjovian.org/titan About Titan
`, uploadMethod, uploadMethod, uploadLink))
	})

	s.AddUploadRoute("/music/upload", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate.")
			return
		} else if request.Upload {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				//return c.Gemini(registerNotification)
				request.TemporaryFailure("You must be registered first before you can upload.")
				return
			} else {
				// First, check mimetype if using Titan
				mimetype := request.DataMime
				if (request.Type == sis.ServerType_Gemini || request.Type == sis.ServerType_Scroll) && !strings.HasPrefix(mimetype, "audio/mpeg") && !strings.HasPrefix(mimetype, "audio/mp3") {
					request.TemporaryFailure("Only mp3 audio files are allowed.")
					return
				} else if !(request.Type == sis.ServerType_Gemini || request.Type == sis.ServerType_Scroll) {
					request.TemporaryFailure("Upload only supported via Titan.")
					return
				}

				// Check the size
				if request.DataSize > 15*1024*1024 { // Max of 15 MB
					request.TemporaryFailure("File too large. Max size is 15 MiB.")
					return
				}

				file, read_err := request.GetUploadData()
				if read_err != nil {
					return //read_err
				}

				// TODO: Check if data folder is mounted properly before doing anything?

				// Then, get hash of file
				hash, _ := tag.Sum(bytes.NewReader(file))
				fmt.Printf("Hash: %s\n", hash)

				// Check if file hash is already in user's library. If so, return error.
				_, existsInUserLibrary := GetFileInUserLibrary_hash(conn, hash, user.Id)
				if existsInUserLibrary {
					request.TemporaryFailure("File already in library.")
					return
				}

				// Otherwise, check if file hash is already in (general) library.
				musicFile, existsInLibrary := GetFileInLibrary_hash(conn, hash)
				if existsInLibrary {
					// If so, it doesn't exist in user's library, so add it to the user's library.
					fmt.Printf("Adding '%s' to User's library\n", hash)

					// TODO: Check new user quota (by adding to the user's current quota)
					// TODO: New songQuotaPerUser() function??

					AddFileToUserLibrary(conn, musicFile.Id, user.Id, false)
				} else {
					// If not, add file to library and to the user's library.
					fmt.Printf("Adding '%s' to General Library and User's library\n", hash)

					// Check GLOBAL Quota Count to make sure harddrive doesn't become too full
					globalQuotaCount := Admin_GetGlobalQuota(conn)
					if globalQuotaCount+1 >= globalSongQuota {
						request.TemporaryFailure("Global quota reached. Server is full.")
						return
					}

					// Check new user quota (by adding ONE to the user's current quota)
					if user.QuotaCount+1 >= float64(userSongQuota) {
						request.TemporaryFailure("User quota reached. You cannot upload any more files.")
						return
					}

					// Get tag info
					m, tag_err := tag.ReadFrom(bytes.NewReader(file))
					if tag_err != nil {
						fmt.Printf("Error getting tags from '%s': %s\n", hash, tag_err)
						request.TemporaryFailure("Failed to get tags from file.")
						return
					}

					// Make sure file is mp3
					if m.FileType() != tag.MP3 {
						request.TemporaryFailure("Only mp3 audio files are allowed.")
						return
					}

					// Check that mp3 is CBR and not VBR
					bitrate_str := m.Raw()["stream_bitrate"].(string)
					if strings.HasSuffix(bitrate_str, "VBR") {
						request.TemporaryFailure("Variable Bitrate mp3 files are not supported. Please upload only CBR mp3 files.")
						return
					}

					// Add to library, and make sure it succeeds
					musicFile, success := AddFileToLibrary(conn, hash, m, false)
					if !success {
						request.TemporaryFailure("Failed to add to library.")
						return
					}

					// Write out the file
					write_err := os.WriteFile(musicDirectory+hash+".mp3", file, 0600)
					if write_err != nil {
						//return c.NoContent(gig.StatusPermanentFailure, "Failed to store file.")
						panic(write_err)
					}
					//SMB_WriteFile(hash+".mp3", file)

					// Add to user library
					AddFileToUserLibrary(conn, musicFile.Id, user.Id, false)
				}

				request.Redirect("%s%s/music/", request.Server.Scheme(), request.Hostname())
				return
				//return c.NoContent(gig.StatusRedirectTemporary, "gemini://%s/music/", c.URL().Host)
			}
		}
	})

	s.AddRoute("/music/*", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				unescape, _ := url.PathUnescape(request.GlobString)
				/*unescape, _ := url.PathUnescape(c.URL().EscapedPath())*/
				hash := strings.TrimSuffix(strings.Replace(unescape, "/music/", "", 1), filepath.Ext(unescape))

				file, exists := GetFileInUserLibrary_hash(conn, hash, user.Id)
				if !exists {
					request.NotFound("File not found.")
					return
				}

				abstract := "# " + file.Title + "\n"
				if file.Album != "" {
					abstract += "Album: " + file.Album + "\n"
				}
				if file.Tracknumber != 0 {
					abstract += "Track: " + strconv.Itoa(file.Tracknumber) + "\n"
				}
				if file.Discnumber != 0 && file.Discnumber != 1 {
					abstract += "Disk: " + strconv.Itoa(file.Discnumber) + "\n"
				}
				if file.Albumartist != "" {
					abstract += "Album Artist: " + file.Albumartist + "\n"
				}
				if file.Composer != "" {
					abstract += "Composer: " + file.Composer + "\n"
				}
				if file.Genre != "" {
					abstract += "Genre: " + file.Genre + "\n"
				}
				if file.Releaseyear != 0 {
					abstract += "Release Year: " + strconv.Itoa(file.Releaseyear) + "\n"
				}
				abstract += "Kbps: " + strconv.Itoa(int(file.CbrKbps)) + "\n"
				if file.Attribution != "" {
					abstract += "\nAttribution:\n" + file.Attribution + "\n"
				}
				request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: file.Artist, Abstract: abstract})
				request.SetNoLanguage()
				if request.ScrollMetadataRequested {
					request.SendAbstract("audio/mpeg")
					return
				}

				StreamFile(request, file)
				return
				//q := `SELECT COUNT(*) FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? AND library.filename=?`
			}
		}
	})

	s.AddRoute("/music/albums", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				albums := GetAlbumsInUserLibrary(conn, user.Id)

				request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: user.Date_joined, Abstract: "# AuraGem Music - " + user.Username + "\n## Albums\n"})
				if request.ScrollMetadataRequested {
					request.SendAbstract("")
					return
				}

				var builder strings.Builder
				for _, album := range albums {
					fmt.Fprintf(&builder, "=> /music/artist/%s/%s %s - %s\n", url.PathEscape(album.Albumartist), url.PathEscape(album.Album), album.Album, album.Albumartist)
				}

				request.Gemini(fmt.Sprintf(`# AuraGem Music - %s
## Albums

=> /music/ Dashboard

%s
`, user.Username, builder.String()))
				return
			}
		}
	})

	s.AddRoute("/music/artists", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				artists := GetArtistsInUserLibrary(conn, user.Id)

				request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: user.Date_joined, Abstract: "# AuraGem Music - " + user.Username + "\n## Artists\n"})
				if request.ScrollMetadataRequested {
					request.SendAbstract("")
					return
				}

				var builder strings.Builder
				for _, artist := range artists {
					fmt.Fprintf(&builder, "=> /music/artist/%s %s\n", url.PathEscape(artist), artist)
				}

				request.Gemini(fmt.Sprintf(`# AuraGem Music - %s
## Artists

=> /music/ Dashboard

%s
`, user.Username, builder.String()))
				return
			}
		}
	})

	s.AddRoute("/music/artist/*", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				unescape, _ := url.PathUnescape(request.GlobString) // url.PathUnescape(c.URL().EscapedPath())
				p := strings.Split(strings.Replace(unescape, "/music/artist/", "", 1), "/")
				artist := p[0]
				album := ""
				if len(p) > 1 {
					album = p[1]
					albumSongs(request, conn, user, artist, album)
					return
				} else {
					artistAlbums(request, conn, user, artist)
					return
				}
			}
		}
	})

	s.AddRoute("/music/stream/artist/*", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				unescape, _ := url.PathUnescape(request.GlobString) //url.PathUnescape(c.URL().EscapedPath())
				p := strings.Split(strings.Replace(unescape, "/music/stream/artist/", "", 1), "/")
				artist := p[0]
				album := ""
				if len(p) > 1 {
					album = p[1]
					// Stream all songs from Album
					streamAlbumSongs(request, conn, user, artist, strings.Replace(album, ".mp3", "", 1))
					return
					//return albumSongs(c, conn, user, artist, album)
				} else {
					// Stream all songs from Artist
					streamArtistSongs(request, conn, user, strings.Replace(artist, ".mp3", "", 1))
					return
					//return artistAlbums(c, conn, user, artist)
				}
			}
		}
	})

	s.AddRoute("/music/stream/random", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# AuraGem Music Shuffled Stream - " + user.Username + "\n"})
				if request.ScrollMetadataRequested {
					request.SendAbstract("audio/mpeg")
					return
				}

				StreamRandomFiles(request, conn, user)
				return
			}
		}
	})

	s.AddRoute("/music/register", func(request sis.Request) {
		cert := request.UserCert

		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			query, err := request.Query()
			if err != nil {
				request.TemporaryFailure(err.Error())
				return
			} else if query == "" {
				request.RequestInput("Enter a username:")
				return
			} else {
				// Do registration
				registerUser(request, conn, query, request.UserCertHash_Gemini())
				return
			}
		}
	})

	// Admin pages
	s.AddRoute("/music/admin", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered || !user.Is_staff {
				request.ClientCertNotAuthorized("Not authorized for this page")
				return
			} else {
				adminPage(request, conn, user)
				return
			}
		}
	})

	s.AddRoute("/music/admin/genre", func(request sis.Request) {
		genre_string, err := request.Query()
		if err != nil {
			request.TemporaryFailure(err.Error())
			return
		}

		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered || !user.Is_staff {
				request.ClientCertNotAuthorized("Not authorized for this page")
				return
			} else {
				adminGenrePage(request, conn, user, genre_string)
				return
			}
		}
	})

	// Library Management Handles
	s.AddRoute("/music/manage", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				manageLibrary(request, user)
				return
			}
		}
	})

	s.AddRoute("/music/manage/delete", func(request sis.Request) {
		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				manageLibrary_deleteSelection(request, conn, user)
				return
			}
		}
	})

	s.AddRoute("/music/manage/delete/:hash", func(request sis.Request) {
		hash := request.GetParam("hash")
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure(err.Error())
			return
		}

		cert := request.UserCert
		if cert == nil {
			request.RequestClientCert("Please enable a certificate")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash_Gemini())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				file, exists := GetFileInUserLibrary_hash(conn, hash, user.Id)
				if !exists {
					request.TemporaryFailure("File not in user library.")
					return
				}
				artist := file.Artist
				if artist == "" {
					artist = file.Albumartist
				}

				if query == "yes" || query == "'yes'" {
					manageLibrary_deleteFile(request, conn, user, hash)
					return
				} else {
					request.RequestInput("Type 'yes' to delete %s by %s from your library.", file.Title, artist)
					return
				}
			}
		}
	})
}

// ---------------------

func registerUser(request sis.Request, conn *sql.DB, username string, certHash string) {
	// Ensure user doesn't already exist
	row := conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM members WHERE certificate=?", certHash)

	var numRows int
	err := row.Scan(&numRows)
	if err != nil {
		panic(err)
	}
	if numRows < 1 {
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# AuraGem Music - Register User " + username + "\n"})
		if request.ScrollMetadataRequested {
			request.SendAbstract("")
			return
		}

		// Certificate doesn't already exist - Register User
		zone, _ := time.Now().Zone()
		conn.ExecContext(context.Background(), "INSERT INTO members (certificate, username, language, timezone, is_staff, is_active, date_joined) VALUES (?, ?, ?, ?, ?, ?, ?)", certHash, username, "en-US", zone, false, true, time.Now())

		user, _ := GetUser(conn, certHash)

		// Add default uploads
		defaultUploads := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9}

		for _, fileid := range defaultUploads {
			//conn.ExecContext(context.Background(), "INSERT INTO uploads (memberid, fileid, date_added) values (?, ?, ?)", user.id, fileid, time.Now())
			//conn.ExecContext(context.Background(), "UPDATE library SET uploadcount = uploadcount + 1 WHERE id=?", fileid)

			AddFileToUserLibrary(conn, fileid, user.Id, false)
		}
	}

	request.Redirect("/music/")
}

func getUserDashboard(request sis.Request, conn *sql.DB, user MusicUser) {
	template := `# AuraGem Music - %s

Quota: %.2f / %d songs (%.1f%%)

=> /music/public_radio Public Radio
=> /music/quota How the Quota System Works

=> /music/manage Manage Library
=> /music/upload Upload MP3

=> /music/albums Albums
=> /music/artists Artists
=> /music/stream/random Shuffled Stream
%s
`

	musicFiles := GetFilesInUserLibrary(conn, user.Id)

	var builder strings.Builder

	if user.Is_staff {
		fmt.Fprintf(&builder, "=> /music/admin Admin Dashboard\n")
	}
	fmt.Fprintf(&builder, "\n")

	for _, file := range musicFiles {
		artist := file.Composer
		if file.Composer == "" || (file.Genre != "Classical" && file.Genre != "String Quartets") {
			artist = file.Albumartist
		}
		fmt.Fprintf(&builder, "=> %s %s (%s)\n", url.PathEscape(file.Filename), file.Title, artist)
	}

	if len(musicFiles) == 0 {
		fmt.Fprintf(&builder, "Your music library is empty.")
	}

	request.Gemini(fmt.Sprintf(template, user.Username, user.QuotaCount, userSongQuota, user.QuotaCount/float64(userSongQuota)*100, builder.String()))
}

func artistAlbums(request sis.Request, conn *sql.DB, user MusicUser, artist string) {
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# AuraGem Music - " + user.Username + "\n## Artist Albums: " + artist + "\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	albums := GetAlbumsFromArtistInUserLibrary(conn, user.Id, artist)

	var builder strings.Builder
	for _, album := range albums {
		fmt.Fprintf(&builder, "=> /music/artist/%s/%s %s\n", url.PathEscape(album.Albumartist), url.PathEscape(album.Album), album.Album)
	}

	request.Gemini(fmt.Sprintf(`# AuraGem Music - %s
## Artist Albums: %s

=> /music/ Dashboard
=> /music/artists Artists
=> /music/stream/artist/%s Stream All Artist's Songs

%s
`, user.Username, artist, url.PathEscape(artist), builder.String()))
}

func albumSongs(request sis.Request, conn *sql.DB, user MusicUser, artist string, album string) {
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# AuraGem Music - " + user.Username + "\n## Album: " + album + " by " + artist + "\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	musicFiles := GetFilesFromAlbumInUserLibrary(conn, user.Id, artist, album)

	var builder strings.Builder
	var albumartist string
	for i, file := range musicFiles {
		/*artist := file.Composer
		if file.Composer == "" {
			artist = file.Albumartist
		}*/
		if i == 0 {
			albumartist = file.Albumartist
		}
		fmt.Fprintf(&builder, "=> /music/%s %d. %s\n", url.PathEscape(file.Filename), file.Tracknumber, file.Title)
	}

	request.Gemini(fmt.Sprintf(`# AuraGem Music - %s
## Album: %s by %s

=> /music/ Dashboard
=> /music/artist/%s %s
=> /music/stream/artist/%s/%s Stream Full Album

%s
`, user.Username, album, artist, url.PathEscape(albumartist), albumartist, url.PathEscape(albumartist), url.PathEscape(album), builder.String()))
}

func adminPage(request sis.Request, conn *sql.DB, user MusicUser) {
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# AuraGem Music - Admin\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	var builder strings.Builder
	radioGenres := Admin_RadioGenreCounts(conn)
	for _, genre := range radioGenres {
		fmt.Fprintf(&builder, "=> /music/admin/genre?%s %s (%d)\n", url.QueryEscape(genre.Name), genre.Name, genre.Count)
	}
	radioGenres = nil // NOTE: Manually garbage collect the radioGenres slice
	template := `# AuraGem Music - Admin

Global Quota: %.2f / %.2f (%.1f%%)
User Quota Average: %.2f / %d (%.1f%%)
User Count: %d
Artist Count: %d
Album Count: %d

## Radio Genres

%s

=> /music/radiotest Public Radio Test
`

	globalQuotaCount := Admin_GetGlobalQuota(conn)
	userCount := Admin_UserCount(conn)
	avgUserQuotaCount := globalQuotaCount / float64(userCount)
	artistCount := Admin_ArtistCount(conn)
	albumCount := Admin_AlbumCount(conn)

	request.Gemini(fmt.Sprintf(template, globalQuotaCount, globalSongQuota, globalQuotaCount/globalSongQuota*100, avgUserQuotaCount, userSongQuota, avgUserQuotaCount/float64(userSongQuota)*100, userCount, artistCount, albumCount, builder.String()))
}

func adminGenrePage(request sis.Request, conn *sql.DB, user MusicUser, genre_string string) {
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# AuraGem Music - Admin: Genre" + genre_string + "\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	songsInGenre := GetFilesInGenre(conn, genre_string)
	var builder strings.Builder
	fmt.Fprintf(&builder, "```\n")
	for _, song := range songsInGenre {
		fmt.Fprintf(&builder, "%-25s %-25s\n", song.Title, song.Artist)
	}
	fmt.Fprintf(&builder, "```\n")
	songsInGenre = nil // Manually garbage collect
	request.Gemini(fmt.Sprintf(`# AuraGem Music - Admin: Genre %s

%s
`, genre_string, builder.String()))
}

// Streams all songs in album in one streams
func streamAlbumSongs(request sis.Request, conn *sql.DB, user MusicUser, artist string, album string) {
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# Stream Album " + album + " by " + artist + "\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	musicFiles := GetFilesFromAlbumInUserLibrary(conn, user.Id, artist, album)
	fmt.Printf("Music Files: %v\n", musicFiles)

	/*filenames := make([]string, 0, len(musicFiles))
	for _, file := range musicFiles {
		filenames = append(filenames, file.Filename)
	}*/

	StreamMultipleFiles(request, musicFiles)
}

func streamArtistSongs(request sis.Request, conn *sql.DB, user MusicUser, artist string) {
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# Stream Songs by " + artist + "\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	musicFiles := GetFilesFromArtistInUserLibrary(conn, user.Id, artist)

	/*filenames := make([]string, 0, len(musicFiles))
	for _, file := range musicFiles {
		filenames = append(filenames, file.Filename)
	}*/

	StreamMultipleFiles(request, musicFiles)
}

// ----- Manage Library Functions -----

func manageLibrary(request sis.Request, user MusicUser) {
	request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: user.Date_joined, Abstract: "# Manage Library - " + user.Username + "\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	request.Gemini(fmt.Sprintf(`# Manage Library - %s

Choose what you want to do. These links will direct you to pages that will allow you to select songs out of your library for the action you selected.

=> /music/ Dashboard
=> /music/manage/delete Delete Songs
=> /music/manage/edit Edit Song Metadata (Coming Soon)
=> /music/upload Upload MP3

`, user.Username))
}

func manageLibrary_deleteSelection(request sis.Request, conn *sql.DB, user MusicUser) {
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# Manage Library: Delete Selection - " + user.Username + "\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	// TODO: Add Pagination
	musicFiles := GetFilesInUserLibrary(conn, user.Id)

	var builder strings.Builder
	for _, file := range musicFiles {
		artist := file.Composer
		if file.Composer == "" || (file.Genre != "Classical" && file.Genre != "String Quartets") {
			artist = file.Albumartist
		}
		fmt.Fprintf(&builder, "=> /music/manage/delete/%s %s (%s)\n", url.PathEscape(file.Filehash), file.Title, artist)
	}

	if len(musicFiles) == 0 {
		fmt.Fprintf(&builder, "Your music library is empty.")
	}

	request.Gemini(fmt.Sprintf(`# Manage Library: Delete Selection - %s

=> /music/ Dashboard
=> /music/manage Manage Library

Click a song you want to delete. Once prompted, type 'yes' to delete. Then, you'll be redirected back here.

%s

`, user.Username, builder.String()))
}

func manageLibrary_deleteFile(request sis.Request, conn *sql.DB, user MusicUser, hash string) {
	file, exists := GetFileInUserLibrary_hash(conn, hash, user.Id)
	if !exists {
		request.TemporaryFailure("File not in user library.")
		return
	}

	request.SetScrollMetadataResponse(sis.ScrollMetadata{Abstract: "# Manage Library: Delete File - " + user.Username + "\nDelete " + file.Title + " by " + file.Artist + " (" + file.Album + ")\n"})
	if request.ScrollMetadataRequested {
		request.SendAbstract("")
		return
	}

	RemoveFileFromUserLibrary(conn, file.Id, user.Id)
	request.Redirect("/music/manage/delete")
	//return c.NoContent(gig.StatusRedirectTemporary, "/music/manage/delete")
}
