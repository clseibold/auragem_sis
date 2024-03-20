package ask

import (
	"context"
	"database/sql"
	"strings"
	"time"

	sis "gitlab.com/clseibold/smallnetinformationservices"
)

func GetUser(conn *sql.DB, certHash string) (AskUser, bool) {
	//query := "SELECT id, username, language, timezone, is_staff, is_active, date_joined FROM members LEFT JOIN membercerts ON membercerts.memberid = members.id WHERE membercerts.certificate=?"
	query := "SELECT membercerts.id, membercerts.memberid, membercerts.title, membercerts.certificate, membercerts.is_active, membercerts.date_added, members.id, members.username, members.language, members.timezone, members.is_staff, members.is_active, members.date_joined FROM membercerts LEFT JOIN members ON membercerts.memberid = members.id WHERE membercerts.certificate=? AND membercerts.is_active = true"
	row := conn.QueryRowContext(context.Background(), query, certHash)

	var user AskUser
	var certTitle interface{}
	err := row.Scan(&user.Certificate.Id, &user.Certificate.MemberId, &certTitle, &user.Certificate.Certificate, &user.Certificate.Is_active, &user.Certificate.Date_added, &user.Id, &user.Username, &user.Language, &user.Timezone, &user.Is_staff, &user.Is_active, &user.Date_joined)
	if err == sql.ErrNoRows {
		return AskUser{}, false
	} else if err != nil {
		//panic(err)
		return AskUser{}, false
	}
	if certTitle != nil {
		user.Certificate.Title = certTitle.(string)
	}

	return user, true
}

func RegisterUser(request sis.Request, conn *sql.DB, username string, certHash string) {
	username = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(username, "register"), "?"), "+"))
	// Ensure user doesn't already exist
	row := conn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM membercerts WHERE certificate=?", certHash)

	var numRows int
	err := row.Scan(&numRows)
	if err != nil {
		panic(err)
	}
	if numRows < 1 {
		// Certificate doesn't already exist - Register User by adding the member first, then the certificate after getting the memberid
		zone, _ := time.Now().Zone()

		// TODO: Handle row2.Error and scan error
		var user AskUser
		row2 := conn.QueryRowContext(context.Background(), "INSERT INTO members (username, language, timezone, is_staff, is_active, date_joined) VALUES (?, ?, ?, ?, ?, ?) returning id, username, language, timezone, is_staff, is_active, date_joined", username, "en-US", zone, false, true, time.Now())
		row2.Scan(&user.Id, &user.Username, &user.Language, &user.Timezone, &user.Is_staff, &user.Is_active, &user.Date_joined)

		// TODO: Handle row3.Error and scan error
		var cert AskUserCert
		row3 := conn.QueryRowContext(context.Background(), "INSERT INTO membercerts (memberid, certificate, is_active, date_added) VALUES (?, ?, ?, ?) returning id, memberid, title, certificate, is_active, date_added", user.Id, certHash, true, time.Now())
		row3.Scan(&cert.Id, &cert.MemberId, &cert.Title, &cert.Is_active, &cert.Date_added)
		user.Certificate = cert
	}

	request.Redirect("/~ask/")
}

func GetUserQuestions(conn *sql.DB, user AskUser) []Question {
	query := "SELECT questions.id, questions.topicid, questions.title, questions.text, questions.tags, questions.memberid, questions.date_added FROM questions WHERE questions.memberid=? ORDER BY questions.date_added DESC"
	rows, rows_err := conn.QueryContext(context.Background(), query, user.Id)

	var questions []Question
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			question, scan_err := scanQuestionRows(rows)
			if scan_err == nil {
				question.User = user
				questions = append(questions, question)
			} else {
				panic(scan_err)
			}
		}
	}

	return questions
}

// With Totals
func GetTopics(conn *sql.DB) []Topic {
	query := "SELECT topics.ID, topics.title, topics.description, topics.date_added, COUNT(questions.id) FROM topics LEFT JOIN questions ON questions.topicid=topics.id GROUP BY topics.ID, topics.title, topics.description, topics.date_added ORDER BY topics.ID"
	rows, rows_err := conn.QueryContext(context.Background(), query)

	var topics []Topic
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var topic Topic
			scan_err := rows.Scan(&topic.Id, &topic.Title, &topic.Description, &topic.Date_added, &topic.QuestionTotal)
			if scan_err == nil {
				topics = append(topics, topic)
			} else {
				panic(scan_err)
			}
		}
	}

	return topics
}

func getTopic(conn *sql.DB, topicid int) (Topic, bool) {
	query := "SELECT * FROM topics WHERE id=?"
	row := conn.QueryRowContext(context.Background(), query, topicid)

	var topic Topic
	err := row.Scan(&topic.Id, &topic.Title, &topic.Description, &topic.Date_added)
	if err == sql.ErrNoRows {
		return Topic{}, false
	} else if err != nil {
		return Topic{}, false
	}

	return topic, true
}

func getRecentActivity_dates(conn *sql.DB) []time.Time {
	query := `SELECT DISTINCT cast(a.activity_date as date)
	FROM (SELECT questions.date_added as activity_date FROM questions
	UNION ALL
	SELECT answers.date_added as activity_date FROM answers LEFT JOIN questions ON questions.id=answers.questionid) a
	ORDER BY a.activity_date DESC`

	rows, rows_err := conn.QueryContext(context.Background(), query)

	var times []time.Time
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var t time.Time
			scan_err := rows.Scan(&t)
			if scan_err == nil {
				times = append(times, t)
			} else {
				panic(scan_err)
			}
		}
	}

	return times
}

func getRecentActivity_Questions(conn *sql.DB) []Activity {
	query := `SELECT a.id, a.topicid, a.title, a.text, a.tags, a.memberid, a.activity, a.activity_date, a.answerid, members.*, topics.title
FROM (SELECT questions.id, questions.topicid, questions.title, questions.text, questions.tags, questions.memberid, 'question' as activity, questions.date_added as activity_date, 0 as answerid FROM questions
	UNION ALL
	SELECT questions.id, questions.topicid, questions.title, questions.text, questions.tags, answers.memberid, 'answer' as activity, answers.date_added as activity_date, answers.id as answerid FROM answers LEFT JOIN questions ON questions.id=answers.questionid) a
LEFT JOIN members ON members.id=a.memberid
LEFT JOIN topics ON a.topicid=topics.id
ORDER BY a.activity_date DESC`

	rows, rows_err := conn.QueryContext(context.Background(), query)

	var activities []Activity
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			activity, scan_err := scanActivityWithUser(rows)
			if scan_err == nil {
				activities = append(activities, activity)
			} else {
				panic(scan_err)
			}
		}
	}

	return activities
}

func getRecentActivityFromDate_Questions(conn *sql.DB, date time.Time) []Activity {
	query := `SELECT a.id, a.topicid, a.title, a.text, a.tags, a.memberid, a.activity, a.activity_date, a.answerid, members.*, topics.title
FROM (SELECT questions.id, questions.topicid, questions.title, questions.text, questions.tags, questions.memberid, 'question' as activity, questions.date_added as activity_date, 0 as answerid FROM questions
	UNION ALL
	SELECT questions.id, questions.topicid, questions.title, questions.text, questions.tags, answers.memberid, 'answer' as activity, answers.date_added as activity_date, answers.id as answerid FROM answers LEFT JOIN questions ON questions.id=answers.questionid) a
LEFT JOIN members ON members.id=a.memberid
LEFT JOIN topics ON a.topicid=topics.id
WHERE cast(a.activity_date as date) = ?
ORDER BY a.activity_date DESC`

	rows, rows_err := conn.QueryContext(context.Background(), query, date)

	var activities []Activity
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			activity, scan_err := scanActivityWithUser(rows)
			if scan_err == nil {
				activities = append(activities, activity)
			} else {
				panic(scan_err)
			}
		}
	}

	return activities
}

func getQuestionsForTopic(conn *sql.DB, topicid int) []Question {
	query := "SELECT questions.id, questions.topicid, questions.title, questions.text, questions.tags, questions.memberid, questions.date_added, members.id, members.username, members.language, members.timezone, members.is_staff, members.is_active, members.date_joined FROM questions LEFT JOIN members ON members.id=questions.memberid WHERE questions.topicid=? ORDER BY questions.date_added DESC"
	rows, rows_err := conn.QueryContext(context.Background(), query, topicid)

	var questions []Question
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			question, scan_err := scanQuestionRowsWithUser(rows)
			if scan_err == nil {
				questions = append(questions, question)
			} else {
				panic(scan_err)
			}
		}
	}

	return questions
}

func getQuestion(conn *sql.DB, topicid int, questionid int) (Question, bool) {
	// TODO: Get Selected Answer as well
	query := "SELECT questions.id, questions.topicid, questions.title, questions.text, questions.tags, questions.memberid, questions.date_added, members.id, members.username, members.language, members.timezone, members.is_staff, members.is_active, members.date_joined FROM questions LEFT JOIN members ON members.id=questions.memberid WHERE questions.id=? AND questions.topicid=?"
	row := conn.QueryRowContext(context.Background(), query, questionid, topicid)

	question, err := scanQuestionWithUser(row)
	if err == sql.ErrNoRows {
		return Question{}, false
	} else if err != nil {
		return Question{}, false
	}

	return question, true
}

func createQuestionWithTitle(conn *sql.DB, topicid int, title string, user AskUser) (Question, error) {
	query := "INSERT INTO questions (topicid, title, memberid, date_added) VALUES (?, ?, ?, CURRENT_TIMESTAMP) RETURNING id, topicid, title, text, tags, memberid, date_added"
	row := conn.QueryRowContext(context.Background(), query, topicid, title, user.Id)
	return scanQuestion(row)
}
func createQuestionWithText(conn *sql.DB, topicid int, text string, user AskUser) (Question, error) {
	query := "INSERT INTO questions (topicid, text, memberid, date_added) VALUES (?, ?, ?, CURRENT_TIMESTAMP) RETURNING id, topicid, title, text, tags, memberid, date_added"
	row := conn.QueryRowContext(context.Background(), query, topicid, text, user.Id)
	return scanQuestion(row)
}
func createQuestionTitan(conn *sql.DB, topicid int, title string, text string, user AskUser) (Question, error) {
	query := "INSERT INTO questions (topicid, title, text, memberid, date_added) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP) RETURNING id, topicid, title, text, tags, memberid, date_added"
	row := conn.QueryRowContext(context.Background(), query, topicid, title, text, user.Id)
	return scanQuestion(row)
}

func updateQuestionTitle(conn *sql.DB, question Question, title string, user AskUser) (Question, error) {
	query := "UPDATE questions SET title=? WHERE id=? AND memberid=? RETURNING id, topicid, title, text, tags, memberid, date_added"
	row := conn.QueryRowContext(context.Background(), query, title, question.Id, user.Id)
	return scanQuestion(row)
}
func updateQuestionText(conn *sql.DB, question Question, text string, user AskUser) (Question, error) {
	query := "UPDATE questions SET text=? WHERE id=? AND memberid=? RETURNING id, topicid, title, text, tags, memberid, date_added"
	row := conn.QueryRowContext(context.Background(), query, text, question.Id, user.Id)
	return scanQuestion(row)
}

func getAnswersForQuestion(conn *sql.DB, question Question) []Answer {
	query := "SELECT answers.*, members.id, members.username, members.language, members.timezone, members.is_staff, members.is_active, members.date_joined FROM answers LEFT JOIN members ON answers.memberid=members.id WHERE answers.questionid=? ORDER BY date_added ASC"
	rows, rows_err := conn.QueryContext(context.Background(), query, question.Id)

	var answers []Answer
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			answer, scan_err := scanAnswerRows(rows)
			if scan_err == nil {
				answers = append(answers, answer)
			} else {
				panic(scan_err)
			}
		}
	}

	return answers
}

func getAnswer(conn *sql.DB, questionid int, answerid int) (Answer, bool) {
	// TODO: Get Selected Answer as well
	query := "SELECT answers.id, answers.questionid, answers.text, answers.gemlog_url, answers.memberid, answers.date_added, members.id, members.username, members.language, members.timezone, members.is_staff, members.is_active, members.date_joined FROM answers LEFT JOIN members ON members.id=answers.memberid WHERE answers.id=? AND answers.questionid=?"
	row := conn.QueryRowContext(context.Background(), query, answerid, questionid)

	answer, err := scanAnswerWithUser(row)
	if err == sql.ErrNoRows {
		return Answer{}, false
	} else if err != nil {
		return Answer{}, false
	}

	return answer, true
}

func createAnswerWithText(conn *sql.DB, questionid int, text string, user AskUser) (Answer, error) {
	query := "INSERT INTO answers (questionid, text, memberid, date_added) VALUES (?, ?, ?, CURRENT_TIMESTAMP) RETURNING id, questionid, text, gemlog_url, memberid, date_added"
	row := conn.QueryRowContext(context.Background(), query, questionid, text, user.Id)
	return scanAnswer(row)
}
func createAnswerAsGemlog(conn *sql.DB, questionid int, url string, user AskUser) (Answer, error) {
	query := "INSERT INTO answers (questionid, gemlog_url, memberid, date_added) VALUES (?, ?, ?, CURRENT_TIMESTAMP) RETURNING id, questionid, text, gemlog_url, memberid, date_added"
	row := conn.QueryRowContext(context.Background(), query, questionid, url, user.Id)
	return scanAnswer(row)
}

func updateAnswerText(conn *sql.DB, answer Answer, text string, user AskUser) (Answer, error) {
	query := "UPDATE answers SET text=? WHERE id=? AND memberid=? RETURNING id, questionid, text, gemlog_url, memberid, date_added"
	row := conn.QueryRowContext(context.Background(), query, text, answer.Id, user.Id)
	return scanAnswer(row)
}

func getUpvotesWithUsers(conn *sql.DB, answer Answer) []Upvote {
	query := "SELECT upvotes.id, upvotes.answerid, upvotes.memberid, upvotes.date_added, members.id, members.username, members.language, members.timezone, members.is_staff, members.is_active, members.date_joined FROM upvotes LEFT JOIN members ON members.id=upvotes.memberid WHERE upvotes.answerid=? ORDER BY upvotes.date_added"
	rows, rows_err := conn.QueryContext(context.Background(), query, answer.Id)

	var upvotes []Upvote
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			upvote, scan_err := scanUpvoteRowsWithUser(rows)
			if scan_err == nil {
				upvotes = append(upvotes, upvote)
			} else {
				panic(scan_err)
			}
		}
	}

	return upvotes
}

func addUpvote(conn *sql.DB, answer Answer, user AskUser) (Upvote, error) {
	query := `UPDATE OR INSERT INTO upvotes (answerid, memberid, date_added)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	MATCHING (answerid, memberid)
	RETURNING id, answerid, memberid, date_added`
	row := conn.QueryRowContext(context.Background(), query, answer.Id, user.Id)
	return scanUpvote(row)
}

func removeUpvote(conn *sql.DB, answer Answer, user AskUser) error {
	query := `DELETE FROM upvotes WHERE answerid=? AND memberid=?`
	_, err := conn.ExecContext(context.Background(), query, answer.Id, user.Id)
	return err
}
