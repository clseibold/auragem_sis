package crawler

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func MediatypeIsTextual(mediatype string) bool {
	if strings.HasPrefix(mediatype, "text/") {
		return true
	}
	switch mediatype {
	case "application/ecmascript", "application/javascript", "application/x-ecmascript", "application/x-javascript", "application/json", "application/mbox", "application/postscript", "application/prql", "application/sparql-query", "application/srgs", "application/x-perl", "application/x-sh", "application/x-shar", "application/x-shellscript", "application/x-tcl", "application/x-tex", "application/x-texinfo", "application/xml", "application/xml-dtd", "application/yaml", "application/x-yaml", "chemical/x-cif", "chemical/x-cml", "chemical/x-csml", "chemical/x-xyz":
		return true
	}

	return false
}

// CutAny slices s around any Unicode code point from chars,
// returning the text before and after it. The found result
// reports whether any Unicode code point was appears in s.
// If it does not appear in s, CutAny returns s, "", false.
func CutAny(s string, chars string) (before string, after string, found bool) {
	if index := strings.IndexAny(s, chars); index >= 0 {
		return s[:index], strings.TrimLeft(s[index:], chars), true
	}
	return s, "", false
}

func filter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func isTimeDate(s string) bool {
	name := strings.TrimSpace(s)
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return false
	}

	_, timeParseErr := time.Parse(ISO8601Layout, parts[0])
	if timeParseErr == nil {
		return true
	}
	_, timeParseErr = time.Parse(time.RFC3339, parts[0])
	if timeParseErr == nil {
		return true
	}
	_, timeParseErr = time.Parse("2006-01-02", parts[0])
	return timeParseErr == nil
	/*if timeParseErr == nil {
		return true
	}

	return false*/
}

// NOTE: Must be utf-8 string
func getTimeDate(s string, file bool) time.Time {
	if len(s) == 0 {
		return time.Time{}
	}
	name := strings.TrimSpace(s)
	var timeString string
	if file {
		firstChar, _ := utf8.DecodeRuneInString(s)
		if !unicode.IsDigit(firstChar) {
			return time.Time{}
		}
		parts := strings.FieldsFunc(name, func(r rune) bool {
			if r == '-' || r == ':' {
				return false
			}
			if r == '_' || unicode.IsLetter(r) {
				return true
			}
			return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || !unicode.IsPrint(r)
		})
		if len(parts) == 0 {
			return time.Time{}
		}
		timeString = strings.TrimRight(parts[0], "-_")
	} else {
		parts := strings.Fields(name)
		if len(parts) == 0 {
			return time.Time{}
		}
		timeString = parts[0]
	}

	t, timeParseErr := time.Parse(ISO8601Layout, timeString)
	if timeParseErr == nil {
		return t
	}
	t, timeParseErr = time.Parse(time.RFC3339, timeString)
	if timeParseErr == nil {
		return t
	}
	if file {
		t, timeParseErr = time.Parse("2006-01-02-15-04", timeString)
		if timeParseErr == nil {
			return t
		}
		t, timeParseErr = time.Parse("2006-01-02_15-04", timeString)
		if timeParseErr == nil {
			return t
		}
	}
	t, timeParseErr = time.Parse("2006-01-02", timeString)
	if timeParseErr == nil {
		return t
	}
	if file {
		t, timeParseErr = time.Parse("2006_01_02_15_04", timeString)
		if timeParseErr == nil {
			return t
		}
		t, timeParseErr = time.Parse("2006_01_02", timeString)
		if timeParseErr == nil {
			return t
		}
	}
	t, timeParseErr = time.Parse("20060102", timeString)
	if timeParseErr == nil {
		return t
	}
	t, timeParseErr = time.Parse("200601021504", timeString)
	if timeParseErr == nil {
		return t
	}

	return time.Time{}
}

// Does not trim the '#' prefix
func GetTagsFromText(s string) []string {
	trimmed := strings.ToLower(strings.TrimSpace(s))
	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		if r == '#' || r == '*' || r == '+' || r == '-' || r == '=' || r == '_' || r == ':' || r == '\'' {
			return false
		}
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || !unicode.IsPrint(r)
	})

	return filter(parts, func(s string) bool {
		return s != "" && s != "#" && strings.HasPrefix(s, "#")
	})
}

// Does not trim the '~' or '@' prefix
func GetMentionsFromText(s string) []string {
	trimmed := strings.ToLower(strings.TrimSpace(s))
	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		if r == '@' || r == '#' || r == '*' || r == '+' || r == '-' || r == '=' || r == '_' || r == ':' || r == '~' {
			return false
		}
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || !unicode.IsPrint(r)
	})

	return filter(parts, func(s string) bool {
		return s != "" && s != "@" && s != "~" && s != "@@" && s != "~~" && (strings.HasPrefix(s, "@") || strings.HasPrefix(s, "~"))
	})
}

// If the string contains any runes of the unicode L (Letter) category, excluding spaces
func ContainsLetterRunes(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsSpace(r) {
			return true
		}
	}

	return false
}
