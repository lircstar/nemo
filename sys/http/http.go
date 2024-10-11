package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/lircstar/nemo/sys/log"
)

func Post(url string, obj any, retObj any) bool {
	contentType := "application/json;charset=utf-8"
	bs, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("Json format error : %v", err.Error())
		return false
	}

	body := bytes.NewBuffer(bs)
	resp, err := http.Post(url, contentType, body)

	if err != nil {
		log.Errorf("Post failed : %v", err.Error())
		return false
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Read failed : %v", err.Error())
		return false
	}

	err = json.Unmarshal(content, retObj)
	if err != nil {
		log.Errorf("Json return format error : %v", err.Error())
		return false
	}
	return true
}

func PostJson(url string, data string, retJson *map[string]any) bool {
	contentType := "application/json;charset=utf-8"
	body := bytes.NewBuffer([]byte(data))
	resp, err := http.Post(url, contentType, body)

	if err != nil {
		log.Errorf("Post failed : %v", err.Error())
		return false
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Read failed : %v", err.Error())
		return false
	}

	err = json.Unmarshal(content, retJson)
	if err != nil {
		log.Errorf("Json return format error : %v", err.Error())
		return false
	}

	return true
}

func PostWithHeader(url string, headParams map[string]string, obj any, retObj any) bool {

	bs, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("Json format error : %v", err.Error())
		return false
	}
	body := bytes.NewBuffer(bs)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Errorf("Request creation failed : %v", err.Error())
		return false
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	for key, val := range headParams {
		req.Header.Set(key, val)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Post failed : %v", err.Error())
		return false
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Read failed: %v", err.Error())
		return false
	}

	err = json.Unmarshal(content, retObj)
	if err != nil {
		log.Errorf("Json return format error : %v", err.Error())
		return false
	}
	return true
}

func PostJsonWithHeader(url string, headParams map[string]string, data string, retJson *map[string]any) bool {
	body := bytes.NewBuffer([]byte(data))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Errorf("Request creation failed : %v", err.Error())
		return false
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	for key, val := range headParams {
		req.Header.Set(key, val)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Post failed : %v", err.Error())
		return false
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Read failed: %v", err.Error())
		return false
	}

	err = json.Unmarshal(content, retJson)
	if err != nil {
		log.Errorf("Json return format error : %v", err.Error())
		return false
	}
	return true
}
