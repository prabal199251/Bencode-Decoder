package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// FileInfo represents the information about the file described in the torrent.
type FileInfo struct {
	PieceLength int64
	Pieces      [][]byte
	Length      int64
	Name        string
}

// Torrent represents the structure of a torrent file.
type Torrent struct {
	Info     FileInfo
	Announce string
}

// BDecode is a function to decode bencoded data from a reader.
func BDecode(reader *bufio.Reader) (interface{}, error) {
	// Read the next byte to determine the type of the bencoded value.
	ch, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch ch {

	// Integer case.
	case 'i':
		var buffer []byte
		for {
			ch, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			// 'e' indicates the end of an integer.
			if ch == 'e' {
				value, err := strconv.ParseInt(string(buffer), 10, 64)
				if err != nil {
					panic(err)
				}

				return value, nil
			}
			buffer = append(buffer, ch)
		}

	// List case.
	case 'l':
		var listHolder []interface{}

		for {
			ch, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			// 'e' indicates the end of a list.
			if ch == 'e' {
				return listHolder, nil
			}

			// Unread the byte to process it again.
			reader.UnreadByte()
			data, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			listHolder = append(listHolder, data)
		}

	// Dictionary case.
	case 'd':
		dictHolder := map[string]interface{}{}

		for {
			ch, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			// 'e' indicates the end of a dictionary.
			if ch == 'e' {
				return dictHolder, nil
			}

			// Unread the byte to process it again.
			reader.UnreadByte()
			data, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			// The key must be a string.
			key, ok := data.(string)
			if !ok {
				return nil, errors.New("key of the dictionary is not a string")
			}

			// Read the corresponding value for the key.
			value, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			dictHolder[key] = value
		}

	// String case.
	default:
		// Unread the byte to process the length.
		reader.UnreadByte()

		var lengthBuf []byte
		for {
			ch, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			// ':' indicates the end of the length prefix.
			if ch == ':' {
				break
			}
			lengthBuf = append(lengthBuf, ch)
		}

		// Convert the length to an integer.
		length, err := strconv.Atoi(string(lengthBuf))
		if err != nil {
			return nil, err
		}

		// Read the actual string content.
		var strBuf []byte

		for i := 0; i < length; i++ {
			ch, err := reader.ReadByte()
			if err != nil {
				panic(err)
			}

			strBuf = append(strBuf, ch)
		}

		return string(strBuf), nil
	}
}

// batch splits the data into chunks of the specified size.
func batch(data []byte, batch int) [][]byte {
	var result [][]byte

	for i := 0; i < len(data); i++ {
		end := i + batch
		if end > len(data) {
			end = len(data)
		}
		result = append(result, data[i:end])
	}
	return result
}

// ParseTorrent parses a torrent file and returns the Torrent structure.
func ParseTorrent(reader *bufio.Reader) Torrent {
	// Decode the bencoded data from the reader.
	data, err := BDecode(reader)
	if err != nil {
		panic(err)
	}

	// Ensure the top-level data is a dictionary.
	tData, ok := data.(map[string]interface{})
	if !ok {
		panic("Invalid torrent file!")
	}

	// Ensure the 'info' field is present and is a dictionary.
	tInfoData, ok := tData["info"].(map[string]interface{})
	if !ok {
		panic("Invalid torrent file!")
	}

	// Create and populate the Torrent structure.
	var torrentData Torrent

	torrentData.Announce = tData["announce"].(string)
	torrentData.Info = FileInfo{
		PieceLength: tInfoData["piece length"].(int64),
		Length:      tInfoData["length"].(int64),
		Name:        tInfoData["name"].(string),
		Pieces:      batch([]byte(tInfoData["pieces"].(string)), 20),
	}

	return torrentData
}

func main() {
	// Open the torrent file.
	fp, err := os.Open("ubuntu-24.04-desktop-amd64.iso.torrent")
	if err != nil {
		panic(err)
	}

	defer fp.Close()

	// Create a buffered reader for the file.
	breader := bufio.NewReader(fp)

	// Parse the torrent file.
	torrentData := ParseTorrent(breader)

	// Print the parsed torrent data.
	fmt.Println(torrentData.Info.Length)
	fmt.Println(torrentData.Info.PieceLength)
	fmt.Println(torrentData.Info.Length / torrentData.Info.PieceLength)
	fmt.Println(torrentData.Info.Pieces)
	fmt.Println(torrentData.Info.Name)
	fmt.Println(torrentData.Announce)
}
