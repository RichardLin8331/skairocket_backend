package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

/*
	type stockaday struct {
		Openv  string `json:"open"`
		Closev string `json:"close"`
	}

	type stockresp struct {
		Stockresult [][]string `json:"stock_result"`
	}
*/
type stock_basic struct {
	Stocknum   string `json:"stocknum"`
	Stockmonth string `json:"stockmonth"`
}

func main() {
	sb := stock_basic{Stocknum: "2330", Stockmonth: "3"}
	sb_json, _ := json.Marshal(sb)
	resp, err := http.Post("http://127.0.0.1:19982/SearchStock", "application/json", bytes.NewBuffer(sb_json))

	if err != nil {
		print(err.Error())
	}

	var data map[string]interface{}

	body, _ := io.ReadAll(resp.Body)

	err2 := json.Unmarshal(body, &data)
	if err2 != nil {
		print(err2.Error())
	}

	for _, v := range data {
		fmt.Println(v)
	}
	/*
		str_slice := make([]string, 0)
		for _, v := range data {
			str_tmp := v.(string)
			str_slice = append(str_slice, str_tmp)
		}

		data_tomodel, _ := json.Marshal(data)
		resp2, err3 := http.Post("http://127.0.0.1:18501/v1/models/AutoEncoderModel:predict", "application/json", bytes.NewBuffer(data_tomodel))

		if err3 != nil {
			print("err3", err3.Error())
		}

		var data2 map[string]interface{}

		body2, _ := io.ReadAll(resp2.Body)
		err4 := json.Unmarshal(body2, &data2)
		if err4 != nil {
			print("err4", err4.Error())
		}
		for k, v := range data2 {

			fmt.Println(k, v)
		}
	*/
}
