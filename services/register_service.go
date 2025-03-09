package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"errors"
	"log"
)

func RegisterUser(registerDTO dtos.RegisterDTO) (string, dtos.RegisterDTO, error) {
	if registerDTO.NormalPin == "" || registerDTO.DuressPin == "" || registerDTO.Username == "" {
		log.Println("Validation failed: normal_pin and duress_pin , and username are required")
		return "", dtos.RegisterDTO{}, errors.New("normal_pin, username & duress_pin are required")
	}

	err := repositories.SaveUser(registerDTO)
	if err != nil {
		log.Printf("Error saving user: %v", err)
		return "", dtos.RegisterDTO{}, err
	}

	log.Printf("User saved successfully: %+v", registerDTO)
	return "User registered successfully", registerDTO, nil
}
