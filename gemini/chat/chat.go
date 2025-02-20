package chat

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/krayzpipes/cronticker/cronticker"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

type ChatContext struct {
	readCond  *sync.Cond
	changeInt int64
	mutex     sync.RWMutex
	//messages     *deque.Deque[ChatText]
	messages     []ChatText
	clientNumber atomic.Int64
	sendChan     chan ChatText
}

func NewChatContext() *ChatContext {
	context := &ChatContext{
		changeInt: 0,
		mutex:     sync.RWMutex{},
		//messages:     deque.New[ChatText](0, 30),
		messages:     make([]ChatText, 0, 100),
		clientNumber: atomic.Int64{},
	}
	context.readCond = sync.NewCond(context.mutex.RLocker())
	context.sendChan = make(chan ChatText, 500)

	// Message sending handler
	go func() {
		//ticker := time.NewTicker(time.Hour * 24)
		// minute hour day-of-month month day-of-week
		ticker, err := cronticker.NewTicker("@daily") // UTC
		if err != nil {
			// TODO
			fmt.Printf("Error creating cronticker: %s\n", err.Error())
		}
		for {
			select {
			case msg := <-context.sendChan:
				{
					context.mutex.Lock()
					{
						//context.messages.PushBack(msg)
						context.messages = append(context.messages, msg)
						//context.newMsgIndex = context.messages.Len()
						context.changeInt += 1
						context.readCond.Broadcast()
					}
					context.mutex.Unlock()
				}
			case t := <-ticker.C:
				{
					// Clear messages history every day
					context.mutex.Lock()
					//context.messages.Clear()
					context.messages = context.messages[:0]
					context.mutex.Unlock()

					context.sendChan <- (ChatText{"SYSTEM", fmt.Sprintf("New Day: %s UTC\n", t.UTC().Format("2006-01-02")), t, ""})
				}
			}
			time.Sleep(time.Second / 500)
		}
	}()

	return context
}

type ChatText struct {
	username string
	text     string
	t        time.Time
	to_user  string
}

type UserInfo struct {
	username     string
	clientNumber atomic.Int64
}

func (context *ChatContext) Attach(s sis.VirtualServerHandle) {
	publishDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-19T13:51:00", time.Local)
	updateDate, _ := time.ParseInLocation(time.RFC3339, "2024-03-19T13:51:00", time.Local)

	s.AddRoute("/chat/", func(request *sis.Request) {
		query, err := request.Query()
		if err != nil {
			request.TemporaryFailure("%s", err.Error())
			return
		} else if query == "" {
			request.RequestInput("Username? ")
			return
		} else {
			//return c.NoContent(gig.StatusRedirectTemporary, "/chat/"+url.PathEscape(query))
			request.Redirect("/chat/%s", url.PathEscape(query))
			return
		}
	})

	s.AddRoute("/chat/:username", func(request *sis.Request) {
		username := request.GetParam("username")
		if username == "" {
			request.Redirect("/chat/")
			return
		}
		request.SetScrollMetadataResponse(sis.ScrollMetadata{PublishDate: publishDate, UpdateDate: updateDate, Language: "en", Abstract: "# AuraGem Live Chat\nThis chat is heavily inspired by Mozz's chat, but the UI has been tailored for most gemini browsers. Message history is cleared every 24 hours. This chat makes use of keepalive packets so that clients (that support them) will not timeout.\n"})
		if request.ScrollMetadataRequested() {
			request.SendAbstract("")
			return
		}
		var builder strings.Builder
		fmt.Fprintf(&builder, "# Live Chat\nThis chat is heavily inspired by Mozz's chat, but the UI has been tailored for most gemini browsers. Message history is cleared every 24 hours. This chat makes use of keepalive packets so that clients (that support them) will not timeout.\n=> gemini://chat.mozz.us/ Mozz's Chat\n\n")
		fmt.Fprintf(&builder, "=> /chat/%s/send Send Message\n=> titan://auragem.ddns.net/chat/%s/send Send Message via Titan\n\n", url.PathEscape(username), url.PathEscape(username))
		context.clientNumber.Add(1)

		// Print out all messages in current history
		context.readCond.L.Lock() // RLock
		length := len(context.messages)
		//fmt.Fprintf(&builder, "Initial number of messages: %d\nCurrent Change Index: %d\n", length, context.changeInt)
		for i := 0; i < length; i++ {
			msg := &context.messages[i]
			if msg.to_user != "" && msg.to_user != username {
				continue
			}
			fmt.Fprintf(&builder, "--[ %-25s @ %s ]---------------------\n %s\n\n", msg.username, msg.t.Format("2006-01-02 03:04:05 PM"), msg.text)
		}
		err := request.Gemini(builder.String())
		if err != nil {
			context.clientNumber.Add(-1)
			context.readCond.L.Unlock()
			return // err
		}

		currentChangeInt := context.changeInt
		context.readCond.L.Unlock() // RUnlock

	outer:
		for {
			context.readCond.L.Lock() // RLock
			for currentChangeInt == context.changeInt {
				//fmt.Printf("Waiting.\n")
				context.readCond.Wait()
			}
			numOfMessagesAdded := int(context.changeInt - currentChangeInt)
			dequeLength := len(context.messages)
			currentChangeInt = context.changeInt

			for i := int(math.Max(float64(dequeLength-numOfMessagesAdded), 0)); i < dequeLength; i++ {
				msg := &context.messages[i]
				if msg.to_user != "" && msg.to_user != username {
					continue
				}
				err := request.Gemini(fmt.Sprintf("--[ %-25s @ %s ]---------------------\n %s\n\n", msg.username, msg.t.Format("2006-01-02 03:04:05 PM"), msg.text))
				if err != nil {
					context.readCond.L.Unlock() // RUnlock
					break outer
				}
			}

			context.readCond.L.Unlock() // RUnlock
		}

		context.clientNumber.Add(-1)
		/*{
			once_timer := time.NewTimer(time.Minute * 1)
			<-once_timer.C
			// Check if user still has a connection. If not, send a disconnect message and remove user from map.
			usersMap.Get(username)
		}*/
	})

	sendFunc := func(request *sis.Request) {
		username := request.GetParam("username")
		var message string = ""
		if username == "" {
			//return c.NoContent(gig.StatusRedirectTemporary, "/chat/")
			request.Redirect("/chat/")
		}
		if request.Upload() {
			//mimetype, hasMimetype := c.Get("mime").(string)
			mimetype := request.DataMime
			if mimetype != "text/gemini" && mimetype != "text/plain" && mimetype != "" {
				request.TemporaryFailure("Can only upload text/gemini messages at the moment. File upload is coming soon.")
				return
			} else if request.DataSize > 1024*8 { // 8 KB max
				request.TemporaryFailure("Message too long.")
				return
			}
			data, read_err := request.GetUploadData()
			if read_err != nil {
				return
			}
			message = string(data)
		} else {
			request.SetSpartanQueryLimit(1024 * 8) // For proxying to Spartan
			query, err := request.Query()
			if err != nil {
				request.TemporaryFailure("%s", err.Error())
				return
			} else if query == "" {
				request.RequestInput("Message: ")
				//return c.NoContent(gig.StatusInput, "Message: ")
				return
			} else {
				message = query
			}
		}

		// TODO: Max line length? If go over, then switch to uploading it as a file.

		message = strings.TrimPrefix(message, "####")
		message = strings.TrimPrefix(message, "###")
		message = strings.TrimPrefix(message, "##")
		message = strings.TrimPrefix(message, "#")
		message = strings.TrimPrefix(message, "-[")
		message = strings.ReplaceAll(message, "\n####", "")
		message = strings.ReplaceAll(message, "\n###", "")
		message = strings.ReplaceAll(message, "\n##", "")
		message = strings.ReplaceAll(message, "\n#", "")
		message = strings.ReplaceAll(message, "\n-[", "")

		if !request.ScrollMetadataRequested() {
			context.sendChan <- (ChatText{username, message, time.Now(), ""})
		}
		//return c.NoContent(gig.StatusRedirectTemporary, "gemini://auragem.ddns.net/chat/"+url.PathEscape(username))
		request.Redirect("%sauragem.ddns.net/chat/%s", request.Server.Scheme(), url.PathEscape(username))
	}

	s.AddRoute("/chat/:username/send", sendFunc)
	s.AddUploadRoute("/chat/:username/send", sendFunc) // Titan Upload
}
