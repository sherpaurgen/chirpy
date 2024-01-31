package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

var filepath string = "./records.json"

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type Chirps struct {
	Chirps map[string]Chirp `json:"chirps"`
}

func readData(fpath string) (map[string]interface{}, error) {
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

func writeData(fpath string, newchirp Chirp) error {
	b, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}
	var chirpsData Chirps
	err = json.Unmarshal(b, &chirpsData)
	// chirpsData has all json file content
	if err != nil {
		return err
	}
	// adding new chirp to chirpsData
	chirpsData.Chirps[fmt.Sprintf(strconv.Itoa(newchirp.ID))] = newchirp
	b, err = json.Marshal(chirpsData)
	if err != nil {
		return err
	}

	err = os.WriteFile(fpath, b, 0644)

	return nil

}
