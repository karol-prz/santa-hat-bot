package main

import (
	"bytes"
	"regexp"
	"net/http"

)

// GetDataFromURL parses webpage and returns byte array from response.body
func GetDataFromURL(url string) ([]byte, error) {
	// don't worry about errors
    response, e := http.Get(url)
    if e != nil {
		return []byte{}, e
    }
    defer response.Body.Close()

	buf := bytes.NewBuffer(make([]byte, 0, response.ContentLength))
	_, readErr := buf.ReadFrom(response.Body)
	return buf.Bytes(), readErr
}

// IsImageExt checks if string end with image extension
func IsImageExt(file string) bool{
	re := regexp.MustCompile(`.(jpg|png|jpeg)$`)
	return re.MatchString(file)
}

// GetImageExt returns image extension
func GetImageExt(file string) string {
	re := regexp.MustCompile(`.(jpg|png|jpeg)$`)
	return re.FindString(file)
}

// GetURL returns a url if its at end of string
func GetURL(content string) string {
	re := regexp.MustCompile(`http.*$`)
	return re.FindString(content)
}