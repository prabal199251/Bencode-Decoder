package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type FileInfo struct {
	PieceLength int64
	Pieces      [][]byte
	Length      int64
	Name        string
}

type Torrent struct {
	Info     FileInfo
	Announce string
}

func BDecode(reader *bufio.Reader) (interface{}, error) {
	ch, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch ch {

	case 'i':
		var buffer []byte
		for {
			ch, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			if ch == 'e' {
				value, err := strconv.ParseInt(string(buffer), 10, 64)
				if err != nil {
					panic(err)
				}

				return value, nil
			}
			buffer = append(buffer, ch)
		}

	case 'l':
		var listHolder []interface{}

		for {
			ch, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			if ch == 'e' {
				return listHolder, nil
			}

			reader.UnreadByte()
			data, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			listHolder = append(listHolder, data)
		}

	case 'd':
		dictHolder := map[string]interface{}{}

		for {
			ch, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			if ch == 'e' {
				return dictHolder, nil
			}

			reader.UnreadByte()
			data, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			key, ok := data.(string)
			if !ok {
				return nil, errors.New("key of the dictionary is not a string")
			}

			value, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			dictHolder[key] = value
		}

	default:
		reader.UnreadByte()

		var lengthBuf []byte
		for {
			ch, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			if ch == ':' {
				break
			}
			lengthBuf = append(lengthBuf, ch)
		}

		length, err := strconv.Atoi(string(lengthBuf))
		if err != nil {
			return nil, err
		}

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

func ParseTorrent(reader *bufio.Reader) Torrent {
	data, err := BDecode(reader)
	if err != nil {
		panic(err)
	}

	tData, ok := data.(map[string]interface{})
	if !ok {
		panic("Invalid torrent file!")
	}

	tInfoData, ok := tData["info"].(map[string]interface{})
	if !ok {
		panic("Invalid torrent file!")
	}

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
	fp, err := os.Open("ubuntu-24.04-desktop-amd64.iso.torrent")
	if err != nil {
		panic(err)
	}

	defer fp.Close()
	breader := bufio.NewReader(fp)
	torrentData := ParseTorrent(breader)
	fmt.Println(torrentData.Info.Length)
	fmt.Println(torrentData.Info.PieceLength)
	fmt.Println(torrentData.Info.Length / torrentData.Info.PieceLength)
	fmt.Println(torrentData.Info.Pieces)
	fmt.Println(torrentData.Info.Name)
	fmt.Println(torrentData.Announce)
}
