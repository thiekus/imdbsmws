package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/thiekus/imdbsmws/imdbtools"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ImportStatus struct {
	Running    bool   `json:"running"`
	Success    bool   `json:"success"`
	Progress   int    `json:"progress"`
	StatusText string `json:"statusText"`
}

type ImportFilter struct {
	Type         string
	Genres       string
	FromYear     int
	ToYear       int
	ExcludeAdult bool
}

const importBasicsTitleCachePath = "./caches/title.basics.tsv.gz"

var mainImportStatus *ImportStatus = nil

func importDatabase(basicUrl string, filter ImportFilter, useCache bool, saveCache bool) {
	importStatus := ImportStatus{
		Running:    true,
		Success:    false,
		Progress:   0,
		StatusText: "Initializing...",
	}
	mainImportStatus = &importStatus
	go func(basicUrl string, filter ImportFilter, useCache bool, saveCache bool) {
		beginTime := time.Now()
		sectionName := "importer"
		log := newLog()
		log.WithField("section", sectionName)
		defer func() { mainImportStatus.Running = false }()
		log.Printf("Downloading basic data from %s", basicUrl)
		tmpDl, err := ioutil.TempFile("", "ThkIMDbDataset")
		if err != nil {
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		tmpDlName := tmpDl.Name()
		log.Printf("Saving downloaded data to temporary file %s", tmpDlName)
		defer os.Remove(tmpDlName)
		defer tmpDl.Close()
		//
		var basicReader io.Reader = nil
		basicLength := int64(0)
		if !useCache || !isFileExists(importBasicsTitleCachePath) {
			resp, err := http.Get(basicUrl)
			if err != nil {
				mainImportStatus.StatusText = err.Error()
				log.Error(err)
				return
			}
			defer resp.Body.Close()
			basicReader = resp.Body
			basicLength = resp.ContentLength
		} else {
			cachedBasic, err := os.Open(importBasicsTitleCachePath)
			if err != nil {
				mainImportStatus.StatusText = err.Error()
				log.Error(err)
				return
			}
			defer cachedBasic.Close()
			basicReader = cachedBasic
			stat, _ := cachedBasic.Stat()
			basicLength = stat.Size()
		}
		// Download progress
		dlProgressChan := make(chan bool)
		go func(targetFile string, downloadSize int64) {
			for {
				if fileStat, err := os.Stat(targetFile); err == nil {
					curSize := fileStat.Size()
					percent := float64(float64(curSize)/float64(downloadSize)) * 33
					mainImportStatus.StatusText = fmt.Sprintf("Downloading basic dataset (%s of %s)",
						humanize.Bytes(uint64(curSize)), humanize.Bytes(uint64(downloadSize)))
					mainImportStatus.Progress = int(math.Round(percent))
					if curSize == downloadSize {
						break
					}
				} else {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
			dlProgressChan <- true
		}(tmpDlName, basicLength)
		if _, err = io.Copy(tmpDl, basicReader); err != nil {
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		<-dlProgressChan
		if saveCache {
			if basicCopy, err := os.Create(importBasicsTitleCachePath); err == nil {
				defer basicCopy.Close()
				tmpDl.Seek(0, io.SeekStart)
				io.Copy(basicCopy, tmpDl)
			}
		}
		log.Printf("%s basic data downloaded", humanize.Bytes(uint64(basicLength)))
		// Now, extract downloaded dataset from gzip stream, prepare tsv temp file
		tmpTsv, err := ioutil.TempFile("", "ThkIMDbDataset")
		if err != nil {
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		tmpTsvName := tmpTsv.Name()
		defer os.Remove(tmpTsvName)
		defer tmpTsv.Close()
		// Reset tmpDl position to first before perform unzip
		tmpDl.Seek(0, io.SeekStart)
		gunzip, err := gzip.NewReader(tmpDl)
		if err != nil {
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		defer gunzip.Close()
		log.Printf("GUnzipping data to %s", tmpTsvName)
		gunzipDone := false
		gunzipProgressChan := make(chan bool)
		go func(targetFile string) {
			for {
				if fileStat, err := os.Stat(targetFile); err == nil {
					mainImportStatus.StatusText = fmt.Sprintf("Unziping dataset (%s)",
						humanize.Bytes(uint64(fileStat.Size())))
					if gunzipDone {
						break
					}
				} else {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
			gunzipProgressChan <- true
		}(tmpTsvName)
		if _, err = io.Copy(tmpTsv, gunzip); err != nil {
			gunzipDone = true
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		gunzipDone = true
		<-gunzipProgressChan
		//
		tsvStat, err := os.Stat(tmpTsvName)
		if err != nil {
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		tsvSize := tsvStat.Size()
		log.Printf("%s GUnziped to tsv file", humanize.Bytes(uint64(tsvSize)))
		if _, err = tmpTsv.Seek(0, io.SeekStart); err != nil {
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		// Now, parse tsv dataset
		var entries []ImdbTitleEntry = nil
		readCount := int64(0)
		inclCount := int64(0)
		parseDone := false
		parseProgressChan := make(chan bool)
		tsvParser := imdbtools.NewImdbTsvParser(tmpTsv)
		go func() {
			for {
				progress := float64(float64(tsvParser.GetReadBytesSize())/float64(tsvSize))*33 + 34
				mainImportStatus.StatusText = fmt.Sprintf("Reading entries (including %s from %s)",
					humanize.Comma(inclCount), humanize.Comma(readCount))
				mainImportStatus.Progress = int(math.Round(progress))
				if parseDone {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
			parseProgressChan <- true
		}()
		expType := strings.Split(filter.Type, ",")
		expGenres := strings.Split(filter.Genres, ",")
		log.Print("Parsing tsv file...")
		for {
			rawEntry, err := tsvParser.ReadBasicsTitleEntry()
			if err == nil {
				readCount++
				// Is year is newer or same as desired?
				if (filter.FromYear != 0) && (rawEntry.StartYear < filter.FromYear) {
					continue
				}
				// Is year is older or same as desired?
				if (filter.ToYear != 0) && (rawEntry.StartYear > filter.ToYear) {
					continue
				}
				// Exclude adults?
				if filter.ExcludeAdult && rawEntry.IsAdult {
					continue
				}
				// Is type match?
				if filter.Type != "" {
					match := false
					entryType := strings.ToLower(rawEntry.Type)
					for _, val := range expType {
						if entryType == val {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}
				// Is genre match?
				if filter.Genres != "" {
					match := false
					entryGenres := strings.Split(strings.ToLower(rawEntry.Genres), ",")
					for _, val := range expGenres {
						for _, genres := range entryGenres {
							if val == genres {
								match = true
								break
							}
						}
						if match {
							break
						}
					}
					if !match {
						continue
					}
				}
				entry := ImdbTitleEntry{
					Id:             rawEntry.Id,
					LastUpdate:     time.Now().String(),
					Type:           rawEntry.Type,
					Title:          rawEntry.PrimaryTitle,
					OriginalTitle:  rawEntry.OriginalTitle,
					Genres:         rawEntry.Genres,
					Year:           rawEntry.StartYear,
					RuntimeMinutes: rawEntry.RuntimeMinutes,
					IsAdult:        rawEntry.IsAdult,
				}
				entries = append(entries, entry)
				inclCount++
			} else {
				break
			}
		}
		parseDone = true
		<-parseProgressChan
		log.Printf("Parse done, %s from %s entries included",
			humanize.Comma(inclCount), humanize.Comma(readCount))
		//
		db, err := CreateDefaultDatabase(true)
		if err != nil {
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		defer db.Close()
		// Limit connection instance to one. It will make insertion slower, but greatly reduce database locking problem
		db.Db().SetMaxOpenConns(1)
		insertProcessCount := 0
		insertFailedCount := 0
		insertDone := false
		insertProgressChan := make(chan bool)
		go func() {
			for {
				progress := float64(float64(insertProcessCount)/float64(inclCount))*33 + 67
				mainImportStatus.StatusText = fmt.Sprintf("Crawling metadata and inserting entries (%s of %s, %s failed)",
					humanize.Comma(int64(insertProcessCount)), humanize.Comma(inclCount), humanize.Comma(int64(insertFailedCount)))
				mainImportStatus.Progress = int(math.Round(progress))
				if insertDone {
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
			insertProgressChan <- true
		}()
		insertStmt, err := db.PrepareInsertTitle()
		if err != nil {
			mainImportStatus.StatusText = err.Error()
			log.Error(err)
			return
		}
		defer insertStmt.Close()
		throttleQueue := 0 // max to 8 simultaneous go routines
		throttleWg := sync.WaitGroup{}
		insertWg := sync.WaitGroup{}
		for _, entry := range entries {
			go func(entry ImdbTitleEntry) {
				beginInsertTime := time.Now()
				insertWg.Add(1)
				insertSucceeded := false
				defer func() {
					if !insertSucceeded {
						insertFailedCount++
					}
					throttleWg.Wait()
					throttleWg.Add(1)
					throttleQueue--
					throttleWg.Done()
					insertWg.Done()
				}()
				log.Printf("[%s] Processing for title \"%s\"", entry.Id, entry.Title)
				meta := imdbtools.ImdbMetadata{}
				cacheFile := fmt.Sprintf("./caches/%s.json", entry.Id)
				canSaveCache := false
				if useCache && isFileExists(cacheFile) {
					log.Printf("[%s] Using cached metadata", entry.Id)
					cacheFh, err := os.Open(cacheFile)
					if err != nil {
						log.Errorf("[%s] %s", entry.Id, err)
						return
					}
					defer cacheFh.Close()
					bytes, err := ioutil.ReadAll(cacheFh)
					if err != nil {
						log.Errorf("[%s] %s", entry.Id, err)
						return
					}
					err = imdbtools.UnmarshalImdbMetadata(bytes, &meta)
					if err != nil {
						log.Errorf("[%s] %s", entry.Id, err)
						return
					}
				} else {
					log.Printf("[%s] Dowloading metadata", entry.Id)
					var err error
					meta, err = imdbtools.GetImdbMetadata(entry.Id)
					if err != nil {
						log.Errorf("[%s] %s", entry.Id, err)
						return
					}
					canSaveCache = true
				}
				// Add missing values
				entry.ReleaseDate = meta.DatePublished
				rating, err := strconv.ParseFloat(meta.AggregateRating.RatingValue, 64)
				if err != nil {
					rating = 0
				}
				entry.Rating = rating
				entry.Description = meta.Description
				entry.ImageUrl = meta.Image
				err = db.InsertTitle(insertStmt, entry)
				if err != nil {
					log.Errorf("[%s] Insert error: %s", entry.Id, err)
					return
				}
				if canSaveCache && saveCache {
					log.Printf("[%s] Saving metadata...", entry.Id)
					if metaByte, err := json.Marshal(meta); err == nil {
						err := ioutil.WriteFile(cacheFile, metaByte, os.ModePerm)
						if err != nil {
							log.Error(err)
							return
						}
					} else {
						log.Error(err)
						return
					}
				}
				elapsedInsertTime := time.Since(beginInsertTime)
				log.Printf("[%s] Data inserted for %s", entry.Id, elapsedInsertTime)
				insertSucceeded = true
			}(entry)
			insertProcessCount++
			for throttleQueue >= 8 {
				time.Sleep(100 * time.Millisecond)
			}
			throttleWg.Wait()
			throttleWg.Add(1)
			throttleQueue++
			throttleWg.Done()
		}
		insertWg.Wait()
		insertDone = true
		<-insertProgressChan
		insertSuccessCount := int64(insertProcessCount - insertFailedCount)
		processResult := fmt.Sprintf("All done! %s entries successfuly included and %s errors since %s",
			humanize.Comma(insertSuccessCount), humanize.Comma(int64(insertFailedCount)), humanize.Time(beginTime))
		mainImportStatus.StatusText = processResult
		mainImportStatus.Success = true
		log.Print(processResult)
	}(basicUrl, filter, useCache, saveCache)
}

func importStatusEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	importStatus := ImportStatus{
		Running:    false,
		Success:    false,
		Progress:   0,
		StatusText: "Import not running!",
	}
	if mainImportStatus != nil {
		importStatus = *mainImportStatus
	}
	if jsonData, err := json.Marshal(importStatus); err == nil {
		w.Write(jsonData)
	}
}
