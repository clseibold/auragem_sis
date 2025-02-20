package youtube

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	//"log"

	ytd "github.com/kkdai/youtube/v2"
	"gitlab.com/clseibold/auragem_sis/config"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	youtubeAPIKey = config.YoutubeApiKey
	maxResults    = int64(25) /*flag.Int64("max-results", 25, "Max YouTube results")*/
)

//go:embed index.gmi
var content embed.FS

func HandleYoutube(s sis.VirtualServerHandle) {
	// Create Youtube Service
	service, err1 := youtube.NewService(context.Background(), option.WithAPIKey(youtubeAPIKey))
	if err1 != nil {
		//log.Fatalf("Error creating new Youtube client: %v", err1)
		panic(err1)
	}
	searchRoute := getSearchRouteFunc(service)
	videoPageRoute := getVideoPageRouteFunc(service)
	videoDownloadRoute := getVideoDownloadRouteFunc()

	s.AddRoute("/cgi-bin/youtube.cgi", func(request *sis.Request) {
		request.Redirect("/youtube/") // TODO: Temporary Redirect
	})
	s.AddRoute("/youtube", indexRoute)
	s.AddRoute("/youtube/search", searchRoute)
	s.AddRoute("/youtube/search/:page", searchRoute)
	s.AddRoute("/youtube/video/:id/", videoPageRoute)
	s.AddRoute("/youtube/downloadVideo/:quality/:id", videoDownloadRoute)
	handleCaptionDownload(s)

	handleChannelPage(s, service)
	handlePlaylistPage(s, service)
}

func indexRoute(request *sis.Request) {
	creationDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-17T11:57:00", time.Local)
	abstract := "#AuraGem YouTube Proxy\n\nProxies YouTube to Scroll/Gemini. Lets you search and download videos and playlists.\n"
	request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: "Christian Lee Seibold", PublishDate: creationDate.UTC(), UpdateDate: creationDate.UTC(), Language: "en", Abstract: abstract})
	if request.ScrollMetadataRequested() {
		request.Scroll(abstract)
		return
	}

	request.Gemini("# AuraGem YouTube Proxy\n\nWelcome to the AuraGem YouTube Proxy!\n\n")
	request.PromptLine("/youtube/search/", "Search")
	request.Gemini("=> / AuraGem Home\n")
	request.Gemini("=> gemini://kwiecien.us/gemcast/20210425.gmi See This Proxy Featured on Gemini Radio\n")
}
func getSearchRouteFunc(service *youtube.Service) sis.RequestHandler {
	return func(request *sis.Request) {
		request.SetNoLanguage()
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure(err.Error())
			return
		}

		if query == "" {
			request.RequestInput("Search Query:")
		} else {
			rawQuery, err := request.RawQuery()
			if err != nil {
				request.TemporaryFailure(err.Error())
				return
			}

			abstract := fmt.Sprintf("# AuraGem YouTube Proxy Search - Query %s\n", query)
			request.SetScrollMetadataResponse(sis.ScrollMetadata{Language: "en", Abstract: abstract})
			if request.ScrollMetadataRequested() {
				request.Scroll(abstract)
				return
			}

			page := request.GetParam("page")
			if page == "" {
				searchYoutube(request, service, query, rawQuery, "")
			} else {
				searchYoutube(request, service, query, rawQuery, page)
			}
		}
	}
}

func handleVideoClassification(video *youtube.Video, request *sis.Request) {
	handleTopicId := false
	switch video.Snippet.CategoryId {
	case "1": // Film & Animation
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "2": // Autos & Vehicles
		request.SetClassification(sis.ScrollResponseUDC_Engineering)
	case "10": // Music
		request.SetClassification(sis.ScrollResponseUDC_Music)
	case "15": // Pets and Animals
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "17": // Sports
		request.SetClassification(sis.ScrollResponseUDC_Sport)
	case "18": // Short Movies
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "19": // Travel & Events
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "20": // Gaming
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "21": // Videoblogging
		request.SetClassification(sis.ScrollResponseUDC_PersonalLog)
	case "22": // People & Blogs
		request.SetClassification(sis.ScrollResponseUDC_PersonalLog)
	case "23": // Comedy
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "24": // Entertainment
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "25": // News and Politics
		request.SetClassification(sis.ScrollResponseUDC_SocialScience)
	case "26": // Howto and Style
		request.SetClassification(sis.ScrollResponseUDC_Reference)
	case "27": // Education
		request.SetClassification(sis.ScrollResponseUDC_SocialScience)
	case "28": // Science and Technology
		handleTopicId = true
		//request.SetClassification(sis.ScrollResponseUDC_Technology) // TODO
	case "29": // Nonprofits & Activism
		request.SetClassification(sis.ScrollResponseUDC_SocialScience)
	case "30": // Movies
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "31": // Anime/Animation
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "32": // Action/Adventure
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "33": // Classics
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "34": // Comedy
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "35": // Documentary
		handleTopicId = true
	case "36": // Drama
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "37": // Family
		request.SetClassification(sis.ScrollResponseUDC_PersonalLog)
	case "38": // Foreign
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "39": // Horror
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "40": // Sci-Fi/Fantasy
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "41": // Thriller
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "42": // Shorts
		handleTopicId = true
	case "43": // Shows
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	case "44": // Trailers
		request.SetClassification(sis.ScrollResponseUDC_Entertainment)
	}

	if handleTopicId {
	outer:
		for _, topic := range video.TopicDetails.TopicIds {
			switch topic {
			case "/m/01k8wb":
				request.SetClassification(sis.ScrollResponseUDC_GeneralKnowledge)
				break outer
			case "/m/04rlf", "/m/02mscn", "/m/0ggq0m", "/m/01lyv", "/m/02lkt", "/m/0glt670", "/m/05rwpb", "/m/03_d0", "/m/028sqc", "/m/0g293", "/m/064t9", "/m/06cqb", "/m/06j6l", "/m/06by7", "/m/0gywn":
				request.SetClassification(sis.ScrollResponseUDC_Music)
				break outer
			case "/m/0bzvm2", "/m/025zzc", "/m/02ntfj", "/m/0b1vjn", "/m/02hygl", "/m/04q1x3q", "/m/01sjng", "/m/0403l3g", "/m/021bp2", "/m/022dc6", "/m/03hf_rm": // Gaming
				request.SetClassification(sis.ScrollResponseUDC_GamingVideos)
				break outer
			case "/m/06ntj", "/m/0jm_", "/m/018jz", "/m/018w8", "/m/01cgz", "/m/09xp_", "/m/02vx4", "/m/037hz", "/m/03tmr", "/m/01h7lh", "/m/0410tth", "/m/07bs0", "/m/07_53": // Sports
				request.SetClassification(sis.ScrollResponseUDC_Sport)
				break outer
			case "/m/02jjt", "/m/09kqc", "/m/02vxn", "/m/05qjc", "/m/066wd", "/m/0f2f9": // Entertainment
				request.SetClassification(sis.ScrollResponseUDC_Entertainment)
				break outer
			case "/m/032tl": // Fashion -> Art
				request.SetClassification(sis.ScrollResponseUDC_Art)
				break outer
			case "/m/027x7n": // Fitness -> Sport
				request.SetClassification(sis.ScrollResponseUDC_Sport)
				break outer
			case "/m/02wbm": // Food -> Art
				request.SetClassification(sis.ScrollResponseUDC_Art)
				break outer
			case "/m/03glg": // Hobby -> Recreation
				request.SetClassification(sis.ScrollResponseUDC_Entertainment)
				break outer
			case "/m/068hy": // Pets -> Recreation
				request.SetClassification(sis.ScrollResponseUDC_Entertainment)
				break outer
			case "/m/041xxh": // Beauty
				request.SetClassification(sis.ScrollResponseUDC_Art)
				break outer
			case "/m/07c1v": // Computer Technology
				request.SetClassification(sis.ScrollResponseUDC_Class0)
				break outer
			case "/m/07bxq": // Tourism -> Recreation
				request.SetClassification(sis.ScrollResponseUDC_Entertainment)
				break outer
			case "/m/07yv9": // Vehicles -> Engineering/General Technology
				request.SetClassification(sis.ScrollResponseUDC_Engineering)
				break outer
			case "/m/06bvp": // Religion
				request.SetClassification(sis.ScrollResponseUDC_Religion)
				break outer
			case "/m/05qt0": // Politics
				request.SetClassification(sis.ScrollResponseUDC_SocialScience)
				break outer
			case "/m/01h6rj": // Military
				request.SetClassification(sis.ScrollResponseUDC_SocialScience)
				break outer
			case "/m/0kt51": // Health
				request.SetClassification(sis.ScrollResponseUDC_Medicine)
				break outer
			case "/m/09s1f": // Business
				request.SetClassification(sis.ScrollResponseUDC_AppliedScience)
				break outer
			case "/m/098wr", "/m/019_rr":
				request.SetClassification(sis.ScrollResponseUDC_SocialScience)
			}
		}
	}
}

func getVideoPageRouteFunc(service *youtube.Service) sis.RequestHandler {
	return func(request *sis.Request) {
		id := request.GetParam("id")
		call := service.Videos.List([]string{"id", "snippet", "status"}).Id(id).MaxResults(1)
		response, err := call.Do()
		if err != nil {
			//log.Fatalf("Error: %v", err) // TODO
			panic(err)
		}

		if len(response.Items) == 0 {
			request.TemporaryFailure("Video not found.")
			return
		}
		video := response.Items[0]
		handleVideoClassification(video, request)

		lang := request.Server.DefaultLanguage()
		if video.Snippet.DefaultLanguage != "" {
			lang = video.Snippet.DefaultLanguage
		}
		publishDate := video.Snippet.PublishedAt
		if video.Status.PrivacyStatus == "private" {
			publishDate = video.Status.PublishAt
		}
		publishDateParsed, _ := time.Parse(time.RFC3339, publishDate)
		abstract := fmt.Sprintf("# Video - %s\n%s\n", html.UnescapeString(video.Snippet.Title), html.UnescapeString(video.Snippet.Description))
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: html.UnescapeString(video.Snippet.ChannelTitle), PublishDate: publishDateParsed.UTC(), UpdateDate: publishDateParsed.UTC(), Language: lang, Abstract: abstract})
		if request.ScrollMetadataRequested() {
			request.Scroll(abstract)
			return
		}

		//video.ContentDetails.RegionRestriction.Allowed

		//video.ContentDetails.Definition

		var downloadFormatsBuilder strings.Builder
		var captionsBuilder strings.Builder
		client := ytd.Client{}
		ytd_vid, err := client.GetVideo(video.Id)
		retries := 0
		for err != nil {
			// Try again, for a maximum of 5 times.
			ytd_vid, err = client.GetVideo(video.Id)
			retries += 1
			if retries == 5 {
				break
			}
			time.Sleep(time.Millisecond * 120)
		}
		if err != nil { // If still getting an error after retrying 5 times.
			fmt.Printf("Couldn't find video in ytd client.\n")
			fmt.Fprintf(&downloadFormatsBuilder, "No downloads available yet. Try again later.\n")
		} else {
			// List Download Formats
			formats := ytd_vid.Formats.WithAudioChannels().Type("video/mp4") // TODO
			formats.Sort()
			if len(formats) == 0 {
				fmt.Fprintf(&downloadFormatsBuilder, "No downloads available yet. The video could be a future livestream or premiere.\n")
			} else {
				for _, format := range formats {
					audioQuality := ""
					switch format.AudioQuality {
					case "AUDIO_QUALITY_HIGH":
						audioQuality = "High Audio Quality"
					case "AUDIO_QUALITY_MEDIUM":
						audioQuality = "Medium Audio Quality"
					case "AUDIO_QUALITY_LOW":
						audioQuality = "Low Audio Quality"
					}
					fmt.Fprintf(&downloadFormatsBuilder, "=> /youtube/downloadVideo/%s/%s.mp4 Download Video - %s (%s)\n", format.Quality, video.Id, format.Quality, audioQuality)
				}
			}

			_, transcript_err := client.GetTranscript(ytd_vid, "en")
			if !errors.Is(transcript_err, ytd.ErrTranscriptDisabled) {
				fmt.Fprintf(&captionsBuilder, "=> /youtube/video/%s/transcript/ View Video Transcript\n\n", video.Id)
			}

			// Captions
			if len(ytd_vid.CaptionTracks) > 0 {
				fmt.Fprintf(&captionsBuilder, "## Captions\n")
				for _, caption := range ytd_vid.CaptionTracks {
					captionString := caption.LanguageCode + ".srv3"
					if caption.Kind != "" {
						captionString = caption.Kind + "_" + captionString
					}
					fmt.Fprintf(&captionsBuilder, "=> /youtube/video/%s/caption/%s %s %s\n", video.Id, url.PathEscape(captionString), caption.Kind, caption.LanguageCode)
				}
			}
		}

		request.Gemini(fmt.Sprintf(`# Video: %s

%s
=> https://youtube.com/watch?v=%s On YouTube
=> gopher://auragem.ddns.net/1/g/youtube/video/%s/ On Gopher
=> spartan://auragem.ddns.net/g/youtube/video/%s/ On Spartan
=> nex://auragem.ddns.net/gemini/youtube/video/%s/ On Nex

%s

## Description
%s
=> /youtube/channel/%s/ Uploaded by %s
`, html.UnescapeString(video.Snippet.Title), downloadFormatsBuilder.String() /*video.Id, */, video.Id, video.Id, video.Id, video.Id, captionsBuilder.String(), html.UnescapeString(video.Snippet.Description), video.Snippet.ChannelId, html.UnescapeString(video.Snippet.ChannelTitle)))
	}
}

func handleCaptionDownload(s sis.VirtualServerHandle) {
	s.AddRoute("/youtube/video/:id/transcript", func(request *sis.Request) {
		client := ytd.Client{}
		videoId := request.GetParam("id")
		video, err := client.GetVideo(videoId)
		retries := 0
		for err != nil {
			// Try again, for a maximum of 5 times.
			video, err = client.GetVideo(videoId)
			retries += 1
			if retries == 5 {
				break
			}
			time.Sleep(time.Millisecond * 120)
		}
		if err != nil {
			//panic(err)
			request.TemporaryFailure("Error: Couldn't find video. %s\n", err.Error())
			return
		}

		time.Sleep(time.Millisecond * 120)

		// Go through each requested language and try to find a transcript for them. Otherwise, fallback to english.
		// If still no transcript found, then error out.
		requestedLanguages := append(request.ScrollRequestedLanguages, "en") // Append fallback language
		var transcript ytd.VideoTranscript
		var found bool = false
		for _, lang := range requestedLanguages {
			var err error
			transcript, err = client.GetTranscript(video, lang)
			if err == nil {
				request.SetLanguage(lang)
				found = true
				break
			}
		}
		if !found {
			request.TemporaryFailure("Video doesn't have a transcript.\n")
			return
		}

		request.Gemini(transcript.String())
	})

	s.AddRoute("/youtube/video/:id/caption/:caption", func(request *sis.Request) {
		client := ytd.Client{}
		videoId := request.GetParam("id")
		video, err := client.GetVideo(videoId)
		retries := 0
		for err != nil {
			// Try again, for a maximum of 5 times.
			video, err = client.GetVideo(videoId)
			retries += 1
			if retries == 5 {
				break
			}
			time.Sleep(time.Millisecond * 120)
		}

		if err != nil {
			//panic(err)
			request.TemporaryFailure("Error: Couldn't find video. %s\n", err.Error())
			return
		}

		captionString := request.GetParam("caption")
		kind, lang, foundKind := strings.Cut(captionString, "_")
		if !foundKind {
			lang = kind
			kind = ""
		}

		lang = strings.TrimSuffix(lang, ".srv3")
		fmt.Printf("Getting caption using kind %s and lang %s\n", kind, lang)

		var foundCaption = false
		var captionFound ytd.CaptionTrack
		for _, caption := range video.CaptionTracks {
			if caption.Kind == kind && caption.LanguageCode == lang {
				captionFound = caption
				foundCaption = true
				request.SetLanguage(caption.LanguageCode)
				break
			}
		}
		if !foundCaption {
			request.TemporaryFailure("Caption not found.")
			return
		} else {
			http_client := http.DefaultClient
			response, err := http_client.Get(captionFound.BaseURL)
			if err != nil {
				request.TemporaryFailure("Couldn't download caption file.")
				return
			} else {
				request.Stream("text/xml; charset=UTF-8", response.Body)
				response.Body.Close()
			}
		}
	})
}

func filterYT(fl ytd.FormatList, test func(ytd.Format) bool) ytd.FormatList {
	var ret []ytd.Format
	for _, format := range fl {
		if test(format) {
			ret = append(ret, format)
		}
	}
	return ytd.FormatList(ret)
}

func getVideoDownloadRouteFunc() sis.RequestHandler {
	ipsDownloading := make(map[string]struct{})
	videoQualities := []string{"hd1080", "hd720", "medium", "tiny"}

	return func(request *sis.Request) {
		_, ok := ipsDownloading[request.IPHash()]
		if ok {
			request.TemporaryFailure("You are already downloading a video from the proxy. Please wait until that is finished before downloading another.\n")
			return
		}
		ipsDownloading[request.IPHash()] = struct{}{}
		defer func() {
			delete(ipsDownloading, request.IPHash())
		}()
		desiredMaxQuality := request.GetParam("quality")

		idStr := request.GetParam("id")
		extension := path.Ext(idStr)
		videoId := strings.TrimSuffix(idStr, "."+extension)

		client := ytd.Client{}
		client.HTTPClient = &http.Client{Transport: &http.Transport{
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}}
		video, err := client.GetVideo(videoId)
		retries := 0
		for err != nil {
			// Try again, for a maximum of 5 times.
			video, err = client.GetVideo(videoId)
			retries += 1
			if retries == 5 {
				break
			}
			time.Sleep(time.Millisecond * 120)
		}

		// If still can't get video after 5 retries, then error out
		if err != nil {
			//panic(err)
			request.TemporaryFailure("Error: Couldn't download video. %s\n", err.Error())
			return
		}

		audioFormats := video.Formats.WithAudioChannels()
		audioFormats.Sort()

		audioFormats_mediumAudioQuality := filterYT(video.Formats, func(format ytd.Format) bool {
			return format.AudioQuality == "AUDIO_QUALITY_MEDIUM"
		})
		audioFormats_lowAudioQuality := filterYT(video.Formats, func(format ytd.Format) bool {
			return format.AudioQuality == "AUDIO_QUALITY_LOW"
		})

		//fmt.Printf("Formats: %v\n", audioFormats)

		var format *ytd.Format = nil
		skip := true
		for _, quality := range videoQualities {
			if quality == desiredMaxQuality {
				skip = false
			} else if skip {
				continue
			}

			// Try medium audio quality first
			//format = audioFormats_mediumAudioQuality.FindByQuality(quality)
			var list ytd.FormatList = audioFormats_mediumAudioQuality.Quality(quality)
			if len(list) <= 0 {
				fmt.Printf("Could not find %s-quality video with medium audio. Trying low audio quality.\n", quality)
				// If not found, then try low audio quality
				//format = audioFormats_lowAudioQuality.FindByQuality(quality)
				list = audioFormats_lowAudioQuality.Quality(quality)
				if len(list) <= 0 {
					fmt.Printf("Could not find %s-quality video with audio. Trying next quality.\n", quality)
					continue
				}
			}

			// If a format was found, break
			format = &list[0]
			break
		}

		if format == nil {
			// No format was found with video and audio
			fmt.Printf("Could not find any video with audio at or below desired quality of %s.\n", desiredMaxQuality)
			if desiredMaxQuality != "hd720" {
				request.TemporaryFailure("Video at or below desired quality of %s not found. Try a higher quality.\n", desiredMaxQuality)
				return
				//return c.Gemini("Error: Video At or Below Desired Quality of %s Not Found. Try a higher quality.\n%v", desiredMaxQuality, err) // TODO: Do different thing?
			} else {
				request.TemporaryFailure("Video with Audio not found. The video could be a future premiere or livestream.\n")
				return
				//return c.Gemini("Error: Video with Audio Not Found.\n%v", err) // TODO: Do different thing?
			}
		}

		// Handle Scroll protocol Metadata
		abstract := fmt.Sprintf("# %s\n%s\n", video.Title, video.Description)
		// TODO: The language should be a BCP47 string
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: video.Author, PublishDate: video.PublishDate.UTC(), Language: format.LanguageDisplayName(), Abstract: abstract})
		if request.ScrollMetadataRequested() {
			request.Scroll(abstract)
			return
		}

		//format := video.Formats.AudioChannels(2).FindByQuality("hd1080")
		//.DownloadSeparatedStreams(ctx, "", video, "hd1080", "mp4")
		//resp, err := client.GetStream(video, format)

		rc, _, err := client.GetStream(video, format)
		if err != nil {
			request.TemporaryFailure("Video not found.\n")
			return
			//return c.Gemini("Error: Video Not Found\n%v", err)
		}
		request.StreamBuffer(format.MimeType, rc, make([]byte, 1*1024*1024)) // 1 MB Buffer
		//err2 := c.Stream(format.MimeType, rc)
		rc.Close()

		//url, err := client.GetStreamURL(video, format)

		//return err2
	}
}

func handleChannelPage(g sis.VirtualServerHandle, service *youtube.Service) {
	// Channel Home
	g.AddRoute("/youtube/channel/:id", func(request *sis.Request) {
		template := `# Channel: %s

=> /youtube/channel/%s/videos/ All Videos
=> /youtube/channel/%s/playlists/ Playlists
=> /youtube/channel/%s/communityposts/ Community Posts
=> /youtube/channel/%s/activity/ Gemini Sub Activity Feed

## About
%s

## Recent Videos

=> /youtube/channel/%s/videos/ All Videos
`
		call := service.Channels.List([]string{"id", "snippet", "contentDetails"}).Id(request.GetParam("id")).MaxResults(1)
		response, err := call.Do()
		if err != nil {
			//log.Fatalf("Error: %v", err) // TODO
			panic(err)
		}

		channel := response.Items[0]

		// Handle Scroll Protocol Metadata
		abstract := fmt.Sprintf("# Channel: %s\n%s\n", html.UnescapeString(channel.Snippet.Title), html.UnescapeString(channel.Snippet.Description))
		request.SetScrollMetadataResponse(sis.ScrollMetadata{Author: html.UnescapeString(channel.Snippet.Title), Language: channel.Snippet.DefaultLanguage, Abstract: abstract})
		if request.ScrollMetadataRequested() {
			request.Scroll(abstract)
			return
		}

		request.Gemini(fmt.Sprintf(template, html.UnescapeString(channel.Snippet.Title), channel.Id, channel.Id, channel.Id, channel.Id, html.UnescapeString(channel.Snippet.Description), channel.Id))
	})

	// Channel Playlists
	g.AddRoute("/youtube/channel/:id/playlists/:page", func(request *sis.Request) {
		getChannelPlaylists(request, service, request.GetParam("id"), request.GetParam("page"))
	})
	g.AddRoute("/youtube/channel/:id/playlists", func(request *sis.Request) {
		getChannelPlaylists(request, service, request.GetParam("id"), "")
	})

	// Channel Videos/Uploads
	g.AddRoute("/youtube/channel/:id/videos/:page", func(request *sis.Request) {
		getChannelVideos(request, service, request.GetParam("id"), request.GetParam("page"))
	})
	g.AddRoute("/youtube/channel/:id/videos", func(request *sis.Request) {
		getChannelVideos(request, service, request.GetParam("id"), "")
	})

	g.AddRoute("/youtube/channel/:id/activity", func(request *sis.Request) {
		getChannelActivity(request, service, request.GetParam("id"))
	})
}

func handlePlaylistPage(g sis.VirtualServerHandle, service *youtube.Service) {
	g.AddRoute("/youtube/playlist/:id/:page", func(request *sis.Request) {
		getPlaylistVideos(request, service, request.GetParam("id"), request.GetParam("page"))
	})
	g.AddRoute("/youtube/playlist/:id", func(request *sis.Request) {
		getPlaylistVideos(request, service, request.GetParam("id"), "")
	})
}
