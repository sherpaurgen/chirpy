package fsdatabase

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type Chirps struct {
	Chirps map[string]Chirp `json:"chirps"`
	Users  map[string]User  `json:"users"`
}

type User struct {
	Id                 int    `json:"id"`
	Email              string `json:"email"`
	Password           string `json:"password"`
	Expires_in_seconds int    `json:"-"`
}
type UserInfo struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func AuthenticateUser(user User, fpath string) ([]byte, error) {
	status, _ := IsFileEmpty(fpath)
	if status {
		file, _ := os.Open(fpath)
		defer file.Close()
		data, _ := io.ReadAll(file)
		var filecontent Chirps
		err := json.Unmarshal(data, &filecontent)
		if err != nil {
			return nil, err
		}
		for _, userobj := range filecontent.Users {
			if userobj.Email == user.Email {
				passwordMatch := checkSecret(userobj.Password, user.Password)
				if passwordMatch {
					authenticatedUser := UserInfo{Id: userobj.Id, Email: userobj.Email}
					jsondata, err := json.Marshal(authenticatedUser)
					return jsondata, err
				} else {
					return nil, fmt.Errorf("email/password invalid")
				}
			} else {
				continue
			}

		}
		return nil, fmt.Errorf("email/password invalid")
	}

	return nil, nil
}

func checkSecret(hashedSecret string, userInputSecret string) bool {
	log.Println(string(hashedSecret), userInputSecret)
	err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(userInputSecret))
	log.Println(bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(userInputSecret)))
	return err == nil
}

func getCurrentUserCount(fpath string) (current_user_count int, err error) {
	status, _ := IsFileEmpty(fpath)
	if status {
		file, _ := os.Open(fpath)
		defer file.Close()
		data, _ := io.ReadAll(file)
		var filecontent Chirps
		err := json.Unmarshal(data, &filecontent)
		if err != nil {
			return 0, err
		}

		current_user_count := len(filecontent.Users)
		log.Println("curr user count:", current_user_count)
		//if it doesnt exist current user count is 0
		return current_user_count, nil
	}
	return
}

func CreateUser(user User, fpath string) ([]byte, error) {
	//the user is struct & not a json
	var alldata Chirps
	var newUser User
	fmt.Printf("Argument in CreateUser :%v\n", user)
	count, err := getCurrentUserCount(fpath)
	if err != nil {
		fmt.Printf("getCurrentUserCount error %v\n", err)
		os.Exit(1)
	}
	if count < 1 {
		alldata.Users = make(map[string]User)
	}
	count = count + 1

	countStr := strconv.Itoa(count)
	newUser = user
	newUser.Id = count
	secretbyte := []byte(user.Password)
	cost := 12 // or any other appropriate cost value

	hashedPassword, err := bcrypt.GenerateFromPassword(secretbyte, cost)
	newUser.Password = string(hashedPassword)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	fmt.Println(os.Getwd())
	file, _ := os.OpenFile(fpath, os.O_RDWR, 0644)

	data, _ := io.ReadAll(file)
	defer file.Close()

	if len(data) == 0 {
		fmt.Println("JSON data is empty")
	}
	//convert json file content to struct
	err = json.Unmarshal(data, &alldata)
	if err != nil {
		fmt.Printf("Unmarshal error %v\n", err)
		fmt.Printf("error body: %v", newUser)
		os.Exit(1)
	}
	log.Println("all Data:")
	log.Println(alldata)
	alldata.Users[countStr] = newUser
	updatedData, err := json.Marshal(alldata)
	fmt.Println("updated Data below:-")
	fmt.Println(string(updatedData))
	if err != nil {
		log.Println("error in marshalling all data:", err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		log.Println("Fileseek error:", err)
		os.Exit(1)
	}
	err = file.Truncate(0)
	if err != nil {
		log.Println("FileTruncate error:", err)
		os.Exit(1)
	}
	_, err = file.Write(updatedData)
	if err != nil {
		log.Println("FileWrite error:", err)
		os.Exit(1)
	}
	//sending email and id only
	userinfo := UserInfo{Id: count, Email: newUser.Email}
	jsondata, _ := json.Marshal(userinfo)
	return jsondata, nil
}

func ModifyUser(fpath string, id string, userinfo User) ([]byte, error) {
	log.Printf("ID %v , Userdata%v\n", id, userinfo)
	fh, err := os.OpenFile(fpath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal("Error when opening file in ModifyUser func:", err)
	}
	defer fh.Close()
	var payload Chirps
	content, err := io.ReadAll(fh)
	if err != nil {
		log.Fatal("Error reading file in modifyUser func:", err)
	}
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal("Error during unmarshal in ModifyUser func:", err)
	}
	//updating the struct with user change request

	intid, _ := strconv.Atoi(id)
	for k, v := range payload.Users {
		if k == id {
			log.Printf("key %s , id %s", k, id)
			v.Email = userinfo.Email
			v.Id = intid
			secretbyte := []byte(userinfo.Password)
			cost := 12
			hashedPassword, _ := bcrypt.GenerateFromPassword(secretbyte, cost)
			v.Password = string(hashedPassword)
			payload.Users[k] = v
			break
		} else {
			continue
		}
	}
	log.Println(payload.Users)
	updatedData, err := json.Marshal(payload)
	if err != nil {
		log.Println("error in Marshal modifydata: ", err)
		log.Fatalln(err)
	}
	err = os.Truncate(fpath, 0)
	if err != nil {
		log.Println("error in Marshal modifydata: ", err)
		log.Fatalln(err)
	}
	_, err = fh.Seek(0, 0)
	if err != nil {
		log.Println("error in file truncate modifyuser func:", err)
		log.Fatalln(err)
	}
	_, err = fh.Write(updatedData)
	if err != nil {
		log.Fatalln("error in updating file modifyuser:", err)
	}
	res, err := json.Marshal(userinfo)
	log.Println(string(updatedData))
	return res, err
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
	b, _ := io.ReadAll(file)
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

		//return nil, errors.New(fmt.Sprintf("Record not found for id: %s", id))
		return nil, fmt.Errorf("record not found for id: %s", id)
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
	defer file.Close()
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
			return nil, err
		}
		chripLength := len(chirpsData.Chirps)
		newchirp.Id = chripLength + 1
		// adding new chirp to chirpsData
		idStr := strconv.Itoa(newchirp.Id)
		chirpsData.Chirps[idStr] = newchirp
		b, err = json.Marshal(chirpsData)
		if err != nil {
			return nil, err
		}

		_ = os.WriteFile(fpath, b, 0644)
		jsondatabyte, _ := json.Marshal(newchirp)

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

		_ = os.WriteFile(fpath, b, 0644)
		jsondatabyte, _ := json.Marshal(newchirp)

		return jsondatabyte, nil
	}

}
