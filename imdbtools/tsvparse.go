package imdbtools

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
)

type ImdbTsvParser struct {
	header  []string
	scanner *bufio.Scanner
	readSize int64
}

type ImdbBasicsTitleTsvEntry struct {
	Id             string // tconst
	Type           string // titleType
	PrimaryTitle   string // primaryTitle
	OriginalTitle  string // originalTitle
	IsAdult        bool   // isAdult
	StartYear      int    // startYear
	EndYear        int    // endYear
	RuntimeMinutes int    // runtimeMinutes
	Genres         string // genres
}

const (
	BasicTsvId             = "tconst"
	BasicTsvType           = "titleType"
	BasicTsvPrimaryTitle   = "primaryTitle"
	BasicTsvOriginalTitle  = "originalTitle"
	BasicTsvIsAdult        = "isAdult"
	BasicTsvStartYear      = "startYear"
	BasicTsvEndYear        = "endYear"
	BasicTsvRuntimeMinutes = "runtimeMinutes"
	BasicTsvGenres         = "genres"
)

func NewImdbTsvParser(reader io.Reader) ImdbTsvParser {
	tsv:= ImdbTsvParser{}
	tsv.scanner = bufio.NewScanner(reader)
	tsv.Reset()
	return tsv
}

func (i *ImdbTsvParser) explodeLine(line string) []string {
	return strings.Split(line, "\t")
}

func (i *ImdbTsvParser) ReadBasicsTitleEntry() (ImdbBasicsTitleTsvEntry, error) {
	if i.scanner.Scan() {
		data:= ImdbBasicsTitleTsvEntry{}
		line:= i.scanner.Text()
		i.readSize += int64(len(line))
		entries:= i.explodeLine(line)
		for index, val:= range entries {
			key:= i.header[index]
			if val == "\\N" {
				val = ""
			}
			switch key {
			case BasicTsvId:
				data.Id = val
			case BasicTsvType:
				data.Type = val
			case BasicTsvPrimaryTitle:
				data.PrimaryTitle = val
			case BasicTsvOriginalTitle:
				data.OriginalTitle = val
			case BasicTsvIsAdult:
				vi, _:= strconv.Atoi(val)
				data.IsAdult = vi != 0
			case BasicTsvStartYear:
				if val == "" {
					val = "0"
				}
				vi, _:= strconv.Atoi(val)
				data.StartYear = vi
			case BasicTsvEndYear:
				if val == "" {
					val = "0"
				}
				vi, _:= strconv.Atoi(val)
				data.EndYear = vi
			case BasicTsvRuntimeMinutes:
				if val == "" {
					val = "0"
				}
				vi, _:= strconv.Atoi(val)
				data.RuntimeMinutes = vi
			case BasicTsvGenres:
				data.Genres = val
			default:
				log.Printf("Unknown extra basic tsv value: %s", val)
			}
		}
		return data, nil
	} else {
		err:= i.scanner.Err()
		if err == nil {
			err = errors.New("EOF")
		}
		return ImdbBasicsTitleTsvEntry{}, err
	}
}

func (i *ImdbTsvParser) GetReadBytesSize() int64 {
	return i.readSize
}

func (i *ImdbTsvParser) Reset() bool {
	i.readSize = 0
	if i.scanner.Scan() {
		line:= i.scanner.Text()
		i.readSize += int64(len(line))
		entries:= i.explodeLine(line)
		i.header = entries
		return true
	} else {
		return false
	}
}
