package ask

import (
	"time"
	"net/url"
	"database/sql"
	"strings"
)

type AskUser struct {
	Id int
	Username string
	Certificate AskUserCert
	Language string
	Timezone string
	Is_staff bool
	Is_active bool
	Date_joined time.Time
}

type AskUserCert struct {
	Id int
	MemberId int
	Title string
	Certificate string
	Is_active bool
	Date_added time.Time
}

func ScanAskUserCert(row *sql.Row) AskUserCert {
	cert := AskUserCert{}
	var title interface{}
	row.Scan(&cert.Id, &cert.MemberId, &title, &cert.Certificate, &cert.Is_active, &cert.Date_added)
	if title != nil {
		cert.Title = title.(string)
	}

	return cert
}

type Topic struct {
	Id int
	Title string
	Description string
	Date_added time.Time

	QuestionTotal int
}

type Question struct {
	Id int
	TopicId int
	Title string // Nullable
	Text string // Nullable
	Tags string // Nullable
	MemberId int // Nullable
	Date_added time.Time

	User AskUser
	SelectedAnswer int
}

func scanQuestion(row *sql.Row) (Question, error) {
	question := Question{}
	var title interface{}
	var text interface{}
	var tags interface{}
	var memberid interface{}
	err := row.Scan(&question.Id, &question.TopicId, &title, &text, &tags, &memberid, &question.Date_added)
	if err != nil {
		return Question{}, err
	}
	if title != nil {
		if s, ok := title.([]uint8); ok {
			question.Title = string(s)
		} else if s, ok := title.(string); ok {
			question.Title = s
		}
	}
	if text != nil {
		if s, ok := text.([]uint8); ok {
			question.Text = string(s)
		} else if s, ok := text.(string); ok {
			question.Text = s
		}
	}
	if tags != nil {
		if s, ok := tags.([]uint8); ok {
			question.Tags = string(s)
		} else if s, ok := tags.(string); ok {
			question.Tags = s
		}
	}
	if memberid != nil {
		question.MemberId = int(memberid.(int64))
	}

	return question, nil
}

func scanQuestionWithUser(row *sql.Row) (Question, error) {
	question := Question{}
	var title interface{}
	var text interface{}
	var tags interface{}
	var memberid interface{}
	err := row.Scan(&question.Id, &question.TopicId, &title, &text, &tags, &memberid, &question.Date_added, &question.User.Id, &question.User.Username, &question.User.Language, &question.User.Timezone, &question.User.Is_staff, &question.User.Is_active, &question.User.Date_joined)
	if err != nil {
		return Question{}, err
	}
	if title != nil {
		if s, ok := title.([]uint8); ok {
			question.Title = string(s)
		} else if s, ok := title.(string); ok {
			question.Title = s
		}
	}
	if text != nil {
		if s, ok := text.([]uint8); ok {
			question.Text = string(s)
		} else if s, ok := text.(string); ok {
			question.Text = s
		}
	}
	if tags != nil {
		if s, ok := tags.([]uint8); ok {
			question.Tags = string(s)
		} else if s, ok := tags.(string); ok {
			question.Tags = s
		}
	}
	if memberid != nil {
		question.MemberId = int(memberid.(int64))
	}

	return question, nil
}


func scanQuestionRows(rows *sql.Rows) (Question, error) {
	question := Question{}
	var title interface{}
	var text interface{}
	var tags interface{}
	var memberid interface{}
	err := rows.Scan(&question.Id, &question.TopicId, &title, &text, &tags, &memberid, &question.Date_added)
	if err != nil {
		return Question{}, err
	}
	if title != nil {
		if s, ok := title.([]uint8); ok {
			question.Title = string(s)
		} else if s, ok := title.(string); ok {
			question.Title = s
		}
	}
	if text != nil {
		if s, ok := text.([]uint8); ok {
			question.Text = string(s)
		} else if s, ok := text.(string); ok {
			question.Text = s
		}
	}
	if tags != nil {
		if s, ok := tags.([]uint8); ok {
			question.Tags = string(s)
		} else if s, ok := tags.(string); ok {
			question.Tags = s
		}
	}
	if memberid != nil {
		question.MemberId = int(memberid.(int64))
	}

	return question, nil
}
func scanQuestionRowsWithUser(rows *sql.Rows) (Question, error) {
	question := Question{}
	var title interface{}
	var text interface{}
	var tags interface{}
	var memberid interface{}
	err := rows.Scan(&question.Id, &question.TopicId, &title, &text, &tags, &memberid, &question.Date_added, &question.User.Id, &question.User.Username, &question.User.Language, &question.User.Timezone, &question.User.Is_staff, &question.User.Is_active, &question.User.Date_joined)
	if err != nil {
		return Question{}, err
	}
	if title != nil {
		if s, ok := title.([]uint8); ok {
			question.Title = string(s)
		} else if s, ok := title.(string); ok {
			question.Title = s
		}
	}
	if text != nil {
		if s, ok := text.([]uint8); ok {
			question.Text = string(s)
		} else if s, ok := text.(string); ok {
			question.Text = s
		}
	}
	if tags != nil {
		if s, ok := tags.([]uint8); ok {
			question.Tags = string(s)
		} else if s, ok := tags.(string); ok {
			question.Tags = s
		}
	}
	if memberid != nil {
		question.MemberId = int(memberid.(int64))
	}

	return question, nil
}

type Answer struct {
	Id int
	QuestionId int // Nullable
	Text string // Nullable
	Gemlog_url *url.URL // Nullable
	MemberId int // Nullable
	Date_added time.Time

	User AskUser
	Upvotes int
}

func scanAnswer(row *sql.Row) (Answer, error) {
	answer := Answer{}
	var text interface{}
	var gemlog_url interface{}
	var memberid interface{}
	err := row.Scan(&answer.Id, &answer.QuestionId, &text, &gemlog_url, &memberid, &answer.Date_added)
	if err != nil {
		return (Answer{}), err
	}
	if text != nil {
		if s, ok := text.([]uint8); ok {
			answer.Text = string(s)
		} else if s, ok := text.(string); ok {
			answer.Text = s
		}
	}
	if gemlog_url != nil {
		var gemlog_url_string = string(gemlog_url.([]uint8))
		var err2 error
		answer.Gemlog_url, err2 = url.Parse(gemlog_url_string)
		if err2 != nil {
			// TODO
		}
	}
	if memberid != nil {
		answer.MemberId = int(memberid.(int64))
	}

	return answer, nil
}

func scanAnswerWithUser(row *sql.Row) (Answer, error) {
	answer := Answer{}
	var text interface{}
	var gemlog_url interface{}
	var memberid interface{}
	err := row.Scan(&answer.Id, &answer.QuestionId, &text, &gemlog_url, &memberid, &answer.Date_added, &answer.User.Id, &answer.User.Username, &answer.User.Language, &answer.User.Timezone, &answer.User.Is_staff, &answer.User.Is_active, &answer.User.Date_joined)
	if err != nil {
		return (Answer{}), err
	}
	if text != nil {
		if s, ok := text.([]uint8); ok {
			answer.Text = string(s)
		} else if s, ok := text.(string); ok {
			answer.Text = s
		}
	}
	if gemlog_url != nil {
		var gemlog_url_string = string(gemlog_url.([]uint8))
		var err2 error
		answer.Gemlog_url, err2 = url.Parse(gemlog_url_string)
		if err2 != nil {
			// TODO
		}
	}
	if memberid != nil {
		answer.MemberId = int(memberid.(int64))
	}

	return answer, nil
}

// With User
func scanAnswerRows(rows *sql.Rows) (Answer, error) {
	answer := Answer{}
	var text interface{}
	var gemlog_url interface{}
	var memberid interface{}
	err := rows.Scan(&answer.Id, &answer.QuestionId, &text, &gemlog_url, &memberid, &answer.Date_added, &answer.User.Id, &answer.User.Username, &answer.User.Language, &answer.User.Timezone, &answer.User.Is_staff, &answer.User.Is_active, &answer.User.Date_joined)
	if err != nil {
		return (Answer{}), err
	}
	if text != nil {
		if s, ok := text.([]uint8); ok {
			answer.Text = string(s)
		} else if s, ok := text.(string); ok {
			answer.Text = s
		}
	}
	if gemlog_url != nil {
		var gemlog_url_string = string(gemlog_url.([]uint8))
		var err2 error
		answer.Gemlog_url, err2 = url.Parse(gemlog_url_string)
		if err2 != nil {
			// TODO
		}
	}
	if memberid != nil {
		answer.MemberId = int(memberid.(int64))
	}

	return answer, nil
}

type Activity struct {
	Q Question
	Activity string
	Activity_date time.Time
	AnswerId int
	User AskUser
	TopicTitle string
}


func scanActivityWithUser(rows *sql.Rows) (Activity, error) {
	activity := Activity{}
	var title interface{}
	var text interface{}
	var tags interface{}
	var memberid interface{}
	err := rows.Scan(&activity.Q.Id, &activity.Q.TopicId, &title, &text, &tags, &memberid, &activity.Activity, &activity.Activity_date, &activity.AnswerId, &activity.User.Id, &activity.User.Username, &activity.User.Language, &activity.User.Timezone, &activity.User.Is_staff, &activity.User.Is_active, &activity.User.Date_joined, &activity.TopicTitle)
	if err != nil {
		return Activity{}, err
	}
	if title != nil {
		activity.Q.Title = string(title.([]uint8))
	}
	if text != nil {
		activity.Q.Text = text.(string)
	}
	if tags != nil {
		activity.Q.Tags = string(tags.([]uint8))
	}
	if memberid != nil {
		activity.Q.MemberId = int(memberid.(int64))
	}

	activity.Activity = strings.TrimSpace(activity.Activity)

	return activity, nil
}

type QuestionComment struct {
	Id int
	QuestionId int
	Text string
	MemberId int
	Date_added time.Time
}

type AnswerComment struct {
	Id int
	AnswerId int
	Text string
	MemberId int
	Date_added time.Time
}

type Upvote struct {
	Id int
	AnswerId int
	MemberId int
	Date_added time.Time

	User AskUser
}

func scanUpvote(row *sql.Row) (Upvote, error) {
	upvote := Upvote{}
	var memberid interface{}
	err := row.Scan(&upvote.Id, &upvote.AnswerId, &memberid, &upvote.Date_added)
	if err != nil {
		return (Upvote{}), err
	}
	if memberid != nil {
		upvote.MemberId = int(memberid.(int64))
	}

	return upvote, nil
}

func scanUpvoteRowsWithUser(rows *sql.Rows) (Upvote, error) {
	upvote := Upvote{}
	var memberid interface{}
	err := rows.Scan(&upvote.Id, &upvote.AnswerId, &memberid, &upvote.Date_added, &upvote.User.Id, &upvote.User.Username, &upvote.User.Language, &upvote.User.Timezone, &upvote.User.Is_staff, &upvote.User.Is_active, &upvote.User.Date_joined)
	if err != nil {
		return (Upvote{}), err
	}
	if memberid != nil {
		upvote.MemberId = int(memberid.(int64))
	}

	return upvote, nil
}
