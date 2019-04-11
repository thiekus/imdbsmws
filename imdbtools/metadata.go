package imdbtools

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

type ImdbMetadata struct {
	Context         string               `json:"@context"`
	Type            string               `json:"@type"`
	Url             string               `json:"url"`
	Name            string               `json:"name"`
	Image           string               `json:"image"`
	//Genre           []string             `json:"genre"`
	ContentRating   string               `json:"contentRating"`
	//Actor           []ImdbMetadataEntity `json:"actor"`
	//Director        ImdbMetadataEntity   `json:"director"`
	//Creator         []ImdbMetadataEntity `json:"creator"`
	Description     string               `json:"description"`
	DatePublished   string               `json:"datePublished"`
	Keywords        string               `json:"keywords"`
	AggregateRating ImdbAggregateRating  `json:"aggregateRating"`
	// skipped: review
	// skipped: duration
	// skipped: trailer
}

type ImdbMetadataEntity struct {
	Type string	`json:"@type"`
	Url  string	`json:"url"`
	Name string	`json:"name"`
}

type ImdbAggregateRating struct {
	Type        string `json:"@type"`
	RatingCount int    `json:"ratingCount"`
	BestRating  string `json:"bestRating"`
	WorstRating string `json:"worstRating"`
	RatingValue string `json:"ratingValue"`
}

func GetImdbMetadata(titleId string) (ImdbMetadata, error) {
	meta:= ImdbMetadata{}
	url:= fmt.Sprintf("https://www.imdb.com/title/%s/", titleId)
	resp, err:= http.Get(url)
	if err != nil {
		return meta, err
	}
	defer resp.Body.Close()
	doc, err:= goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return meta, err
	}
	script:= doc.Find("script[type=\"application/ld+json\"]")
	if script.Length() > 0 {
		sel:= script.First()
		val:= sel.Text()
		err = UnmarshalImdbMetadata([]byte(val), &meta)
	} else {
		err = errors.New("metadata not found")
	}
	return meta, err
}

func UnmarshalImdbMetadata(data []byte, meta *ImdbMetadata) error {
	if err:= json.Unmarshal(data, meta); err == nil {
		return nil
	} else {
		return err
	}
}
