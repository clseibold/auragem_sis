package ask

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	gemini "git.sr.ht/~adnano/go-gemini"
	gemini2 "github.com/clseibold/go-gemini"
	"gitlab.com/clseibold/auragem_sis/db"
	sis "gitlab.com/clseibold/smallnetinformationservices"
)

var registerNotification = `# AuraGem Ask

You have selected a certificate that has not been registered yet. Registering associates a username to your certificate and allows you to start posting. Please register here:

=> /~ask/?register Register Cert
`

func HandleAsk(s sis.ServerHandle) {
	conn := db.NewConn(db.AskDB)
	conn.SetMaxOpenConns(500)
	conn.SetMaxIdleConns(3)
	conn.SetConnMaxLifetime(time.Hour * 4)

	s.AddRoute("/ask", func(request sis.Request) {
		request.Redirect("/~ask/")
	})
	s.AddRoute("/~ask", func(request sis.Request) {
		request.Redirect("/~ask/")
	})
	s.AddRoute("/~ask/", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}

		if !request.HasUserCert() {
			if query == "register" || query != "" {
				request.RequestClientCert("Please enable a certificate.")
				return
			} else {
				getHomepage(request, conn)
			}
		} else {
			fmt.Printf("%s\n", query)
			user, isRegistered := GetUser(conn, request.UserCertHash())
			if !isRegistered {
				if query == "register" {
					request.RequestInput("Enter a username:")
					return
				} else if query != "" {
					// Do registration
					RegisterUser(request, conn, query, request.UserCertHash())
				} else {
					request.Gemini(registerNotification)
					return
					//return getUserDashboard(c, conn, user)
				}
			} else {
				if query == "register" {
					request.Redirect("/~ask/")
				} else {
					getUserDashboard(request, conn, user)
				}
			}
		}
	})

	s.AddRoute("/~ask/register", func(request sis.Request) {
		if !request.HasUserCert() {
			request.Redirect("/~ask/?register")
		} else {
			// Check if user already exists, if so, give error.
			_, isRegistered := GetUser(conn, request.UserCertHash())
			if isRegistered {
				request.TemporaryFailure("You are already registered. Be sure to select your certificate on the homepage.")
				return
			}

			query, err2 := request.Query()
			if err2 != nil {
				return
			} else if query == "" {
				//request.RequestInput("Enter a username:")
				request.Redirect("/~ask/?register")
			} else {
				// Do registration
				RegisterUser(request, conn, query, request.UserCertHash())
			}
		}
	})

	s.AddRoute("/~ask/recent", func(request sis.Request) {
		var recentQuestionsBuilder strings.Builder

		activities := getRecentActivity_Questions(conn)
		prevYear, prevMonth, prevDay := 0, time.Month(1), 0
		for _, activity := range activities {
			year, month, day := activity.Activity_date.Date()
			if prevYear != 0 && (year != prevYear || month != prevMonth || day != prevDay) {
				fmt.Fprintf(&recentQuestionsBuilder, "\n")
			}

			if activity.Activity == "question" {
				fmt.Fprintf(&recentQuestionsBuilder, "=> /~ask/%d/%d %s %s > %s (%s)\n", activity.Q.TopicId, activity.Q.Id, activity.Activity_date.Format("2006-01-02"), activity.TopicTitle, activity.Q.Title, activity.User.Username)
			} else if activity.Activity == "answer" {
				fmt.Fprintf(&recentQuestionsBuilder, "=> /~ask/%d/%d/a/%d %s %s > Re %s (%s)\n", activity.Q.TopicId, activity.Q.Id, activity.AnswerId, activity.Activity_date.Format("2006-01-02"), activity.TopicTitle, activity.Q.Title, activity.User.Username)
			} else {
				fmt.Fprintf(&recentQuestionsBuilder, "'%s'\n", activity.Activity)
			}

			prevYear, prevMonth, prevDay = year, month, day
		}

		request.Gemini(fmt.Sprintf(`# AuraGem Ask - Recent Activity on All Topics

=> /~ask/ AuraGem Ask Root

%s
`, recentQuestionsBuilder.String()))
	})

	s.AddRoute("/~ask/dailydigest", func(request sis.Request) {
		dates := getRecentActivity_dates(conn)

		var builder strings.Builder
		fmt.Fprintf(&builder, "# AuraGem Ask Daily Digest\n\n")
		fmt.Fprintf(&builder, "=> /~ask/ AuraGem Ask Root\n")
		query := strings.ReplaceAll(url.QueryEscape("gemini://auragem.ddns.net/~ask/dailydigest"), "%", "%%")
		fmt.Fprintf(&builder, "=> gemini://warmedal.se/~antenna/submit?%s Update Digest on Antenna\n\n", query)
		prevYear, prevMonth := 0, time.Month(1)
		for _, date := range dates {
			year, month, _ := date.Date()
			if prevYear != 0 && (year != prevYear || month != prevMonth) {
				fmt.Fprintf(&builder, "\n")
			}
			fmt.Fprintf(&builder, "=> /~ask/dailydigest/%s %s\n", date.Format("2006-01-02"), date.Format("2006-01-02"))

			prevYear, prevMonth = year, month
		}

		request.Gemini(builder.String())
	})
	s.AddRoute("/~ask/dailydigest/:date", func(request sis.Request) {
		dateString := request.GetParam("date")
		date, err := time.Parse("2006-01-02", dateString)
		if err != nil {
			request.TemporaryFailure("Malformed date string.")
			return
		}

		var builder strings.Builder
		activities := getRecentActivityFromDate_Questions(conn, date)
		for _, activity := range activities {
			if activity.Activity == "question" {
				fmt.Fprintf(&builder, "=> /~ask/%d/%d %s %s > %s (%s)\n", activity.Q.TopicId, activity.Q.Id, activity.Activity_date.Format("2006-01-02"), activity.TopicTitle, activity.Q.Title, activity.User.Username)
			} else if activity.Activity == "answer" {
				fmt.Fprintf(&builder, "=> /~ask/%d/%d/a/%d %s %s > Re %s (%s)\n", activity.Q.TopicId, activity.Q.Id, activity.AnswerId, activity.Activity_date.Format("2006-01-02"), activity.TopicTitle, activity.Q.Title, activity.User.Username)
			} else {
				fmt.Fprintf(&builder, "'%s'\n", activity.Activity)
			}
		}

		request.Gemini(fmt.Sprintf(`# %s AuraGem Ask Activity

=> /~ask/ What Is AuraGem Ask?
=> /~ask/dailydigest Daily Digest

%s
`, date.Format("2006-01-02"), builder.String()))
	})

	s.AddRoute("/~ask/myquestions", func(request sis.Request) {
		if !request.HasUserCert() {
			request.RequestClientCert("Please enable a certificate.")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			if isRegistered {
				getUserQuestionsPage(request, conn, user)
			} else {
				request.Gemini(registerNotification)
				return
			}
		}
	})

	s.AddRoute("/~ask/:topicid", func(request sis.Request) {
		if !request.HasUserCert() {
			getTopicHomepage(request, conn, (AskUser{}), false)
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getTopicHomepage(request, conn, user, isRegistered)
		}
	})

	s.AddUploadRoute("/~ask/:topicid/create", func(request sis.Request) {
		if !request.HasUserCert() {
			getCreateQuestion(request, conn, (AskUser{}), false)
		} else if request.Upload {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			doCreateQuestion(request, conn, user, isRegistered)
		}
	})
	s.AddRoute("/~ask/:topicid/create", func(request sis.Request) {
		if !request.HasUserCert() {
			getCreateQuestion(request, conn, (AskUser{}), false)
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateQuestion(request, conn, user, isRegistered)
		}
	})

	s.AddRoute("/~ask/:topicid/create/title", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}

		if !request.HasUserCert() {
			getCreateQuestionTitle(request, conn, (AskUser{}), false, 0, query)
		} else if query == "" {
			request.RequestInput("Title of Question:")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateQuestionTitle(request, conn, user, isRegistered, 0, query)
		}
	})

	s.AddRoute("/~ask/:topicid/create/text", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}

		if !request.HasUserCert() {
			getCreateQuestionText(request, conn, (AskUser{}), false, 0, query)
		} else if query == "" {
			request.RequestInput("Text of Question:")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateQuestionText(request, conn, user, isRegistered, 0, query)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid", func(request sis.Request) {
		if !request.HasUserCert() {
			getQuestionPage(request, conn, (AskUser{}), false)
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getQuestionPage(request, conn, user, isRegistered)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/addtitle", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}

		questionId, err := strconv.Atoi(request.GetParam("questionid"))
		if err != nil {
			return
		}

		if !request.HasUserCert() {
			getCreateQuestionTitle(request, conn, (AskUser{}), false, questionId, query)
		} else if query == "" {
			request.RequestInput("Title of Question:")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateQuestionTitle(request, conn, user, isRegistered, questionId, query)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/addtext", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}

		questionId, err := strconv.Atoi(request.GetParam("questionid"))
		if err != nil {
			return
		}

		if !request.HasUserCert() {
			getCreateQuestionText(request, conn, (AskUser{}), false, questionId, query)
		} else if query == "" {
			request.RequestInput("Text of Question:")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateQuestionText(request, conn, user, isRegistered, questionId, query)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/raw", func(request sis.Request) { // TODO
		// Used for titan edits
		if !request.HasUserCert() {
			//return getQuestionPage(c, conn, (AskUser{}), false)
			request.TemporaryFailure("Certificate required.")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else {
				getQuestionPage(request, conn, user, isRegistered)
			}
			//return getQuestionPage(c, conn, user, isRegistered)
		}
	})

	s.AddUploadRoute("/~ask/:topicid/:questionid/a/create", func(request sis.Request) {
		if !request.HasUserCert() {
			getCreateAnswer(request, conn, (AskUser{}), false)
		} else if request.Upload {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			doCreateAnswer(request, conn, user, isRegistered)
		}
	})
	s.AddRoute("/~ask/:topicid/:questionid/a/create", func(request sis.Request) {
		if !request.HasUserCert() {
			getCreateAnswer(request, conn, (AskUser{}), false)
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateAnswer(request, conn, user, isRegistered)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/a/create/text", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}

		if !request.HasUserCert() {
			getCreateAnswerText(request, conn, (AskUser{}), false, 0, query)
		} else if query == "" {
			request.RequestInput("Text of Answer:")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateAnswerText(request, conn, user, isRegistered, 0, query)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/a/create/gemlog", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}

		if !request.HasUserCert() {
			getCreateAnswerGemlog(request, conn, (AskUser{}), false, query)
		} else if query == "" {
			request.RequestInput("Gemlog URL:")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateAnswerGemlog(request, conn, user, isRegistered, query)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/a/:answerid", func(request sis.Request) {
		if !request.HasUserCert() {
			getAnswerPage(request, conn, (AskUser{}), false)
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getAnswerPage(request, conn, user, isRegistered)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/a/:answerid/addtext", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}

		answerId, err := strconv.Atoi(request.GetParam("answerid"))
		if err != nil {
			return
		}

		if !request.HasUserCert() {
			getCreateAnswerText(request, conn, (AskUser{}), false, answerId, query)
		} else if query == "" {
			request.RequestInput("Text of Answer:")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			getCreateAnswerText(request, conn, user, isRegistered, answerId, query)
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/a/:answerid/upvote", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}
		query = strings.ToLower(query)

		if !request.HasUserCert() {
			request.RequestClientCert("Please enable a certificate.")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else if query == "" {
				request.RequestInput("Upvote? [yes/no]")
				return
			} else {
				getUpvoteAnswer(request, conn, user, query)
			}
		}
	})

	s.AddRoute("/~ask/:topicid/:questionid/a/:answerid/removeupvote", func(request sis.Request) {
		query, err2 := request.Query()
		if err2 != nil {
			return
		}
		query = strings.ToLower(query)

		if !request.HasUserCert() {
			request.RequestClientCert("Please enable a certificate.")
			return
		} else {
			user, isRegistered := GetUser(conn, request.UserCertHash())
			if !isRegistered {
				request.Gemini(registerNotification)
				return
			} else if query == "" {
				request.RequestInput("Remove Upvote? [yes/no]")
				return
			} else {
				getRemoveUpvoteAnswer(request, conn, user, query)
			}
		}
	})
}

// -- Homepages Handling --

func getHomepage(request sis.Request, conn *sql.DB) {
	topics := GetTopics(conn)

	// TODO: Show user's questions

	var builder strings.Builder
	for _, topic := range topics {
		fmt.Fprintf(&builder, "=> /~ask/%d %s\n%s\nQuestions Asked: %d\n\n", topic.Id, topic.Title, topic.Description, topic.QuestionTotal)
	}

	request.Gemini(fmt.Sprintf(`# AuraGem Ask

AuraGem Ask is a Gemini-first Question and Answer service, similar to Quora and StackOverflow.

AuraGem Ask supports uploading content via Gemini or Titan. Optionally, one can submit a URL to a gemlog that answers a particular question, and the content of the gemlog will be cached in case it goes down for any reason.

You must register first before being able to post. Registering will simply associate a username to your certificate. To register, create and enable a client certificate and then click the link below to register your cert:

=> /~ask/?register Register Cert
=> gemini://transjovian.org/titan About Titan

=> /~ask/dailydigest Daily Digest
=> /~ask/recent Recent Activity on All Topics

## Topics

%s
`, builder.String()))
}

func getUserDashboard(request sis.Request, conn *sql.DB, user AskUser) {
	topics := GetTopics(conn)

	// TODO: Show user's questions

	var builder strings.Builder
	for _, topic := range topics {
		fmt.Fprintf(&builder, "=> /~ask/%d %s\n%s\nQuestions Asked: %d\n\n", topic.Id, topic.Title, topic.Description, topic.QuestionTotal)
	}

	request.Gemini(fmt.Sprintf(`# AuraGem Ask - %s

=> /~ask/dailydigest Daily Digest
=> /~ask/recent Recent Activity on All Topics
=> /~ask/myquestions List of Your Questions

## Topics

%s
`, user.Username, builder.String()))
}

func getUserQuestionsPage(request sis.Request, conn *sql.DB, user AskUser) {
	var builder strings.Builder
	userQuestions := GetUserQuestions(conn, user)
	for _, question := range userQuestions {
		fmt.Fprintf(&builder, "=> /~ask/%d/%d %s %s\n", question.TopicId, question.Id, question.Date_added.Format("2006-01-02"), question.Title)
	}

	request.Gemini(fmt.Sprintf(`# AuraGem Ask - Your Questions

=> /~ask/ AuraGem Ask Root

%s
`, builder.String()))
}

// TODO: Pagination for all questions
func getTopicHomepage(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool) {
	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# AuraGem Ask > %s\n\n", topic.Title)

	if isRegistered {
		fmt.Fprintf(&builder, "Welcome %s\n\n", user.Username)
	}

	fmt.Fprintf(&builder, "=> /~ask/ AuraGem Ask Root\n")
	if isRegistered {
		fmt.Fprintf(&builder, "=> /~ask/%d/create Create New Question\n", topic.Id)
	} else {
		fmt.Fprintf(&builder, "=> /~ask/?register Register Cert\n")
	}

	// TODOShow user's questions for this topic

	fmt.Fprintf(&builder, "\n## Recent Questions\n")
	questions := getQuestionsForTopic(conn, topic.Id)
	for _, question := range questions {
		fmt.Fprintf(&builder, "=> /~ask/%d/%d %s %s (%s)\n", topic.Id, question.Id, question.Date_added.Format("2006-01-02"), question.Title, question.User.Username)
	}

	request.Gemini(builder.String())
}

// -- Question Handling --

func getCreateQuestion(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool) {
	if !isRegistered {
		request.Gemini(registerNotification)
		return
	}

	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	titanHost := "titan://auragem.letz.dev/"
	if request.Hostname() == "192.168.0.60" {
		titanHost = "titan://192.168.0.60/"
	} else if request.Hostname() == "auragem.ddns.net" {
		titanHost = "titan://auragem.ddns.net/"
	}

	// /create/title?TitleHere  -> Creates the question with title and blank text -> Redirects to question's new page with questionid
	// /create/text?TextHere -> Creates the question with text and blank title -> Redirects to question's new page with questionId
	request.Gemini(fmt.Sprintf(`# AuraGem Ask > %s - Create New Question

=> /~ask/%d %s Homepage

To create a new question, you can do one of the following:

1. Upload text to this url with Titan. Use a level-1 heading ('#') for the Question Title. All other heading lines are disallowed and will be stripped. This option is suitable for long posts. Make sure your certificate/identity is selected when uploading with Titan.
=> %s/~ask/%d/create Upload with Titan

2. Or add a Title and Text via Gemini with the links below. Note that Gemini limits these to 1024 bytes total.
=> /~ask/%d/create/title Add Title
=> /~ask/%d/create/text Add Text
`, topic.Title, topic.Id, topic.Title, titanHost, topic.Id, topic.Id, topic.Id))
}

func doCreateQuestion(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool) {
	if !isRegistered {
		request.TemporaryFailure("You must be registered.")
		return
	}

	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	// Check mimetype and size
	if request.DataMime != "text/plain" && request.DataMime != "text/gemini" {
		request.TemporaryFailure("Wrong mimetype. Only text/plain and text/gemini supported.")
		return
	}
	if request.DataSize > 16*1024 {
		request.TemporaryFailure("Size too large. Max size allowed is 16 KiB.")
		return
	}

	data, read_err := request.GetUploadData()
	if read_err != nil {
		return
	}

	text := string(data)
	if !utf8.ValidString(text) {
		request.TemporaryFailure("Not a valid UTF-8 text file.")
		return
	}
	if ContainsCensorWords(text) {
		request.TemporaryFailure("Profanity or slurs were detected. Your edit is rejected.")
		return
	}

	strippedText, title := StripGeminiText(text)
	question, q_err := createQuestionTitan(conn, topic.Id, title, strippedText, user)
	if q_err != nil {
		return
	}

	request.Redirect(fmt.Sprintf("%s%s/~ask/%d/%d", request.Server.Scheme(), request.Hostname(), topic.Id, question.Id))
}

func getQuestionPage(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool) {
	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	questionId, err := strconv.Atoi(request.GetParam("questionid"))
	if err != nil {
		return
	}
	question, questionSuccess := getQuestion(conn, topic.Id, questionId)
	if !questionSuccess {
		request.TemporaryFailure("Question Id doesn't exist.")
		return
	}

	var builder strings.Builder
	if question.Title != "" {
		fmt.Fprintf(&builder, "# AuraGem Ask > %s: %s\n\n", topic.Title, question.Title)
	} else {
		fmt.Fprintf(&builder, "# AuraGem Ask > %s: [No Title]\n\n", topic.Title)
	}
	fmt.Fprintf(&builder, "=> /~ask/%d %s Homepage\n", topic.Id, topic.Title)
	if question.Title == "" && question.MemberId == user.Id { // Only display if user owns the question
		fmt.Fprintf(&builder, "=> /~ask/%d/%d/addtitle Add Title\n", topic.Id, questionId)
	} else if question.MemberId == user.Id {
		fmt.Fprintf(&builder, "=> /~ask/%d/%d/addtitle Edit Title\n", topic.Id, questionId)
	}

	if question.Text != "" {
		if question.MemberId == user.Id {
			fmt.Fprintf(&builder, "=> /~ask/%d/%d/addtext Edit Text\n\n", topic.Id, questionId)
		} else {
			fmt.Fprintf(&builder, "\n")
		}
		fmt.Fprintf(&builder, "%s\n\n", question.Text)
	} else if question.MemberId == user.Id { // Only display if user owns the question
		fmt.Fprintf(&builder, "\n")
		fmt.Fprintf(&builder, "=> /~ask/%d/%d/addtext Add Text\n\n", topic.Id, questionId)
	} else {
		fmt.Fprintf(&builder, "\n")
		fmt.Fprintf(&builder, "[No Body Text]\n\n")
	}

	fmt.Fprintf(&builder, "Asked %s UTC by %s\n\n", question.Date_added.Format("2006-01-02 15:04"), question.User.Username)
	fmt.Fprintf(&builder, "## Answers List\n\n")

	if isRegistered {
		fmt.Fprintf(&builder, "=> /~ask/%d/%d/a/create Create New Answer\n\n", topic.Id, question.Id)
	} else {
		fmt.Fprintf(&builder, "=> /~ask/?register Register Cert\n")
	}

	answers := getAnswersForQuestion(conn, question)
	for _, answer := range answers {
		// TODO: Check if gemlog answer
		if answer.Gemlog_url != nil {
			fmt.Fprintf(&builder, "=> %s %s Gemlog Answer by %s\n", GetNormalizedURL(answer.Gemlog_url), answer.Date_added.Format("2006-01-02"), answer.User.Username)
		} else {
			fmt.Fprintf(&builder, "=> /~ask/%d/%d/a/%d %s %s\n", topic.Id, question.Id, answer.Id, answer.Date_added.Format("2006-01-02"), answer.User.Username)
		}
	}

	request.Gemini(builder.String())
}

func getCreateQuestionTitle(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool, questionId int, title string) {
	if !isRegistered {
		request.Gemini(registerNotification)
		return
	}

	if ContainsCensorWords(title) {
		request.TemporaryFailure("Profanity or slurs were detected. They are not allowed.")
		return
	}

	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	// Make sure no newlines in title // TODO
	//title = strings.Fields(title)[0]
	title = strings.FieldsFunc(title, func(r rune) bool {
		return r == '\n'
	})[0]

	if questionId == 0 {
		// Question hasn't been created yet. This is from the initial /create page. Create a new question.
		question, q_err := createQuestionWithTitle(conn, topic.Id, title, user)
		if q_err != nil {
			return
		}
		request.Redirect(fmt.Sprintf("/~ask/%d/%d", topic.Id, question.Id))
		return
	} else {
		question, questionSuccess := getQuestion(conn, topic.Id, questionId)
		if !questionSuccess {
			request.TemporaryFailure("Question Id doesn't exist.")
			return
		}

		// Check that the current user owns this question
		if question.MemberId != user.Id {
			request.TemporaryFailure("You cannot edit this question, since you did not post it.")
			return
		}

		var q_err error
		question, q_err = updateQuestionTitle(conn, question, title, user)
		if q_err != nil {
			return
		}
		request.Redirect(fmt.Sprintf("/~ask/%d/%d", topic.Id, question.Id))
		return
	}
}

func getCreateQuestionText(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool, questionId int, text string) {
	if !isRegistered {
		request.Gemini(registerNotification)
		return
	}

	if ContainsCensorWords(text) {
		request.TemporaryFailure("Profanity or slurs were detected. They are not allowed.")
		return
	}

	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	strippedText, _ := StripGeminiText(text)
	if questionId == 0 {
		// Question hasn't been created yet. This is from the initial /create page. Create a new question.
		question, q_err := createQuestionWithText(conn, topic.Id, strippedText, user)
		if q_err != nil {
			return
		}
		request.Redirect(fmt.Sprintf("/~ask/%d/%d", topic.Id, question.Id))
		return
	} else {
		question, questionSuccess := getQuestion(conn, topic.Id, questionId)
		if !questionSuccess {
			request.TemporaryFailure("Question Id doesn't exist.")
			return
		}

		// Check that the current user owns this question
		if question.MemberId != user.Id {
			request.TemporaryFailure("You cannot edit this question, since you did not post it.")
			return
		}

		var q_err error
		question, q_err = updateQuestionText(conn, question, strippedText, user)
		if q_err != nil {
			return
		}

		request.Redirect(fmt.Sprintf("/~ask/%d/%d", topic.Id, question.Id))
		return
	}
}

// -- Answer Handling --

func getCreateAnswer(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool) {
	if !isRegistered {
		request.Gemini(registerNotification)
		return
	}

	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	questionid, err := strconv.Atoi(request.GetParam("questionid"))
	if err != nil {
		return
	}
	question, questionSuccess := getQuestion(conn, topic.Id, questionid)
	if !questionSuccess {
		request.TemporaryFailure("Question Id doesn't exist.")
		return
	}

	titanHost := "titan://auragem.letz.dev/"
	if request.Hostname() == "192.168.0.60" {
		titanHost = "titan://192.168.0.60/"
	} else if request.Hostname() == "auragem.ddns.net" {
		titanHost = "titan://auragem.ddns.net/"
	}

	// /create/text?TextHere -> Creates the question with text and blank title -> Redirects to question's new page with questionId
	request.Gemini(fmt.Sprintf(`# AuraGem Ask > %s - Create New Answer

=> /~ask/%d %s Homepage
=> /~ask/%d/%d Back to Question

To create a new answer, you can do one of the following:

1. Upload text to this url with Titan. All heading lines are disallowed and will be stripped. This option is suitable for long posts. Make sure your certificate/identity is selected when uploading with Titan.
=> %s/~ask/%d/%d/a/create Upload with Titan

2. Add Text via Gemini with the link below. Note that Gemini limits these to 1024 bytes total.
=> /~ask/%d/%d/a/create/text Add Text

3. Or Submit a URL to a gemlog post in response to the question
=> /~ask/%d/%d/a/create/gemlog Post URL of Gemlog Response
`, topic.Title, topic.Id, topic.Title, topic.Id, question.Id, titanHost, topic.Id, question.Id, topic.Id, question.Id, topic.Id, question.Id))
}

func doCreateAnswer(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool) {
	if !isRegistered {
		request.TemporaryFailure("You must be registered.")
		return
	}

	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	questionid, err2 := strconv.Atoi(request.GetParam("questionid"))
	if err2 != nil {
		return
	}
	question, questionSuccess := getQuestion(conn, topic.Id, questionid)
	if !questionSuccess {
		request.TemporaryFailure("Question Id doesn't exist.")
		return
	}

	// Check mimetype and size
	if request.DataMime != "text/plain" && request.DataMime != "text/gemini" {
		request.TemporaryFailure("Wrong mimetype. Only text/plain and text/gemini supported.")
		return
	}
	if request.DataSize > 16*1024 {
		request.TemporaryFailure("Size too large. Max size allowed is 16 KiB.")
		return
	}

	data, read_err := request.GetUploadData()
	if read_err != nil {
		return
	}

	text := string(data)
	if !utf8.ValidString(text) {
		request.TemporaryFailure("Not a valid UTF-8 text file.")
		return
	}
	if ContainsCensorWords(text) {
		request.TemporaryFailure("Profanity or slurs were detected. Your edit is rejected.")
		return
	}

	strippedText, _ := StripGeminiText(text)
	answer, a_err := createAnswerWithText(conn, question.Id, strippedText, user)
	if a_err != nil {
		return
	}
	request.Redirect(fmt.Sprintf("%s%s/~ask/%d/%d/a/%d", request.Server.Scheme(), request.Hostname(), topic.Id, question.Id, answer.Id))
}

func getCreateAnswerText(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool, answerId int, text string) {
	if !isRegistered {
		request.Gemini(registerNotification)
		return
	}

	if ContainsCensorWords(text) {
		request.TemporaryFailure("Profanity or slurs were detected. They are not allowed.")
		return
	}

	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	questionid, err2 := strconv.Atoi(request.GetParam("questionid"))
	if err2 != nil {
		return
	}
	question, questionSuccess := getQuestion(conn, topic.Id, questionid)
	if !questionSuccess {
		request.TemporaryFailure("Question Id doesn't exist.")
		return
	}

	strippedText, _ := StripGeminiText(text)
	if answerId == 0 {
		// Answer hasn't been created yet. This is from the initial /create page. Create a new question.
		answer, a_err := createAnswerWithText(conn, question.Id, strippedText, user)
		if a_err != nil {
			return
		}
		request.Redirect(fmt.Sprintf("/~ask/%d/%d/a/%d", topic.Id, question.Id, answer.Id))
		return
	} else {
		answer, answerSuccess := getAnswer(conn, question.Id, answerId)
		if !answerSuccess {
			request.TemporaryFailure("Answer Id doesn't exist.")
			return
		}

		// Check that the current user owns this question
		if answer.MemberId != user.Id {
			request.TemporaryFailure("You cannot edit this answer, since you did not post it.")
			return
		}

		var a_err error
		answer, a_err = updateAnswerText(conn, answer, strippedText, user)
		if a_err != nil {
			return
		}

		request.Redirect(fmt.Sprintf("/~ask/%d/%d/a/%d", topic.Id, question.Id, answer.Id))
		return
	}
}

func getCreateAnswerGemlog(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool, gemlogUrl string) {
	if !isRegistered {
		request.Gemini(registerNotification)
		return
	}

	/*if ContainsCensorWords(text) {
			request.TemporaryFailure("Profanity or slurs were detected. They are not allowed.")
	return
		}*/

	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	questionid, err2 := strconv.Atoi(request.GetParam("questionid"))
	if err2 != nil {
		return
	}
	question, questionSuccess := getQuestion(conn, topic.Id, questionid)
	if !questionSuccess {
		request.TemporaryFailure("Question Id doesn't exist.")
		return
	}

	gemlogUrlNormalized, url_err := checkValidUrl(gemlogUrl)
	if url_err != nil {
		return
	}

	// Get text of gemlog so we can cache it in the DB
	client := gemini2.DefaultClient
	resp, fetch_err := client.Fetch(gemlogUrlNormalized)
	if fetch_err != nil {
		request.TemporaryFailure("Failed to fetch gemlog at given url.")
		return
	} else if resp.Status == 30 || resp.Status == 31 {
		request.TemporaryFailure("Failed to fetch gemlog at given url. Links that redirect are not allowed.")
		return
	} else if resp.Status != 20 {
		request.TemporaryFailure("Failed to fetch gemlog at given url.")
		return
	}
	mediatype, _, _ := mime.ParseMediaType(resp.Meta)
	if mediatype != "text/plain" && mediatype != "text/gemini" && mediatype != "text/nex" && mediatype != "application/gopher-menu" && mediatype != "text/scroll" && mediatype != "text/markdown" {
		request.TemporaryFailure("Gemlog mimetype not supported. Ask only supports gemtext, nex, gophermenus, scrolltext, markdown, or plain text.")
		return
	}
	textData, read_err := io.ReadAll(resp.Body)
	if read_err != nil {
		request.TemporaryFailure("Failed to fetch gemlog at given url.")
		return
	}

	_, a_err := createAnswerAsGemlog(conn, question.Id, gemlogUrlNormalized, user, string(textData))
	if a_err != nil {
		return
	}
	request.Redirect(fmt.Sprintf("/~ask/%d/%d", topic.Id, question.Id))
}

func getAnswerPage(request sis.Request, conn *sql.DB, user AskUser, isRegistered bool) {
	// Get Topic
	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	// Get Question
	questionId, err := strconv.Atoi(request.GetParam("questionid"))
	if err != nil {
		return
	}
	question, questionSuccess := getQuestion(conn, topic.Id, questionId)
	if !questionSuccess {
		request.TemporaryFailure("Question Id doesn't exist.")
		return
	}

	// Get Answer
	answerId, a_err := strconv.Atoi(request.GetParam("answerid"))
	if a_err != nil {
		return
	}
	answer, answerSuccess := getAnswer(conn, question.Id, answerId)
	if !answerSuccess {
		request.TemporaryFailure("Answer Id doesn't exist.")
		return
	}

	// Get Upvotes
	upvotes := getUpvotesWithUsers(conn, answer)
	upvotesCount := len(upvotes)
	currentUserHasUpvoted := false
	for _, upvote := range upvotes {
		if upvote.MemberId == user.Id {
			currentUserHasUpvoted = true
			break
		}
	}

	var builder strings.Builder
	if question.Title != "" {
		fmt.Fprintf(&builder, "# AuraGem Ask > %s: Re %s\n\n", topic.Title, question.Title)
	} else {
		fmt.Fprintf(&builder, "# AuraGem Ask > %s: Re [No Title]\n\n", topic.Title)
	}

	fmt.Fprintf(&builder, "=> /~ask/%d/%d Back to the Question\n", topic.Id, question.Id)

	if answer.Text != "" {
		if answer.MemberId == user.Id && len(answer.Text) < 1024 { // Don't show Gemini Edit if text is under 1024 bytes
			fmt.Fprintf(&builder, "=> /~ask/%d/%d/a/%d/addtext Edit Text\n\n", topic.Id, question.Id, answer.Id)
		} else {
			fmt.Fprintf(&builder, "\n")
		}
		fmt.Fprintf(&builder, "%s\n\n", answer.Text)
	} else if answer.MemberId == user.Id && answer.Gemlog_url == nil { // Only display if user owns the question
		fmt.Fprintf(&builder, "\n")
		fmt.Fprintf(&builder, "=> /~ask/%d/%d/a/%d/addtext Add Text\n\n", topic.Id, question.Id, answer.Id)
	} else if answer.Gemlog_url != nil {
		fmt.Fprintf(&builder, "\n")
		fmt.Fprintf(&builder, "=> %s Link to Gemlog\n", GetNormalizedURL(answer.Gemlog_url))
	} else {
		fmt.Fprintf(&builder, "\n")
		fmt.Fprintf(&builder, "[No Body Text]\n\n")
	}

	fmt.Fprintf(&builder, "Answered %s UTC by %s\n", answer.Date_added.Format("2006-01-02 15:04"), answer.User.Username)
	fmt.Fprintf(&builder, "Upvotes: %d\n\n", upvotesCount)

	if isRegistered {
		if currentUserHasUpvoted {
			fmt.Fprintf(&builder, "=> /~ask/%d/%d/a/%d/removeupvote Remove Upvote\n", topic.Id, question.Id, answer.Id)
		} else {
			fmt.Fprintf(&builder, "=> /~ask/%d/%d/a/%d/upvote Upvote\n", topic.Id, question.Id, answer.Id)
		}
	} else {
		fmt.Fprintf(&builder, "=> /~ask/?register Register Cert\n")
	}
	/*fmt.Fprintf(&builder, "## Comments\n\n")*/

	request.Gemini(builder.String())
}

func getUpvoteAnswer(request sis.Request, conn *sql.DB, user AskUser, query string) {
	// Get Topic
	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	// Get Question
	questionId, err := strconv.Atoi(request.GetParam("questionid"))
	if err != nil {
		return
	}
	question, questionSuccess := getQuestion(conn, topic.Id, questionId)
	if !questionSuccess {
		request.TemporaryFailure("Question Id doesn't exist.")
		return
	}

	// Get Answer
	answerId, a_err := strconv.Atoi(request.GetParam("answerid"))
	if a_err != nil {
		return
	}
	answer, answerSuccess := getAnswer(conn, question.Id, answerId)
	if !answerSuccess {
		request.TemporaryFailure("Answer Id doesn't exist.")
		return
	}

	if query == "yes" || query == "y" {
		_, db_err := addUpvote(conn, answer, user)
		if db_err != nil {
			return
		}
	}

	request.Redirect("/~ask/%d/%d/a/%d", topic.Id, question.Id, answer.Id)
}

func getRemoveUpvoteAnswer(request sis.Request, conn *sql.DB, user AskUser, query string) {
	// Get Topic
	topicId, err := strconv.Atoi(request.GetParam("topicid"))
	if err != nil {
		return
	}
	topic, topicSuccess := getTopic(conn, topicId)
	if !topicSuccess {
		request.TemporaryFailure("Topic Id doesn't exist.")
		return
	}

	// Get Question
	questionId, err := strconv.Atoi(request.GetParam("questionid"))
	if err != nil {
		return
	}
	question, questionSuccess := getQuestion(conn, topic.Id, questionId)
	if !questionSuccess {
		request.TemporaryFailure("Question Id doesn't exist.")
		return
	}

	// Get Answer
	answerId, a_err := strconv.Atoi(request.GetParam("answerid"))
	if a_err != nil {
		return
	}
	answer, answerSuccess := getAnswer(conn, question.Id, answerId)
	if !answerSuccess {
		request.TemporaryFailure("Answer Id doesn't exist.")
		return
	}

	if query == "yes" || query == "y" {
		db_err := removeUpvote(conn, answer, user)
		if db_err != nil {
			return
		}
	}

	request.Redirect("/~ask/%d/%d/a/%d", topic.Id, question.Id, answer.Id)
}

func CensorWords(str string) string {
	wordCensors := []string{"fuck", "kill", "die", "damn", "ass", "shit", "stupid", "faggot", "fag", "whore", "cock", "cunt", "motherfucker", "fucker", "asshole", "nigger", "abbie", "abe", "abie", "abid", "abeed", "ape", "armo", "nazi", "ashke-nazi", "אשכנאצי", "bamboula", "barbarian", "beaney", "beaner", "bohunk", "boerehater", "boer-hater", "burrhead", "burr-head", "chode", "chad", "penis", "vagina", "porn", "bbc", "stealthing", "bbw", "Hentai", "milf", "dilf", "tummysticks", "heeb", "hymie", "kike", "jidan", "sheeny", "shylock", "zhyd", "yid", "shyster", "smouch"}

	var result string = str
	for _, forbiddenWord := range wordCensors {
		replacement := strings.Repeat("*", len(forbiddenWord))
		result = strings.Replace(result, forbiddenWord, replacement, -1)
	}

	return result
}

func ContainsCensorWords(str string) bool {
	wordCensors := map[string]bool{"fuck": true, "f*ck": true, "kill": true, "k*ll": true, "die": true, "damn": true, "ass": true, "*ss": true, "shit": true, "sh*t": true, "stupid": true, "faggot": true, "fag": true, "f*g": true, "whore": true, "wh*re": true, "cock": true, "c*ck": true, "cunt": true, "c*nt": true, "motherfucker": true, "fucker": true, "f*cker": true, "asshole": true, "*sshole": true, "nigger": true, "n*gger": true, "n*gg*r": true, "abbie": true, "abe": true, "abie": true, "abid": true, "abeed": true, "ape": true, "armo": true, "nazi": true, "ashke-nazi": true, "אשכנאצי": true, "bamboula": true, "barbarian": true, "beaney": true, "beaner": true, "bohunk": true, "boerehater": true, "boer-hater": true, "burrhead": true, "burr-head": true, "chode": true, "chad": true, "penis": true, "vagina": true, "porn": true, "stealthing": true, "bbw": true, "Hentai": true, "milf": true, "dilf": true, "tummysticks": true, "heeb": true, "hymie": true, "kike": true, "k*ke": true, "jidan": true, "sheeny": true, "shylock": true, "zhyd": true, "yid": true, "shyster": true, "smouch": true}

	fields := strings.FieldsFunc(strings.ToLower(str), func(r rune) bool {
		if r == '*' {
			return false
		}
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsDigit(r) || !unicode.IsPrint(r)
	})

	for _, word := range fields {
		if _, ok := wordCensors[word]; ok {
			return true
		}
	}

	return false
}

var InvalidURLString = errors.New("URL is not a valid UTF-8 string.")
var URLTooLong = errors.New("URL exceeds 1024 bytes.")
var InvalidURL = errors.New("URL is not valid.")
var URLRelative = errors.New("URL is relative. Only absolute URLs can be added.")
var URLNotGemini = errors.New("Must be a Gemini URL.")

func checkValidUrl(s string) (string, error) {
	// Make sure URL is a valid UTF-8 string
	if !utf8.ValidString(s) {
		return "", InvalidURLString
	}
	// Make sure URL doesn't exceed 1024 bytes
	if len(s) > 1024 {
		return "", URLTooLong
	}
	// Make sure URL has gemini:// scheme
	if !strings.HasPrefix(s, "gemini://") && !strings.Contains(s, "://") && !strings.HasPrefix(s, ".") && !strings.HasPrefix(s, "/") {
		s = "gemini://" + s
	}

	// Make sure the url is parseable and that only the hostname is being added
	u, urlErr := url.Parse(s)
	if urlErr != nil { // Check if able to parse
		return "", InvalidURL
	}
	if !u.IsAbs() { // Check if Absolute URL
		return "", URLRelative
	}
	if u.Scheme != "gemini" { // Make sure scheme is gemini
		return "", URLNotGemini
	}

	return GetNormalizedURL(u), nil
}

func GetNormalizedURL(u *url.URL) string {
	var buf strings.Builder

	// Hostname
	if u.Port() == "" || u.Port() == "1965" {
		buf.WriteString(u.Scheme)
		buf.WriteString("://")
		buf.WriteString(u.Hostname())
		//buf.WriteString("/")
	} else {
		buf.WriteString(u.Scheme)
		buf.WriteString("://")
		buf.WriteString(u.Hostname())
		buf.WriteByte(':')
		buf.WriteString(u.Port())
		//buf.WriteString("/")
	}

	// Path
	path := u.EscapedPath()
	if path == "" || (path != "" && path[0] != '/' && u.Host != "") {
		buf.WriteByte('/')
	}
	buf.WriteString(path)

	// Queries and Fragments
	if u.ForceQuery || u.RawQuery != "" {
		buf.WriteByte('?')
		buf.WriteString(u.RawQuery)
	}
	if u.Fragment != "" {
		buf.WriteByte('#')
		buf.WriteString(u.EscapedFragment())
	}

	return buf.String()
}

// Go through gemini document to get keywords, title, and links
func StripGeminiText(s string) (string, string) {
	var strippedTextBuilder strings.Builder
	text, _ := gemini.ParseText(strings.NewReader(s))
	title := ""

	for _, line := range text {
		switch v := line.(type) {
		case gemini.LineHeading1:
			{
				if title == "" {
					title = string(v)
				}
			}
		case gemini.LineHeading2:
			{
			}
		case gemini.LineHeading3:
			{
			}
		case gemini.LineLink:
			{
				fmt.Fprintf(&strippedTextBuilder, "%s\n", v.String())
			}
		case gemini.LineListItem:
			{
				fmt.Fprintf(&strippedTextBuilder, "%s\n", v.String())
			}
		case gemini.LinePreformattingToggle:
			{
				fmt.Fprintf(&strippedTextBuilder, "%s\n", v.String())
			}
		case gemini.LinePreformattedText:
			{
				fmt.Fprintf(&strippedTextBuilder, "%s\n", v.String())
			}
		case gemini.LineQuote:
			{
				fmt.Fprintf(&strippedTextBuilder, "%s\n", v.String())
			}
		case gemini.LineText:
			{
				fmt.Fprintf(&strippedTextBuilder, "%s\n", v.String())
			}
		}
	}

	// TODO: Strip blank lines at beginning and end of string
	// Use strings.TrimSpace?

	return strings.TrimSpace(strippedTextBuilder.String()), title
}
