package fsdatabase

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

var filepath string = "./records.json"

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type Chirps struct {
	Chirps map[string]Chirp `json:"chirps"`
}

func ReadData(fpath string) (map[string]interface{}, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func IsFileEmpty(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	log.Println(fileInfo.Size())
	// File is empty if size is 0
	return fileInfo.Size() > 30, nil
}

func WriteData(fpath string, newchirp Chirp) ([]byte, error) {
	file, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// If the file doesn't exist, create it
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	//fmt.Println(os.Getwd())
	status, _ := IsFileEmpty(fpath)
	log.Println(status)
	var chirpsData Chirps
	if status {
		b, err := os.ReadFile(fpath)
		if err != nil {
			fmt.Print(err)
			fmt.Println(os.Getwd())
			return nil, err
		}

		err = json.Unmarshal(b, &chirpsData)
		// chirpsData has all json file content
		if err != nil {
			fmt.Printf("this second section %v", err)
			return nil, err
		}
		chripLength := len(chirpsData.Chirps)
		newchirp.Id = chripLength + 1
		// adding new chirp to chirpsData
		chirpsData.Chirps[fmt.Sprintf(strconv.Itoa(newchirp.Id))] = newchirp
		b, err = json.Marshal(chirpsData)
		if err != nil {
			fmt.Printf("this third section")
			return nil, err
		}

		err = os.WriteFile(fpath, b, 0644)
		jsondatabyte, err := json.Marshal(newchirp)
		fmt.Printf("this is jsondatabyte %s\n", string(jsondatabyte))
		log.Println("from IF")
		return jsondatabyte, nil
	} else {
		newchirp.Id = 1
		log.Println(newchirp)
		chirpsData := Chirps{
			Chirps: make(map[string]Chirp),
		}
		chirpsData.Chirps["1"] = newchirp
		log.Println(chirpsData)
		b, err := json.Marshal(chirpsData)
		if err != nil {
			fmt.Printf("this third section")
			return nil, err
		}

		err = os.WriteFile(fpath, b, 0644)
		jsondatabyte, err := json.Marshal(newchirp)
		fmt.Printf("this is jsondatabyte %s\n", string(jsondatabyte))
		log.Println("from ELSE")

		return jsondatabyte, nil
	}

}
