package fsdatabase

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type Chirps struct {
	Chirps map[string]Chirp `json:"chirps"`
}

func ReadChirpData(fpath string, id string) ([]byte, error) {
	chirpsData := Chirps{
		Chirps: make(map[string]Chirp),
	}
	file, err := os.Open(fpath)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	log.Print(string(b))
	err = json.Unmarshal(b, &chirpsData)
	if err != nil {
		return nil, err
	}
	// Extract values from the map and convert them into an array
	var chirpsArray []Chirp
	log.Println(chirpsData)
	// if url param has id of chirp return just particular id
	// get api/chirps/{id} START
	log.Println("The one before id", id)
	if id != "" {
		for k, v := range chirpsData.Chirps {
			if id == k { //compares the outer id with id url param ; both are string

				resultbyte, err := json.Marshal(Chirp{Id: v.Id, Body: v.Body})
				if err != nil {
					log.Println(err)
					return nil, err
				}
				log.Printf("get api/chirps/{%s} - %v", id, string(resultbyte))
				return resultbyte, nil
			}
		}

		return nil, errors.New(fmt.Sprintf("Record not found for id: %s", id))
	}
	// get api/chirps/{id} end
	for _, v := range chirpsData.Chirps {
		chirpsArray = append(chirpsArray, v)
	}

	// Print the transformed data
	transformedData, err := json.Marshal(chirpsArray)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("transformed data:--")
	fmt.Println(string(transformedData))

	return transformedData, nil
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

func WriteChirpData(fpath string, newchirp Chirp) ([]byte, error) {
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
