package core

import (
	"errors"
	"bytes"
	"io"
	"reflect"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func Contain(obj interface{}, target interface{}) (bool, error) {
    targetValue := reflect.ValueOf(target)
    switch reflect.TypeOf(target).Kind() {
    case reflect.Slice, reflect.Array:
        for i := 0; i < targetValue.Len(); i++ {
            if targetValue.Index(i).Interface() == obj {
                return true, nil
            }
        }
    case reflect.Map:
        if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
            return true, nil
        }
    }

    return false, errors.New("not in array")
}


func ParseRequest(req *http.Request) (map[string]interface{}, error) {
	requestData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	var reqData interface{}
	err = json.Unmarshal(requestData, &reqData)
	if err != nil {
		return nil, err
	}
	reqMapData := reqData.(map[string]interface{})
	return reqMapData, nil
}

func ParseRequestForProxy(req *http.Request) (map[string]interface{}, error) {
	newBody := bytes.NewBuffer(make([]byte, 0))
	reader := io.TeeReader(req.Body, newBody)
	requestData,err:=ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var reqData interface{}
	err = json.Unmarshal(requestData, &reqData)
	if err != nil {
		return nil, err
	}
	reqMapData := reqData.(map[string]interface{})
	return reqMapData, nil
}