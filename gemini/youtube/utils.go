package youtube

import (
	"fmt"
	"html"
	"strings"
	"time"

	sis "gitlab.com/clseibold/smallnetinformationservices"
	"google.golang.org/api/youtube/v3"
)

func searchYoutube(request sis.Request, service *youtube.Service, query string, rawQuery string, currentPage string) {
	template := `# Search

=> /youtube/ Home
=> /youtube/search/ New Search
%s`

	var call *youtube.SearchListCall
	if currentPage != "" {
		call = service.Search.List([]string{"snippet"}).Q(query).MaxResults(maxResults).PageToken(currentPage)
	} else {
		call = service.Search.List([]string{"snippet"}).Q(query).MaxResults(maxResults)
	}

	response, err := call.Do()
	if err != nil {
		//log.Fatalf("Error: %v", err)
		panic(err)
	}

	var builder strings.Builder
	if response.PrevPageToken != "" {
		fmt.Fprintf(&builder, "=> /youtube/search/%s/?%s Previous Page\n", response.PrevPageToken, rawQuery)
	}
	if response.NextPageToken != "" {
		fmt.Fprintf(&builder, "=> /youtube/search/%s/?%s Next Page\n", response.NextPageToken, rawQuery)
	}
	fmt.Fprintf(&builder, "\n## %d/%d Results for '%s'\n\n", response.PageInfo.ResultsPerPage, response.PageInfo.TotalResults, query)

	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			fmt.Fprintf(&builder, "=> /youtube/video/%s/ Video: %s\nUploaded by %s\n\n", item.Id.VideoId, html.UnescapeString(item.Snippet.Title), html.UnescapeString(item.Snippet.ChannelTitle))
		case "youtube#channel":
			fmt.Fprintf(&builder, "=> /youtube/channel/%s/ Channel: %s\n\n", item.Id.ChannelId, html.UnescapeString(item.Snippet.Title))
		case "youtube#playlist":
			fmt.Fprintf(&builder, "=> /youtube/playlist/%s/ Playlist: %s\n\n", item.Id.PlaylistId, html.UnescapeString(item.Snippet.Title))
		}
	}

	request.Gemini(fmt.Sprintf(template, builder.String()))
}

func getChannelPlaylists(request sis.Request, service *youtube.Service, channelId string, currentPage string) {
	template := `# Playlists for '%s'

=> /youtube/channel/%s/ ChannelPage
%s`

	var call *youtube.PlaylistsListCall
	if currentPage != "" {
		call = service.Playlists.List([]string{"id", "snippet"}).ChannelId(channelId).MaxResults(50).PageToken(currentPage)
	} else {
		call = service.Playlists.List([]string{"id", "snippet"}).ChannelId(channelId).MaxResults(50)
	}
	response, err := call.Do()
	if err != nil {
		//log.Fatalf("Error: %v", err)
		panic(err)
	}

	var builder strings.Builder
	if response.PrevPageToken != "" {
		fmt.Fprintf(&builder, "=> /youtube/channel/%s/playlists/%s/ Previous Page\n", channelId, response.PrevPageToken)
	}
	if response.NextPageToken != "" {
		fmt.Fprintf(&builder, "=> /youtube/channel/%s/playlists/%s/ Next Page\n", channelId, response.NextPageToken)
	}
	fmt.Fprintf(&builder, "\n")

	if len(response.Items) == 0 {
		request.TemporaryFailure("This channel doesn't have any playlists.\n")
		return
	}

	for _, item := range response.Items {
		fmt.Fprintf(&builder, "=> /youtube/playlist/%s/ %s\n", item.Id, html.UnescapeString(item.Snippet.Title))
	}

	request.Gemini(fmt.Sprintf(template, html.UnescapeString(response.Items[0].Snippet.ChannelTitle), response.Items[0].Snippet.ChannelId, builder.String()))
}

func getChannelVideos(request sis.Request, service *youtube.Service, channelId string, currentPage string) {
	template := `# Uploads for '%s'

=> /youtube/channel/%s/ Channel Page
%s`

	call := service.Channels.List([]string{"id", "snippet", "contentDetails"}).Id(channelId).MaxResults(1)
	response, err := call.Do()
	if err != nil {
		//log.Fatalf("Error: %v", err)
		panic(err)
	}

	channel := response.Items[0]
	uploadsPlaylistId := channel.ContentDetails.RelatedPlaylists.Uploads
	time.Sleep(time.Millisecond * 120)

	var call2 *youtube.PlaylistItemsListCall
	if currentPage != "" {
		call2 = service.PlaylistItems.List([]string{"id", "snippet"}).PlaylistId(uploadsPlaylistId).MaxResults(25).PageToken(currentPage)
	} else {
		call2 = service.PlaylistItems.List([]string{"id", "snippet"}).PlaylistId(uploadsPlaylistId).MaxResults(25)
	}
	response2, err2 := call2.Do()
	if err2 != nil {
		//log.Fatalf("Error: %v", err)
		panic(err2)
	}

	var builder strings.Builder
	if response2.PrevPageToken != "" {
		fmt.Fprintf(&builder, "=> /youtube/channel/%s/videos/%s/ Previous Page\n", channelId, response2.PrevPageToken)
	}
	if response2.NextPageToken != "" {
		fmt.Fprintf(&builder, "=> /youtube/channel/%s/videos/%s/ Next Page\n", channelId, response2.NextPageToken)
	}
	fmt.Fprintf(&builder, "\n")

	for _, item := range response2.Items {
		date := strings.Split(item.Snippet.PublishedAt, "T")[0]
		fmt.Fprintf(&builder, "=> /youtube/video/%s %s %s\n", item.Snippet.ResourceId.VideoId, date, html.UnescapeString(item.Snippet.Title))
	}

	request.Gemini(fmt.Sprintf(template, html.UnescapeString(channel.Snippet.Title), channel.Id, builder.String()))
}

func getChannelActivity(request sis.Request, service *youtube.Service, channelId string) {
	template := `# Activity for '%s'

%s`

	call := service.Channels.List([]string{"id", "snippet", "contentDetails"}).Id(channelId).MaxResults(1)
	response, err := call.Do()
	if err != nil {
		//log.Fatalf("Error: %v", err)
		//panic(err)
		request.TemporaryFailure("Failed to get channel info.")
		return
	}

	channel := response.Items[0]
	uploadsPlaylistId := channel.ContentDetails.RelatedPlaylists.Uploads

	time.Sleep(time.Millisecond * 120)
	call2 := service.PlaylistItems.List([]string{"id", "snippet"}).PlaylistId(uploadsPlaylistId).MaxResults(100) // TODO
	response2, err2 := call2.Do()
	if err2 != nil {
		//log.Fatalf("Error: %v", err)
		//panic(err)
		request.TemporaryFailure("Failed to get channel activity.")
		return
	}

	var builder strings.Builder
	for _, item := range response2.Items {
		date := strings.Split(item.Snippet.PublishedAt, "T")[0]
		fmt.Fprintf(&builder, "=> /youtube/video/%s/ %s %s\n", item.Snippet.ResourceId.VideoId, date, html.UnescapeString(item.Snippet.Title))
	}

	request.Gemini(fmt.Sprintf(template, html.UnescapeString(channel.Snippet.Title), builder.String()))
}

func getPlaylistVideos(request sis.Request, service *youtube.Service, playlistId string, currentPage string) {
	template := `# Playlist: %s

=> /youtube/channel/%s/ Created by %s
%s`

	call_pl := service.Playlists.List([]string{"id", "snippet"}).Id(playlistId).MaxResults(1)
	response_pl, err_pl := call_pl.Do()
	if err_pl != nil {
		//log.Fatalf("Error: %v", err_pl)
		//panic(err_pl)
		request.TemporaryFailure("Failed to get playlist.")
		return
	}

	playlist := response_pl.Items[0]
	playlistTitle := playlist.Snippet.Title

	time.Sleep(time.Millisecond * 120)
	var call *youtube.PlaylistItemsListCall
	if currentPage != "" {
		call = service.PlaylistItems.List([]string{"id", "snippet"}).PlaylistId(playlistId).MaxResults(50).PageToken(currentPage)
	} else {
		call = service.PlaylistItems.List([]string{"id", "snippet"}).PlaylistId(playlistId).MaxResults(50)
	}
	response, err := call.Do()
	if err != nil {
		//log.Fatalf("Error: %v", err)
		//panic(err)
		request.TemporaryFailure("Failed to get playlist videos.")
		return
	}

	var builder strings.Builder
	if response.PrevPageToken != "" {
		fmt.Fprintf(&builder, "=> /youtube/playlist/%s/%s/ Previous Page\n", playlistId, response.PrevPageToken)
	}
	if response.NextPageToken != "" {
		fmt.Fprintf(&builder, "=> /youtube/playlist/%s/%s/ Next Page\n", playlistId, response.NextPageToken)
	}
	fmt.Fprintf(&builder, "\n")

	for _, item := range response.Items {
		fmt.Fprintf(&builder, "=> /youtube/video/%s/ %s\nUploaded by %s\n\n", item.Snippet.ResourceId.VideoId, html.UnescapeString(item.Snippet.Title), html.UnescapeString(item.Snippet.ChannelTitle))
	}

	request.Gemini(fmt.Sprintf(template, playlistTitle, playlist.Snippet.ChannelId, html.UnescapeString(playlist.Snippet.ChannelTitle), builder.String()))
}
