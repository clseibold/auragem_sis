package music

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
	"github.com/efarrer/iothrottler"
	sis "gitlab.com/sis-suite/smallnetinformationservices"
)

/*
func SMB_WriteFile(filename string, data []byte) {
	// SMB Server Connection
	// Connection Info
	//smb_conn, smb_err := net.Dial("tcp", "192.168.7.24:445")
	smb_conn, smb_err := net.Dial("tcp", "192.168.0.60:445")
	if smb_err != nil {
		panic(smb_err)
	}
	defer smb_conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     "krixa",
			Password: config.SMBPassword,
			Domain:   "WORKSTATION",
		},
	}

	// Dial the SMB server
	s, d_err := d.Dial(smb_conn)
	if d_err != nil {
		panic(d_err)
	}
	defer s.Logoff()

	// Mount ServerData Share
	fs, err := s.Mount("ServerData")
	if err != nil {
		panic(err)
	}
	defer fs.Umount()

	write_err := fs.WriteFile(musicDirectory_SMB+filename, data, 0600)
	if write_err != nil {
		panic(write_err)
	}
}
*/

/*
func SMB_DeleteFile(filename string) {
	// SMB Server Connection
	// Connection Info
	//smb_conn, smb_err := net.Dial("tcp", "192.168.7.24:445")
	smb_conn, smb_err := net.Dial("tcp", "192.168.0.60:445")
	if smb_err != nil {
		panic(smb_err)
	}
	defer smb_conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     "krixa",
			Password: config.SMBPassword,
			Domain:   "WORKSTATION",
		},
	}

	// Dial the SMB server
	s, d_err := d.Dial(smb_conn)
	if d_err != nil {
		panic(d_err)
	}
	defer s.Logoff()

	// Mount ServerData Share
	fs, err := s.Mount("ServerData")
	if err != nil {
		panic(err)
	}
	defer fs.Umount()

	remove_err := fs.Remove(musicDirectory_SMB + filename)
	if remove_err != nil {
		panic(remove_err)
	}
}
*/

func StreamFile(request *sis.Request, music_file MusicFile) error {
	throttlePool := iothrottler.NewIOThrottlerPool(iothrottler.Bandwidth(music_file.CbrKbps * 1000 / 8 * 2))
	defer throttlePool.ReleasePool()

	file, err := os.OpenFile(filepath.Join(musicDirectory, music_file.Filename), os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}

	streamBuffer := make([]byte, music_file.CbrKbps*1000/8)

	// Read 5 seconds worth of data at once to have a buffer (and reduce the effects of stuttering/buffering)
	buffer_5sec_backing := make([]byte, 0, music_file.CbrKbps*1000/8*5)
	buffer_5sec := bytes.NewBuffer(buffer_5sec_backing)
	_, copy_err := io.CopyN(buffer_5sec, file, music_file.CbrKbps*1000/8*5)
	if copy_err != nil {
		fmt.Printf("Copy error: %v\n", copy_err)
	}
	request.StreamBuffer("audio/mpeg", buffer_5sec, streamBuffer)

	// Throttle the rest of the file based on the bandwidth throttle of the pool
	file_throttled, throttle_err := throttlePool.AddReader(file)
	if throttle_err != nil {
		panic(throttle_err)
	}
	result_err := request.StreamBuffer("audio/mpeg", file_throttled, streamBuffer)
	file_throttled.Close()
	return result_err
}

func StreamMultipleFiles(request *sis.Request, musicFiles []MusicFile) error {
	throttlePool := iothrottler.NewIOThrottlerPool(320 * 1000 / 8)
	defer throttlePool.ReleasePool()

	var result_err error
	first := true
	streamBuffer := make([]byte, 96*1000/8) // Assume 96 kbps (min bitrate for mp3 files)
	for _, music_file := range musicFiles {
		openFile, err := os.OpenFile(filepath.Join(musicDirectory, music_file.Filename), os.O_RDONLY, 0600)
		if err != nil {
			panic(err)
		}

		// Skip ID3v2 Tags at start of file
		skip_err := tag.SkipID3v2Tags(openFile) // TODO
		if skip_err != nil {
			fmt.Printf("Failed to skip ID3 Headers\n")
		}

		// When the first file, read 5 seconds worth of data at once to have a buffer (and reduce the effects of stuttering/buffering)
		if first {
			buffer_5sec_backing := make([]byte, 0, music_file.CbrKbps*1000/8*5)
			buffer_5sec := bytes.NewBuffer(buffer_5sec_backing)
			_, copy_err := io.CopyN(buffer_5sec, openFile, music_file.CbrKbps*1000/8*5)
			if copy_err != nil {
				fmt.Printf("Copy error: %v\n", copy_err)
			}
			request.StreamBuffer("audio/mpeg", buffer_5sec, streamBuffer)
			first = false
		}

		throttlePool.SetBandwidth(iothrottler.Bandwidth(music_file.CbrKbps * 1000 / 8))

		// Throttle (the rest of) the file based on the bandwidth throttle of the pool
		throttledFile, throttle_err := throttlePool.AddReader(openFile)
		if throttle_err != nil {
			panic(throttle_err)
		}

		err2 := request.StreamBuffer("audio/mpeg", throttledFile, streamBuffer)
		if err2 != nil {
			//return err2
			result_err = err2
			// TODO: Break here? Don't want to open files and stream when it's not needed
		}
		throttledFile.Close()
	}
	return result_err
}

func StreamRandomFiles(request *sis.Request, conn *sql.DB, user MusicUser) error {
	throttlePool := iothrottler.NewIOThrottlerPool(320 * 1000 / 8)
	defer throttlePool.ReleasePool()

	streamBuffer := make([]byte, 96*1000/8) // Assume 96 kbps (min bitrate for mp3 files)

	var result_err error
	//first := true
	var lastFileId int64 = 0 // So the same song doesn't repeat twice in a row
	for {
		file, success := GetRandomFileInUserLibray_excludeId(conn, user.Id, lastFileId)
		if !success {
			fmt.Printf("Error getting random file from user's library\n")
			request.TemporaryFailure("Error getting random file from user's library. Make sure you have uploaded music first.")
			return fmt.Errorf("Error getting random file from user's library\n")
		}
		lastFileId = file.Id

		openFile, err := os.OpenFile(filepath.Join(musicDirectory, file.Filename), os.O_RDONLY, 0600)
		if err != nil {
			//panic(err)
			fmt.Printf("Filed to open file '%s': %v\n", filepath.Join(musicDirectory, file.Filename), err)
			continue
		}

		// Skip ID3v2 Tags at start of file
		skip_err := tag.SkipID3v2Tags(openFile) // TODO
		if skip_err != nil {
			fmt.Printf("Failed to skip ID3 Headers\n")
		}

		throttlePool.SetBandwidth(iothrottler.Bandwidth(file.CbrKbps * 1000 / 8))

		// When the first file, read 5 seconds worth of data at once to have a buffer (and reduce the effects of stuttering/buffering)
		/*if first {
			buffer_5sec_backing := make([]byte, 0, 40*1024*5)
			buffer_5sec := bytes.NewBuffer(buffer_5sec_backing)
			_, copy_err := io.CopyN(buffer_5sec, openFile, 40*1024*5)
			if copy_err != nil {
				fmt.Printf("Copy error: %v\n", copy_err)
			}
			c.Stream("audio/mpeg", buffer_5sec)
			first = false
		}*/

		// Throttle (the rest of) the file based on the bandwidth throttle of the pool
		throttledFile, throttle_err := throttlePool.AddReader(openFile)
		if throttle_err != nil {
			panic(throttle_err)
		}

		err2 := request.StreamBuffer("audio/mpeg", throttledFile, streamBuffer)
		if err2 != nil {
			fmt.Printf("Failed to stream file: '%s': %v\n", musicDirectory+file.Filename, err2)
			result_err = err2
			throttledFile.Close()
			break // make sure stream doesn't go on forever
			//return err2
		}
		throttledFile.Close()
	}

	return result_err
}
