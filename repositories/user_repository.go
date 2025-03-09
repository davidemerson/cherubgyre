package repositories

import (
	"cherubgyre/dtos"
	"encoding/json"
	"errors"
	"log"
	"os"
)

func SaveUser(registerDTO dtos.RegisterDTO) error {
	file, err := os.OpenFile("users.json", os.O_RDWR, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var users []dtos.RegisterDTO
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		log.Printf("Error decoding user data: %v", err)
		return err
	}

	// Check if username already exists
	for _, user := range users {
		if user.Username == registerDTO.Username {
			return errors.New("user already exists")
		}
	}

	users = append(users, registerDTO)

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(users); err != nil {
		log.Printf("Error encoding user data: %v", err)
		return err
	}

	log.Printf("User data written to file: %+v", registerDTO)
	return nil
}

func GetUserByID(username string) (dtos.RegisterDTO, error) {
	file, err := os.Open("users.json")
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return dtos.RegisterDTO{}, err
	}
	defer file.Close()

	var users []dtos.RegisterDTO
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		log.Printf("Error decoding user data: %v", err)
		return dtos.RegisterDTO{}, err
	}

	for _, user := range users {
		if user.Username == username {
			return user, nil
		}
	}

	return dtos.RegisterDTO{}, errors.New("user not found")
}

func ValidateUserCredentials(username, pin string) (bool, error) {
	user, err := GetUserByID(username)
	if err != nil {
		return false, err
	}

	if user.NormalPin != pin {
		return false, errors.New("invalid credentials")
	}

	return true, nil
}
