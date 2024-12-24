package processor

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
)

type Response struct {
	Image    string                 `json:"url"`
	Hash     string                 `json:"hash"`
	Metadata map[string]interface{} `json:"metadata"`
	Error    string                 `json:"error"`
}

var client = http.DefaultClient

const (
	baseURL  = "http://localhost:8080"
	imgPath  = "/image"
	metaPath = "/meta"
)

func generator(n int) chan *Response {
	c := make(chan *Response)
	go func() {
		defer close(c)
		for i := 0; i < n; i++ {
			c <- &Response{Image: strconv.Itoa(i)}
		}
	}()

	return c
}

func getMetadata(imgs chan *Response) chan *Response {
	out := make(chan *Response)
	go func() {
		defer close(out)

		for img := range imgs {
			if img.Error != "" {
				out <- img
				continue
			}

			url := fmt.Sprintf("%s%s?id=%s", baseURL, metaPath, img.Image)
			req, err := client.Get(url)
			if err != nil {
				img.Error = err.Error()
				out <- img
				continue
			}
			defer req.Body.Close()

			data, err := io.ReadAll(req.Body)
			if err != nil {
				img.Error = err.Error()
				out <- img
				continue
			}

			err = json.Unmarshal(data, &img.Metadata)
			if err != nil {
				img.Error = err.Error()
				out <- img
				continue
			}

			out <- img
		}
	}()

	return out
}

func calculateHash(imgs chan *Response) chan *Response {
	out := make(chan *Response)
	go func() {
		defer close(out)

		for img := range imgs {
			if img.Error != "" {
				out <- img
				continue
			}

			req, err := client.Get(fmt.Sprintf("%s%s?id=%s", baseURL, imgPath, img.Image))
			if err != nil {
				img.Error = err.Error()
				out <- img
				continue
			}
			defer req.Body.Close()

			hasher := md5.New()
			_, err = io.Copy(hasher, req.Body)
			if err != nil {
				img.Error = err.Error()
				out <- img
				continue
			}

			img.Hash = fmt.Sprintf("%x", hasher.Sum(nil))
			out <- img
		}
	}()

	return out
}

func merge(channels []chan *Response) <-chan *Response {
	out := make(chan *Response)
	wg := sync.WaitGroup{}

	for _, ch := range channels {
		wg.Add(1)
		go func(imgs chan *Response) {
			defer wg.Done()

			for img := range imgs {
				out <- img
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func Start(n int) {
	imgs := generator(n)
	workers := 3
	workerCh := make([]chan *Response, workers)

	for i := 0; i < workers; i++ {
		ch := getMetadata(calculateHash(imgs))
		workerCh[i] = ch
	}

	imgsMerged := merge(workerCh)
	for img := range imgsMerged {
		fmt.Printf("%v\n", img)
	}
}
