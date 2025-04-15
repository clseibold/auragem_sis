package music

import (
	"bytes"
	"fmt"
	"io"
)

func SkipId3HeaderTags(r io.ReadSeeker) error {
	b, err := readBytes(r, 11)
	if err != nil {
		return err
	}

	_, err = r.Seek(-11, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("could not seek back to the original position: %v", err)
	}

	switch {
	case string(b[0:4]) == "fLaC":
		return fmt.Errorf("not an ID3 header: %v", err)
	case string(b[4:11]) == "ftypM4A":
		return fmt.Errorf("not an ID3 header: %v", err)
	case string(b[0:3]) == "ID3":
		header, offset, err := readID3v2Header(r)
		if err != nil {
			return fmt.Errorf("error reading ID3v2 header: %v", err)
		}

		// Seek to end of header
		start := int64(offset + header.Size)
		_, err = r.Seek(start, io.SeekStart)
		if err != nil {
			return fmt.Errorf("error seeking to end of ID3V2 header: %v", err)
		}

		// TODO: Handle XING header (and LAME header, which uses XING)

		/*n -= start
		if n < 0 {
			return "", fmt.Errorf("file size must be greater than 128 bytes for MP3: %v bytes", n)
		}*/
	}

	// Handle ID3v1 here:
	return nil
}

// readID3v2Header reads the ID3v2 header from the given io.Reader.
// offset it number of bytes of header that was read
func readID3v2Header(r io.Reader) (h *id3v2Header, offset uint, err error) {
	offset = 10
	b, err := readBytes(r, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("expected to read 10 bytes (ID3v2Header): %v", err)
	}

	if string(b[0:3]) != "ID3" {
		return nil, 0, fmt.Errorf("expected to read \"ID3\"")
	}

	b = b[3:]
	var vers Format
	switch uint(b[0]) {
	case 2:
		vers = ID3v2_2
	case 3:
		vers = ID3v2_3
	case 4:
		vers = ID3v2_4
	case 0, 1:
		fallthrough
	default:
		return nil, 0, fmt.Errorf("ID3 version: %v, expected: 2, 3 or 4", uint(b[0]))
	}

	// NB: We ignore b[1] (the revision) as we don't currently rely on it.
	h = &id3v2Header{
		Version:           vers,
		Unsynchronisation: getBit(b[2], 7),
		ExtendedHeader:    getBit(b[2], 6),
		Experimental:      getBit(b[2], 5),
		Size:              uint(get7BitChunkedInt(b[3:7])),
	}

	if h.ExtendedHeader {
		switch vers {
		case ID3v2_3:
			b, err := readBytes(r, 4)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read 4 bytes (ID3v23 extended header len): %v", err)
			}
			// skip header, size is excluding len bytes
			extendedHeaderSize := uint(getInt(b))
			_, err = readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read %d bytes (ID3v23 skip extended header): %v", extendedHeaderSize, err)
			}
			offset += extendedHeaderSize
		case ID3v2_4:
			b, err := readBytes(r, 4)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read 4 bytes (ID3v24 extended header len): %v", err)
			}
			// skip header, size is synchsafe int including len bytes
			extendedHeaderSize := uint(get7BitChunkedInt(b)) - 4
			_, err = readBytes(r, extendedHeaderSize)
			if err != nil {
				return nil, 0, fmt.Errorf("expected to read %d bytes (ID3v24 skip extended header): %v", extendedHeaderSize, err)
			}
			offset += extendedHeaderSize
		default:
			// nop, only 2.3 and 2.4 should have extended header
		}
	}

	return h, offset, nil
}

// ----- Utils -----

const readBytesMaxUpfront = 10 << 20 // 10MB
func readBytes(r io.Reader, n uint) ([]byte, error) {
	if n > readBytesMaxUpfront {
		b := &bytes.Buffer{}
		if _, err := io.CopyN(b, r, int64(n)); err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	}

	b := make([]byte, n)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func getBit(b byte, n uint) bool {
	x := byte(1 << n)
	return (b & x) == x
}

func get7BitChunkedInt(b []byte) int {
	var n int
	for _, x := range b {
		n = n << 7
		n |= int(x)
	}
	return n
}

func getInt(b []byte) int {
	var n int
	for _, x := range b {
		n = n << 8
		n |= int(x)
	}
	return n
}
