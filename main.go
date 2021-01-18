package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

type ArtworkResponse struct {
	Categories        []Category `json:"categories"`
	TotalCombinations int        `json:"total_combinations"`
}

type Category struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Images []Image `json:"images"`
}

type Image struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Href          string `json:"href"`
	ThumbnailHref string `json:"thumbnail_href"`
}

func generateRequired(opt []Image) *Image {
	return &opt[rand.Intn(len(opt))]
}

func generateOptional(opt []Image) *Image {
	r := rand.Intn(len(opt) + 1)
	if r == len(opt) {
		return nil
	}
	return &opt[r]
}

func generateGopher(opt map[string][]Image) string {
	var chars []string
	for k, v := range opt {
		if k == "Body" || k == "Eyes" {
			feature := generateRequired(v)
			chars = append(chars, feature.ID)
		} else if feature := generateOptional(v); feature != nil {
			chars = append(chars, feature.ID)
		}
	}

	req, err := http.NewRequest("GET", "https://gopherize.me/save", nil)
	if err != nil {
		log.Fatal(err)
	}
	q := req.URL.Query()

	q.Set("images", strings.Join(chars, "|"))
	req.URL.RawQuery = q.Encode()
	c := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	str := string(body)
	str = strings.TrimPrefix(str, "<a href=\"/gopher")
	str = strings.TrimSuffix(str, "\">Permanent Redirect</a>.\n")
	return str
}

func downloadGopher(gopherID string) {
	url := "https://storage.googleapis.com/gopherizeme.appspot.com/gophers" + gopherID + ".png"
	// don't worry about errors
	response, e := http.Get(url)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()

	//open a file for writing
	file, err := os.Create("gopher.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success!")

}
func main() {
	resp, err := http.Get("https://gopherize.me/api/artwork/")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	options := make(map[string][]Image)
	var awr *ArtworkResponse
	if err := json.Unmarshal(body, &awr); err != nil {
		log.Fatal(err)
	}
	for _, c := range awr.Categories {
		options[c.Name] = c.Images
	}
	id := generateGopher(options)
	downloadGopher(id)
}
