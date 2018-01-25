package request

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
)

func newMultipartBody(a *Args, vs url.Values) (body io.Reader, contentType string, err error) {
	files := a.Files
	bodyBuffer := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuffer)
	for _, file := range files {
		fileWriter, err := bodyWriter.CreateFormFile(file.FieldName, file.FileName)
		if err != nil {
			return nil, "", err
		}
		_, err = io.Copy(fileWriter, file.File)
		if err != nil {
			return nil, "", err
		}
	}
	if a.Data != nil {
		for k, v := range a.Data {
			bodyWriter.WriteField(k, v)
		}
	}
	if vs != nil {
		for k, arr := range vs {
			for n := range arr {
				bodyWriter.WriteField(k, arr[n])
			}
		}
	}
	contentType = bodyWriter.FormDataContentType()
	defer bodyWriter.Close()
	body = bodyBuffer
	return
}

func newJSONBody(a *Args) (body io.Reader, contentType string, err error) {
	b, err := json.Marshal(a.Json)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(b), DefaultJsonType, err
}

// data can be map[string]string or map[string][]string
func newFormBody(a *Args, data interface{}) (body io.Reader, contentType string, err error) {
	vs := url.Values{}
	switch data.(type) {
	case map[string]string:
		for k, v := range data.(map[string]string) {
			vs.Set(k, v)
		}
	case map[string][]string:
		for k, arr := range data.(map[string][]string) {
			for n := range arr {
				vs.Add(k, arr[n])
			}
		}
	}
	if a.Files != nil {
		return newMultipartBody(a, vs)
	}
	return strings.NewReader(vs.Encode()), DefaultContentType, nil
}
