package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	// "database/sql"
	DocType1 "Kintine-with-Go/models/kintoneDocument"
)

const (
	// KINTONE_API_URL   = "https://kweth.cybozu.com/k/v1/records.json?app=169"
	KINTONE_API_URL   = "https://xxx.cybozu.com/k/v1/%s.json"
	KINTONE_API_TOKEN = "xxxx" //View only
)

func main() {
	// var wg sync.WaitGroup
	go Consume()

	waitingForSignal(os.Interrupt, syscall.SIGTERM)
	log.Println("The service is shutting down...")

	log.Println("terminated...")

	os.Exit(0)
}

func makeAPI(apiFormat string, command string) string {
	return fmt.Sprintf(apiFormat, command)
}

func UpdateSyned(appId int, recordId int) {
	// params := fmt.Sprintf(`{
	// 	"app":%v,
	// 	"records":[
	// 		{
	// 			"id":%v,
	// 			"record":{
	// 				"IsSync":{
	// 					"value":"1"
	// 				}
	// 			}
	// 		}
	// 	]
	// }`, appId, recordId)
	params := fmt.Sprintf(`{
		"app":%v,
		"id":%v,
		"record":{
			"IsSync":{
				"value":"1"
			}
		}
	}`, appId, recordId)
	_, err := KintoneReviseAPI(makeAPI(KINTONE_API_URL, "record"), params)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Consume() {

	for true {
		dataObj, err := GetRecords(169)
		if err != nil {
			fmt.Println(err.Error())
			log.Fatal(err.Error())
		}

		if len(dataObj.Records) > 0 {
			for index, record := range dataObj.Records {
				fmt.Printf("Document No.%v - Record_number : %v - IsSync : %v\n", index, record.Record_number.Value, record.IsSync.Value)
				recordId, err := strconv.Atoi(record.Record_number.Value)
				if err != nil {
					log.Fatal(err.Error())
					continue
				}
				UpdateSyned(169, recordId)
			}

		} else {
			fmt.Println("No data")
		}

		// fmt.Println("status : ", res.StatusCode)
		// fmt.Println("body : ", string(data))
		time.Sleep(time.Second * 5)
	}
}

func GetRecords(appId int) (DocType1.DataObj, error) {
	var errRes error
	reqUrl, _ := url.Parse(makeAPI(KINTONE_API_URL, "records"))
	params := fmt.Sprintf(`{
		"app":%v,
		"query":" IsSync != 1 "
	}`, appId)
	headers := map[string][]string{
		"Content-Type":       {"application/json; charset=UTF-8"},
		"X-Cybozu-API-Token": {KINTONE_API_TOKEN},
	}
	data, err := KintoneQueryAPI(reqUrl, headers, params)
	if err != nil {
		log.Fatal(err.Error())
		return DocType1.DataObj{}, err
	}

	dataObj := DocType1.DataObj{}

	json.Unmarshal([]byte(string(data)), &dataObj)
	errRes = nil
	return dataObj, errRes
}

func GetRecord(appId int, recordId int) (DocType1.Record, error) {

	var errRes error
	reqUrl, _ := url.Parse(makeAPI(KINTONE_API_URL, "record"))
	params := fmt.Sprintf(`{
		"app":%v,
		"query":"$id = %v"
	}`, appId, recordId)
	headers := map[string][]string{
		"Content-Type":       {"application/json; charset=UTF-8"},
		"X-Cybozu-API-Token": {KINTONE_API_TOKEN},
	}
	data, err := KintoneQueryAPI(reqUrl, headers, params)
	if err != nil {
		log.Fatal(err.Error())
		return DocType1.Record{}, err
	}
	dataObj := DocType1.DataObj{}

	json.Unmarshal([]byte(string(data)), &dataObj)
	errRes = nil

	return dataObj.Records[0], errRes
}

func KintoneReviseAPI(reqUrl string, params string) (bool, error) {
	reqBody := ioutil.NopCloser(strings.NewReader(params))
	// fmt.Println(reqBody)
	req, err := http.NewRequest("PUT", reqUrl, reqBody)
	if err != nil {
		log.Fatal(err.Error())
		return false, err
	}
	req.Header.Set("X-Cybozu-API-Token", KINTONE_API_TOKEN)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	fmt.Println("response Status:", res.Status)
	//fmt.Println("response Headers:", res.Header)
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println("response Body:", string(body))

	return true, nil
}

func KintoneQueryAPI(reqUrl *url.URL, headers map[string][]string, params string) ([]byte, error) {

	reqBody := ioutil.NopCloser(strings.NewReader(params))
	// fmt.Println(reqBody)
	req := &http.Request{
		Method: "GET",
		URL:    reqUrl,
		Header: headers,
		Body:   reqBody,
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte(""), err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte(""), err
	}
	res.Body.Close()
	// fmt.Println(res)
	return data, nil
}

func waitingForSignal(sig ...os.Signal) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, sig...)

	s := <-stop
	log.Println("Got signal ", s.String())
}
