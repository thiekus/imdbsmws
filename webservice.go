package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"math"
	"net/http"
	"strconv"
)

type WsStatus struct {
	Status string `json:"status"`
}

type WsStatusGetMovie struct {
	WsStatus
	Data ImdbTitleEntry `json:"data"`
}

type WsStatusGetMovies struct {
	WsStatus
	Count      int `json:"count"`
	TotalCount int `json:"totalCount"`
	Page       int `json:"page"`
	MaxResult  int `json:"maxResult"`
	MaxPage    int `json:"maxPage"`
	Data       []ImdbTitleEntry `json:"data"`
}

func parseTitleFormValues(r *http.Request) (ImdbTitleEntry, error) {
	err:= r.ParseForm()
	if err != nil {
		return ImdbTitleEntry{}, err
	}
	id:= r.FormValue("id")
	ptype:= r.FormValue("type")
	title:= r.FormValue("title")
	originalTitle:= r.FormValue("originalTitle")
	genres:= r.FormValue("genres")
	syear:= r.FormValue("year")
	year, err:= strconv.Atoi(syear)
	if err != nil {
		year = 0
	}
	releaseDate:= r.FormValue("releaseDate")
	sruntimeMinutes:= r.FormValue("runtimeMinutes")
	runtimeMinutes, err:= strconv.Atoi(sruntimeMinutes)
	if err != nil {
		runtimeMinutes = 0
	}
	sisAdult:= r.FormValue("isAdult")
	isAdult:= sisAdult == "true"
	srating:= r.FormValue("rating")
	rating, err:= strconv.ParseFloat(srating, 64)
	if err != nil {
		rating = 0
	}
	description:= r.FormValue("description")
	imageUrl:= r.FormValue("imageUrl")
	entry:= ImdbTitleEntry{
		Id:id,
		Type:ptype,
		Title:title,
		OriginalTitle:originalTitle,
		Genres:genres,
		Year:year,
		ReleaseDate:releaseDate,
		RuntimeMinutes:runtimeMinutes,
		IsAdult:isAdult,
		Rating:rating,
		Description:description,
		ImageUrl:imageUrl,
	}
	return entry, nil
}

func moviesGetEndpoint(w http.ResponseWriter, r *http.Request) {
	log:= newLog()
	vars:= mux.Vars(r)
	id:= vars["id"]
	if id != "" {
		log.Printf("%s getting title id=%s", r.RemoteAddr, id)
		db, err:= OpenDefaultDatabase()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer db.Close()
		data, err:= db.GetTitleById(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		dataStatus:= WsStatusGetMovie{
			Data: data,
		}
		dataStatus.Status = "success"
		jsonData, err:= json.Marshal(dataStatus)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(jsonData)
	} else {
		log.Printf("%s getting movie list", r.RemoteAddr)
		err:= r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		title:= r.Form.Get("title")
		maxResult, err:= strconv.Atoi(r.Form.Get("maxResult"))
		if err != nil {
			maxResult = 20
		}
		page, err:= strconv.Atoi(r.Form.Get("page"))
		if err != nil {
			page = 1
		}
		if page <= 0 {
			page = 1
		}
		sortBy:= r.Form.Get("sortBy")
		filter:= ImdbTitleSearchFilter{
			Title:title,
			MaxResult:maxResult,
			Page:page,
			SortBy:sortBy,
			Ascending:true,
		}
		db, err:= OpenDefaultDatabase()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer db.Close()
		totalCount:= db.GetTitlesCount(title)
		maxPage:= int(math.Ceil(float64(totalCount) / float64(maxResult)))
		// Get titles
		result, err:= db.GetTitles(filter)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		resultStatus:= WsStatusGetMovies{
			Count:len(result),
			TotalCount:totalCount,
			Page:page,
			MaxResult:maxResult,
			MaxPage:maxPage,
			Data:result,
		}
		resultStatus.Status = "success"
		jsonData, err:= json.Marshal(resultStatus)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(jsonData)
	}
}

func moviesPostEndpoint(w http.ResponseWriter, r *http.Request) {
	log:= newLog()
	val, err:= parseTitleFormValues(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("%s posting table id=%s", r.RemoteAddr, val.Id)
	db, err:= OpenDefaultDatabase()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer db.Close()
	err = db.InsertTitleOnce(val)
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	status:= WsStatus{Status:"success"}
	data, err:= json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(data)
}

func moviesPutEndpoint(w http.ResponseWriter, r *http.Request) {
	log:= newLog()
	val, err:= parseTitleFormValues(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("%s updating table id=%s", r.RemoteAddr, val.Id)
	db, err:= OpenDefaultDatabase()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer db.Close()
	id:= val.Id
	curEntry, err:= db.GetTitleById(id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	if len(r.Form["type"]) == 0 {
		val.Type = curEntry.Type
	}
	if len(r.Form["title"]) == 0 {
		val.Title = curEntry.Title
	}
	if len(r.Form["originalTitle"]) == 0 {
		val.OriginalTitle = curEntry.OriginalTitle
	}
	if len(r.Form["genres"]) == 0 {
		val.Genres = curEntry.Genres
	}
	if len(r.Form["year"]) == 0 {
		val.Year = curEntry.Year
	}
	if len(r.Form["releaseDate"]) == 0 {
		val.ReleaseDate = curEntry.ReleaseDate
	}
	if len(r.Form["runtimeMinutes"]) == 0 {
		val.RuntimeMinutes = curEntry.RuntimeMinutes
	}
	if len(r.Form["isAdult"]) == 0 {
		val.IsAdult = curEntry.IsAdult
	}
	if len(r.Form["rating"]) == 0 {
		val.Rating = curEntry.Rating
	}
	if len(r.Form["description"]) == 0 {
		val.Description = curEntry.Description
	}
	if len(r.Form["imageUrl"]) == 0 {
		val.ImageUrl = curEntry.ImageUrl
	}
	// Now update
	err = db.UpdateTitle(val)
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	status:= WsStatus{Status:"success"}
	data, err:= json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(data)
}

func moviesDeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	log:= newLog()
	vars:= mux.Vars(r)
	id:= vars["id"]
	if id != "" {
		log.Printf("%s deleting table id=%s", r.RemoteAddr, id)
		db, err:= OpenDefaultDatabase()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer db.Close()
		err = db.DeleteEntry(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		dataStatus:= WsStatus{Status:"success"}
		jsonData, err:= json.Marshal(dataStatus)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(jsonData)
	}
}
