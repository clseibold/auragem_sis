package music

import (
	"context"
	crypto_rand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

type MusicUser struct {
	Id          int64
	Username    string
	Certificate string
	Language    string
	Timezone    string
	Is_staff    bool
	Is_active   bool
	Date_joined time.Time

	QuotaCount float64
}
type MusicFile struct {
	Id               int64
	Filehash         string
	Filename         string
	Mimetype         string
	Title            string
	Album            string
	Artist           string
	Albumartist      string
	Composer         string
	Genre            string
	Releaseyear      int
	Tracknumber      int
	Discnumber       int
	UploadCount      int
	AllowPublicRadio bool
	CbrKbps          int64
	Attribution      string
	Date_added       time.Time
}
type MusicAlbum struct {
	Album       string
	Albumartist string
}

func GetFileInLibrary_hash(conn *sql.DB, hash string) (MusicFile, bool) {
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM library WHERE filehash=?`

	row := conn.QueryRowContext(context.Background(), query, hash)

	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows {
		return MusicFile{}, false
	}
	return file, true
}

func GetFileInLibrary_id(conn *sql.DB, fileid int64) (MusicFile, bool) {
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM library WHERE id=?`

	row := conn.QueryRowContext(context.Background(), query, fileid)

	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows {
		return MusicFile{}, false
	}
	return file, true
}

func AddFileToLibrary(conn *sql.DB, hash string, m tag.Metadata, check bool) (MusicFile, bool) {
	albumartist := m.AlbumArtist()
	if albumartist == "" {
		albumartist = "Unknown Album Artist"
	}

	// Check first if library already has the file
	exists := false
	if check {
		_, exists = GetFileInLibrary_hash(conn, hash)
	}

	if !exists {
		var cbr_bitrate int64 = 0
		bitrate_str := strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(m.Raw()["stream_bitrate"].(string), "kbps CBR"), "kbps VBR"))
		var err error
		cbr_bitrate, err = strconv.ParseInt(bitrate_str, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing int from string %q: %v\n", bitrate_str, err)
			return MusicFile{}, false
		}

		query := `INSERT INTO library (FILEHASH, FILENAME, MIMETYPE, TITLE, ALBUM, ARTIST, ALBUMARTIST, COMPOSER, GENRE, RELEASEYEAR, TRACKNUMBER, DISCNUMBER, UPLOADCOUNT, CBR_KBPS, DATE_ADDED) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		trackNumber, _ := m.Track()
		discNumber, _ := m.Disc()
		conn.ExecContext(context.Background(), query, hash, hash+".mp3", "audio/mpeg", m.Title(), m.Album(), m.Artist(), albumartist, m.Composer(), m.Genre(), m.Year(), trackNumber, discNumber, 0, cbr_bitrate, time.Now())
	}

	return GetFileInLibrary_hash(conn, hash)
}

func GetFileInUserLibrary(conn *sql.DB, musicFileId int64, userId int64) (MusicFile, bool) {
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE library.id=? AND uploads.memberid=?`

	row := conn.QueryRowContext(context.Background(), query, musicFileId, userId)

	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, false
	}
	return file, true
}

func GetFileInUserLibrary_hash(conn *sql.DB, musicFileHash string, userId int64) (MusicFile, bool) {
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE library.filehash=? AND uploads.memberid=?`

	row := conn.QueryRowContext(context.Background(), query, musicFileHash, userId)

	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, false
	}
	return file, true
}

func AddFileToUserLibrary(conn *sql.DB, musicFileId int64, userId int64, check bool) {
	// Check first if user already has the file
	exists := false
	if check {
		_, exists = GetFileInUserLibrary(conn, musicFileId, userId)
	}

	if !exists {
		conn.ExecContext(context.Background(), "INSERT INTO uploads (memberid, fileid, date_added) values (?, ?, ?)", userId, musicFileId, time.Now())
		conn.ExecContext(context.Background(), "UPDATE library SET uploadcount = uploadcount + 1 WHERE id=?", musicFileId)
	}
}

func RemoveFileFromUserLibrary(conn *sql.DB, musicFileId int64, userId int64) {
	conn.ExecContext(context.Background(), "DELETE FROM uploads WHERE uploads.memberid = ? AND uploads.fileid = ?", userId, musicFileId)
	conn.ExecContext(context.Background(), "UPDATE library SET uploadcount = uploadcount - 1 WHERE id=?", musicFileId)

	time.Sleep(20 * time.Millisecond)

	// TODO: Remove the file from library and harddrive if uploadcount is now 0
	// Get the file again
	file_updated, _ := GetFileInLibrary_id(conn, musicFileId)
	if file_updated.UploadCount == 0 && !file_updated.AllowPublicRadio { // NOTE: Don't delete if allowed on public radio
		// If Upload Count is 0, then delete the file from the db and from the Samba Share.
		conn.ExecContext(context.Background(), "DELETE FROM library WHERE library.id = ?", file_updated.Id)
		//SMB_DeleteFile(file_updated.Filename)
		remove_err := os.Remove(musicDirectory + file_updated.Filename)
		if remove_err != nil {
			panic(remove_err)
		}
	}
}

func cryptoRandomSeed() int64 {
	var b [8]byte
	_, r_err := crypto_rand.Read(b[:])
	if r_err != nil {
		panic("cannot create seed with cryptographically secure random number generator.")
	}
	return int64(binary.LittleEndian.Uint64(b[:]))
}

func GetRandomFileInUserLibray(conn *sql.DB, userId int64) (MusicFile, bool) {
	randomSeed := cryptoRandomSeed()
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? ORDER BY (library.id + cast(? as bigint))*4294967291-((library.id + cast(? as bigint))*4294967291/49157)*49157`
	row := conn.QueryRowContext(context.Background(), query, userId, randomSeed, randomSeed)

	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, false
	}
	return file, true
}

// Will exclude one music file (use this when you don't want the next random to not match the previous one)
func GetRandomFileInUserLibray_excludeId(conn *sql.DB, userId int64, exclude_id int64) (MusicFile, bool) {
	randomSeed := cryptoRandomSeed()
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? AND library.id<>? ORDER BY (library.id + cast(? as bigint))*4294967291-((library.id + cast(? as bigint))*4294967291/49157)*49157`
	row := conn.QueryRowContext(context.Background(), query, userId, exclude_id, randomSeed, randomSeed)

	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, false
	}
	return file, true
}

// NOTE: Will exclude BeOS and Classical music (to get these, use GetRandomPublicDomainFileInLibrary_RadioGenre instead)
func GetRandomPublicDomainFileInLibrary(conn *sql.DB, exclude_ids []int64) (MusicFile, bool) {
	var builder strings.Builder
	for _, id := range exclude_ids {
		fmt.Fprintf(&builder, "AND library.id<>%d ", id)
	}

	randomSeed := cryptoRandomSeed()
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM library WHERE library.allowpublicradio=true AND library.radio_genre<>'BeOS' AND library.radio_genre<>'Classical' ` + builder.String() + ` ORDER BY (library.id + cast(? as bigint))*4294967291-((library.id + cast(? as bigint))*4294967291/49157)*49157`
	row := conn.QueryRowContext(context.Background(), query, randomSeed, randomSeed)

	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, false
	}
	return file, true
}

// NOTE: Prefer GetRandomPublicDomainFileInLibrary_RadioStation over this
func GetRandomPublicDomainFileInLibrary_RadioGenre(conn *sql.DB, exclude_ids []int64, radioGenre string) (MusicFile, bool) {
	if radioGenre == "Any" {
		return GetRandomPublicDomainFileInLibrary(conn, exclude_ids)
	}

	var builder strings.Builder
	for _, id := range exclude_ids {
		fmt.Fprintf(&builder, "AND library.id<>%d ", id)
	}

	randomSeed := cryptoRandomSeed()
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM library WHERE library.allowpublicradio=true AND radio_genre=? ` + builder.String() + ` ORDER BY (library.id + cast(? as bigint))*4294967291-((library.id + cast(? as bigint))*4294967291/49157)*49157`
	row := conn.QueryRowContext(context.Background(), query, radioGenre, randomSeed, randomSeed)

	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, false
	}
	return file, true
}

func GetRandomPublicDomainFileInLibrary_RadioStation_Any(conn *sql.DB, exclude_ids []int64, station *RadioStation) (MusicFile, bool) {
	var builder strings.Builder
	for _, id := range exclude_ids {
		fmt.Fprintf(&builder, "AND library.id<>%d ", id)
	}

	fmt.Fprintf(&builder, "AND (")
	for i, genre := range station.AnyCategory {
		fmt.Fprintf(&builder, "library.radio_genre='%s' ", genre)
		if i < len(station.AnyCategory)-1 {
			fmt.Fprintf(&builder, "OR ")
		}
	}
	fmt.Fprintf(&builder, ")")

	randomSeed := cryptoRandomSeed()
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM library WHERE library.allowpublicradio=true ` + builder.String() + ` ORDER BY (library.id + cast(? as bigint))*4294967291-((library.id + cast(? as bigint))*4294967291/49157)*49157`
	//fmt.Printf("Query: %s\n", query)
	row := conn.QueryRowContext(context.Background(), query, randomSeed, randomSeed)
	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, false
	}

	return file, true
}

func GetRandomPublicDomainFileInLibrary_RadioStation_Announcer(conn *sql.DB, station *RadioStation) (MusicFile, bool) {
	var builder strings.Builder

	fmt.Fprintf(&builder, "AND library.radio_genre='Announcer' AND library.title='%s Announcer' ", station.Name)

	randomSeed := cryptoRandomSeed()
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM library WHERE library.allowpublicradio=true ` + builder.String() + ` ORDER BY (library.id + cast(? as bigint))*4294967291-((library.id + cast(? as bigint))*4294967291/49157)*49157`
	//fmt.Printf("Query: %s\n", query)
	row := conn.QueryRowContext(context.Background(), query, randomSeed, randomSeed)
	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, false
	}

	return file, true
}

func GetRandomPublicDomainFileInLibrary_RadioStation(conn *sql.DB, exclude_ids []int64, station *RadioStation) (MusicFile, string, bool) {
	currentTime := time.Now()
	radioGenre := GetRadioGenre(currentTime, station)

	var builder strings.Builder
	for _, id := range exclude_ids {
		fmt.Fprintf(&builder, "AND library.id<>%d ", id)
	}

	if radioGenre == "Any" {
		//return GetRandomPublicDomainFileInLibrary(conn, exclude_ids)
		fmt.Fprintf(&builder, "AND (")
		for i, genre := range station.AnyCategory {
			fmt.Fprintf(&builder, "library.radio_genre='%s' ", genre)
			if i < len(station.AnyCategory)-1 {
				fmt.Fprintf(&builder, "OR ")
			}
		}
		fmt.Fprintf(&builder, ")")
	} else if radioGenre == "OTR-Program" || radioGenre == "OTR-Program-Rerun" {
		// Get program for weekday
		program := station.ProgramInfo[currentTime.Weekday()]
		fmt.Printf("Getting program: %s\n", program)

		// Check the current episode from the station's map
		currentEpisode, ok := station.CurrentEpisode[program]
		if !ok || currentEpisode == 0 {
			station.CurrentEpisode[program] = 1
			currentEpisode = 1
		}
		if radioGenre == "OTR-Program-Rerun" {
			// If in rerun spot, play the previous episode instead
			currentEpisode -= 1
		}

		// Get the episode from the database, using a mod to wrap around the total number of episodes. if resulting episode number is different (it wrapped around), then set the current episode to the proper number
		fmt.Fprintf(&builder, "AND library.radio_genre='%s' AND library.album='%s' AND library.tracknumber=(Mod((%d - 1), (select COUNT(*) from library l where l.radio_genre='%s' AND l.album = '%s')) + 1) ", "OTR-Program", program, currentEpisode, "OTR-Program", program)
	} else {
		fmt.Fprintf(&builder, "AND library.radio_genre='%s' ", radioGenre)
	}

	randomSeed := cryptoRandomSeed()
	query := `SELECT FIRST 1 library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM library WHERE library.allowpublicradio=true ` + builder.String() + ` ORDER BY (library.id + cast(? as bigint))*4294967291-((library.id + cast(? as bigint))*4294967291/49157)*49157`
	//fmt.Printf("Query: %s\n", query)
	row := conn.QueryRowContext(context.Background(), query, randomSeed, randomSeed)
	var file MusicFile
	err := row.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
	if err == sql.ErrNoRows || err != nil {
		return MusicFile{}, radioGenre, false
	}

	if radioGenre == "OTR-Program" {
		//program := station.ProgramInfo[currentTime.Weekday()]
		//station.CurrentEpisode[program] = file.Tracknumber

		// If last Program re-run of the day, switch the current episode to the next episode
		program := station.ProgramInfo[currentTime.Weekday()]
		station.CurrentEpisode[program] = file.Tracknumber + 1
	}

	return file, radioGenre, true
}

func GetFilesInGenre(conn *sql.DB, radioGenre string) []MusicFile {
	var musicFiles []MusicFile
	query := `SELECT library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM library WHERE library.allowpublicradio=true AND library.radio_genre=? ORDER BY library.albumartist ASC, library.releaseyear DESC, library.album ASC, library.discnumber ASC, library.tracknumber ASC`
	rows, rows_err := conn.QueryContext(context.Background(), query, radioGenre)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var file MusicFile
			scan_err := rows.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
			if scan_err == nil {
				musicFiles = append(musicFiles, file)
			}
		}
	}

	return musicFiles
}

// ---------

func GetUser(conn *sql.DB, certHash string) (MusicUser, bool) {
	row := conn.QueryRowContext(context.Background(), "SELECT id, username, language, timezone, is_staff, is_active, date_joined FROM members WHERE certificate=?", certHash)

	var user MusicUser
	user.Certificate = certHash
	err := row.Scan(&user.Id, &user.Username, &user.Language, &user.Timezone, &user.Is_staff, &user.Is_active, &user.Date_joined)
	if err == sql.ErrNoRows {
		return MusicUser{}, false
	} else if err != nil {
		panic(err)
		//return MusicUser{}, false
	}

	// Get user quota
	quotaCount := GetUserQuota(conn, user.Id)
	user.QuotaCount = quotaCount

	return user, true
}

func GetFilesInUserLibrary(conn *sql.DB, userId int64) []MusicFile {
	var musicFiles []MusicFile
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? ORDER BY library.albumartist ASC, library.releaseyear DESC, library.album ASC, library.discnumber ASC, library.tracknumber ASC", userId)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var file MusicFile
			scan_err := rows.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
			if scan_err == nil {
				musicFiles = append(musicFiles, file)
			}
		}
	}

	return musicFiles
}

func GetArtistsInUserLibrary(conn *sql.DB, userId int64) []string {
	var artists []string
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT DISTINCT library.albumartist FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? ORDER By library.albumartist ASC", userId)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var artist string
			scan_err := rows.Scan(&artist)
			if scan_err == nil {
				artists = append(artists, artist)
			}
		}
	}

	return artists
}

func GetAlbumsInUserLibrary(conn *sql.DB, userId int64) []MusicAlbum {
	var albums []MusicAlbum
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT DISTINCT library.album, library.albumartist FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? ORDER BY library.albumartist, library.releaseyear DESC, library.album ASC", userId)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var album MusicAlbum
			scan_err := rows.Scan(&album.Album, &album.Albumartist)
			if scan_err == nil {
				albums = append(albums, album)
			}
		}
	}

	return albums
}

func GetAlbumsFromArtistInUserLibrary(conn *sql.DB, userId int64, albumArtist string) []MusicAlbum {
	var albums []MusicAlbum
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT DISTINCT library.album, library.albumartist FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? AND library.albumartist=? ORDER BY library.releaseyear DESC, library.album ASC", userId, albumArtist)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var album MusicAlbum
			scan_err := rows.Scan(&album.Album, &album.Albumartist)
			if scan_err == nil {
				albums = append(albums, album)
			}
		}
	}

	return albums
}

func GetFilesFromAlbumInUserLibrary(conn *sql.DB, userId int64, albumArtist string, album string) []MusicFile {
	var musicFiles []MusicFile
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? AND albumartist=? AND album=? ORDER BY library.discnumber, library.tracknumber ASC", userId, albumArtist, album)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var file MusicFile
			scan_err := rows.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
			if scan_err == nil {
				musicFiles = append(musicFiles, file)
			}
		}
	}

	return musicFiles
}

// TODO: Add randomization to order of albums?
func GetFilesFromArtistInUserLibrary(conn *sql.DB, userId int64, albumArtist string) []MusicFile {
	var musicFiles []MusicFile
	rows, rows_err := conn.QueryContext(context.Background(), "SELECT library.id, library.filehash, library.filename, library.mimetype, library.title, library.album, library.artist, library.albumartist, library.composer, library.genre, library.releaseyear, library.tracknumber, library.discnumber, library.uploadcount, library.allowpublicradio, library.cbr_kbps, library.attribution, library.date_added FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? AND albumartist=? ORDER BY library.album ASC, library.discnumber, library.tracknumber ASC", userId, albumArtist)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var file MusicFile
			scan_err := rows.Scan(&file.Id, &file.Filehash, &file.Filename, &file.Mimetype, &file.Title, &file.Album, &file.Artist, &file.Albumartist, &file.Composer, &file.Genre, &file.Releaseyear, &file.Tracknumber, &file.Discnumber, &file.UploadCount, &file.AllowPublicRadio, &file.CbrKbps, &file.Attribution, &file.Date_added)
			if scan_err == nil {
				musicFiles = append(musicFiles, file)
			}
		}
	}

	return musicFiles
}

func GetUserQuota(conn *sql.DB, userId int64) float64 {
	q := `SELECT COUNT(*) / (CAST(SUM(library.uploadcount) as float) / COUNT(*) * 1.0) FROM uploads INNER JOIN library ON uploads.fileid=library.id WHERE uploads.memberid=? GROUP BY library.uploadcount`
	rows, rows_err := conn.QueryContext(context.Background(), q, userId)
	if rows_err == nil {
		defer rows.Close()
		quotaCount := 0.0
		for rows.Next() {
			var i float64
			scan_err := rows.Scan(&i)
			if scan_err == nil {
				quotaCount += i
			}
		}

		return quotaCount
	}

	return 0
}

// ------ Admin stuff

func Admin_GetGlobalQuota(conn *sql.DB) float64 {
	q := `SELECT COUNT(DISTINCT filehash) FROM library`
	row := conn.QueryRowContext(context.Background(), q)
	var quotaCount float64
	err := row.Scan(&quotaCount)
	if err != nil {
		panic(err)
		//return 0
	}

	return quotaCount
}

func Admin_UserCount(conn *sql.DB) int {
	q := `SELECT COUNT(DISTINCT id) FROM members`
	row := conn.QueryRowContext(context.Background(), q)
	var userCount int
	err := row.Scan(&userCount)
	if err != nil {
		panic(err)
		//return 0
	}

	return userCount
}

func Admin_ArtistCount(conn *sql.DB) int {
	q := `SELECT COUNT(DISTINCT albumartist) FROM library`
	row := conn.QueryRowContext(context.Background(), q)
	var count int
	err := row.Scan(&count)
	if err != nil {
		panic(err)
		//return 0
	}

	return count
}

func Admin_AlbumCount(conn *sql.DB) int {
	q := `SELECT COUNT(DISTINCT album || albumartist) FROM library`
	row := conn.QueryRowContext(context.Background(), q)
	var count int
	err := row.Scan(&count)
	if err != nil {
		panic(err)
		//return 0
	}

	return count
}

type RadioGenre struct {
	Name  string
	Count int64
}

func Admin_RadioGenreCounts(conn *sql.DB) []RadioGenre {
	var radioGenres []RadioGenre
	q := `SELECT radio_genre, COUNT(*) FROM library WHERE allowpublicradio=true GROUP BY radio_genre ORDER BY COUNT(*) DESC`
	rows, rows_err := conn.QueryContext(context.Background(), q)
	if rows_err == nil {
		defer rows.Close()
		for rows.Next() {
			var radioGenre RadioGenre
			scan_err := rows.Scan(&radioGenre.Name, &radioGenre.Count)
			if scan_err == nil {
				radioGenres = append(radioGenres, radioGenre)
			}
		}
	}

	return radioGenres
}
