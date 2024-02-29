package youtube

import (
	"context"
	"embed"
	"fmt"
	"html"

	//"log"
	"strings"

	ytd "github.com/kkdai/youtube/v2"
	"gitlab.com/clseibold/auragem_sis/config"
	sis "gitlab.com/clseibold/smallnetinformationservices"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	youtubeAPIKey = config.YoutubeApiKey
	maxResults    = int64(25) /*flag.Int64("max-results", 25, "Max YouTube results")*/
)

//go:embed index.gmi
var content embed.FS

func HandleYoutube(g sis.ServerHandle) {
	// Create Youtube Service
	service, err1 := youtube.NewService(context.Background(), option.WithAPIKey(youtubeAPIKey))
	if err1 != nil {
		//log.Fatalf("Error creating new Youtube client: %v", err1)
		panic(err1)
	}
	searchRoute := getSearchRouteFunc(service)
	videoPageRoute := getVideoPageRouteFunc(service)
	videoDownloadRoute := getVideoDownloadRouteFunc(service)

	g.AddRoute("/cgi-bin/youtube.cgi", func(request sis.Request) {
		request.Redirect("/youtube") // TODO: Temporary Redirect
	})
	g.AddRoute("/youtube", indexRoute)
	g.AddRoute("/youtube/search", searchRoute)
	g.AddRoute("/youtube/search/:page", searchRoute)
	g.AddRoute("/youtube/video/:id/", videoPageRoute)
	g.AddRoute("/youtube/downloadVideo/:quality/:id", videoDownloadRoute)

	handleChannelPage(g, service)
	handlePlaylistPage(g, service)
}

func indexRoute(request sis.Request) {
	index, _ := content.ReadFile("index.gmi")
	request.Gemini(string(index))
}
func getSearchRouteFunc(service *youtube.Service) sis.RequestHandler {
	return func(request sis.Request) {
		query := request.Query()

		if query == "" {
			request.RequestInput("Search Query:")
		} else {
			rawQuery := request.RawQuery()
			page := request.GetParam("page")
			if page == "" {
				searchYoutube(request, service, query, rawQuery, "")
			} else {
				searchYoutube(request, service, query, rawQuery, page)
			}
		}
	}
}

func getVideoPageRouteFunc(service *youtube.Service) sis.RequestHandler {
	return func(request sis.Request) {
		id := request.GetParam("id")
		call := service.Videos.List([]string{"id", "snippet"}).Id(id).MaxResults(1)
		response, err := call.Do()
		if err != nil {
			//log.Fatalf("Error: %v", err) // TODO
			panic(err)
		}

		if len(response.Items) == 0 {
			request.TemporaryFailure("Video not found.")
			return
			//return c.NoContent(gig.StatusTemporaryFailure, "Video not found.")
		}
		video := response.Items[0] // TODO: Error if video is not found

		caption_call := service.Captions.List([]string{"id", "snippet"}, id)
		caption_response, err := caption_call.Do()

		var captionsBuilder strings.Builder
		for _, caption := range caption_response.Items {
			fmt.Fprintf(&captionsBuilder, "=> /youtube/caption/%s %s", caption.Id, caption.Snippet.Name)
		}

		request.Gemini(fmt.Sprintf(`# Video: %s

=> /youtube/downloadVideo/hd720/%s Download Video - 720p
=> /youtube/downloadVideo/medium/%s Download Video - medium
=> https://youtube.com/watch?v=%s On YouTube

## Caption Transcripts
%s

## Description
%s
=> /youtube/channel/%s Uploaded by %s
`, html.UnescapeString(video.Snippet.Title), video.Id, video.Id /*video.Id, */, video.Id, captionsBuilder.String(), html.UnescapeString(video.Snippet.Description), video.Snippet.ChannelId, html.UnescapeString(video.Snippet.ChannelTitle)))
	}
}

func handleCaptionPage(g *sis.Server, service *youtube.Service) {
	g.AddRoute("/youtube/caption/:id", func(request sis.Request) {
		//caption_call := service.Captions.Download(c.Param("id"))
		//caption_response, err := caption_call.Do()

		request.TemporaryFailure("Unfinished.")
		//return c.Gemini(`Video: %s, Caption: %s`, "", "")
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

func getVideoDownloadRouteFunc(service *youtube.Service) sis.RequestHandler {
	videoQualities := []string{"hd1080", "hd720", "medium", "tiny"}

	return func(request sis.Request) {
		desiredMaxQuality := request.GetParam("quality")
		client := ytd.Client{}
		video, err := client.GetVideo(request.GetParam("id"))
		if err != nil {
			//panic(err)
			request.TemporaryFailure("Error: Couldn't download video. %s\n", err.Error())
			return
		}

		//format := video.Formats.AudioChannels(2).FindByQuality("medium")
		//video.Formats.Sort()
		audioFormats := video.Formats.WithAudioChannels()
		audioFormats.Sort()

		audioFormats_mediumAudioQuality := filterYT(video.Formats, func(format ytd.Format) bool {
			return format.AudioQuality == "AUDIO_QUALITY_MEDIUM"
		})
		audioFormats_lowAudioQuality := filterYT(video.Formats, func(format ytd.Format) bool {
			return format.AudioQuality == "AUDIO_QUALITY_LOW"
		})

		//fmt.Printf("Formats: %v\n", audioFormats)

		var format *ytd.Format
		skip := true
		for _, quality := range videoQualities {
			if quality == desiredMaxQuality {
				skip = false
			} else if skip {
				continue
			}

			// Try medium audio quality first
			format = audioFormats_mediumAudioQuality.FindByQuality(quality)
			if format == nil {
				fmt.Printf("Could not find %s-quality video with medium audio. Trying low audio quality.\n", quality)
				// If not found, then try low audio quality
				format = audioFormats_lowAudioQuality.FindByQuality(quality)
				if format == nil {
					fmt.Printf("Could not find %s-quality video with audio. Trying next quality.\n", quality)
				}
			}

			// If a format was found, break
			if format != nil {
				break
			}
		}

		if format == nil {
			// No format was found with video and audio
			fmt.Printf("Could not find any video with audio at or below desired quality of %s.\n", desiredMaxQuality)
			if desiredMaxQuality != "hd720" {
				request.TemporaryFailure("Video at or below desired quality of %s not found. Try a higher quality.\n", desiredMaxQuality)
				return
				//return c.Gemini("Error: Video At or Below Desired Quality of %s Not Found. Try a higher quality.\n%v", desiredMaxQuality, err) // TODO: Do different thing?
			} else {
				request.TemporaryFailure("Video with Audio not found.\n")
				return
				//return c.Gemini("Error: Video with Audio Not Found.\n%v", err) // TODO: Do different thing?
			}
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
		request.Stream(format.MimeType, rc)
		//err2 := c.Stream(format.MimeType, rc)
		rc.Close()

		//url, err := client.GetStreamURL(video, format)

		//return err2
	}
}

func handleChannelPage(g sis.ServerHandle, service *youtube.Service) {
	// Channel Home
	g.AddRoute("/youtube/channel/:id", func(request sis.Request) {
		template := `# Channel: %s

=> /youtube/channel/%s/videos All Videos
=> /youtube/channel/%s/playlists Playlists
=> /youtube/channel/%s/communityposts Community Posts
=> /youtube/channel/%s/activity Gemini Sub Activity Feed

## About
%s

## Recent Videos

=> /youtube/channel/%s/videos All Videos
`
		call := service.Channels.List([]string{"id", "snippet", "contentDetails"}).Id(request.GetParam("id")).MaxResults(1)
		response, err := call.Do()
		if err != nil {
			//log.Fatalf("Error: %v", err) // TODO
			panic(err)
		}

		channel := response.Items[0]

		request.Gemini(fmt.Sprintf(template, html.UnescapeString(channel.Snippet.Title), channel.Id, channel.Id, channel.Id, channel.Id, html.UnescapeString(channel.Snippet.Description), channel.Id))
	})

	// Channel Playlists
	g.AddRoute("/youtube/channel/:id/playlists/:page", func(request sis.Request) {
		getChannelPlaylists(request, service, request.GetParam("id"), request.GetParam("page"))
	})
	g.AddRoute("/youtube/channel/:id/playlists", func(request sis.Request) {
		getChannelPlaylists(request, service, request.GetParam("id"), "")
	})

	// Channel Videos/Uploads
	g.AddRoute("/youtube/channel/:id/videos/:page", func(request sis.Request) {
		getChannelVideos(request, service, request.GetParam("id"), request.GetParam("page"))
	})
	g.AddRoute("/youtube/channel/:id/videos", func(request sis.Request) {
		getChannelVideos(request, service, request.GetParam("id"), "")
	})

	g.AddRoute("/youtube/channel/:id/activity", func(request sis.Request) {
		getChannelActivity(request, service, request.GetParam("id"))
	})
}

func handlePlaylistPage(g sis.ServerHandle, service *youtube.Service) {
	g.AddRoute("/youtube/playlist/:id/:page", func(request sis.Request) {
		getPlaylistVideos(request, service, request.GetParam("id"), request.GetParam("page"))
	})
	g.AddRoute("/youtube/playlist/:id", func(request sis.Request) {
		getPlaylistVideos(request, service, request.GetParam("id"), "")
	})
}
