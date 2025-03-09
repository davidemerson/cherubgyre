package repositories

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type FollowerRelation struct {
	Follower string `json:"follower"`
	Followed string `json:"followed"`
}

func AddFollower(followerID, followedID string) error {
	log.Printf("Adding follower: %s to user: %s", followerID, followedID)
	file, err := os.OpenFile("followers.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil {
		log.Printf("Error decoding follower data: %v", err)
		return err
	}

	for _, relation := range relations {
		if relation.Follower == followerID && relation.Followed == followedID {
			log.Printf("User: %s is already following: %s", followerID, followedID)
			return errors.New("already following")
		}
	}

	relations = append(relations, FollowerRelation{Follower: followerID, Followed: followedID})

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(relations); err != nil {
		log.Printf("Error encoding follower data: %v", err)
		return err
	}

	log.Printf("Successfully added follower: %s to user: %s", followerID, followedID)
	return nil
}

func RemoveFollower(followerID, followedID string) error {
	log.Printf("Removing follower: %s from user: %s", followerID, followedID)
	file, err := os.OpenFile("followers.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil {
		log.Printf("Error decoding follower data: %v", err)
		return err
	}

	for i, relation := range relations {
		if relation.Follower == followerID && relation.Followed == followedID {
			relations = append(relations[:i], relations[i+1:]...)
			break
		}
	}

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(relations); err != nil {
		log.Printf("Error encoding follower data: %v", err)
		return err
	}

	log.Printf("Successfully removed follower: %s from user: %s", followerID, followedID)
	return nil
}

func GetFollowers(userID string) ([]string, error) {
	log.Printf("Getting followers for user: %s", userID)
	file, err := os.OpenFile("followers.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil {
		log.Printf("Error decoding follower data: %v", err)
		return nil, err
	}

	var followers []string
	for _, relation := range relations {
		if relation.Followed == userID {
			followers = append(followers, relation.Follower)
		}
	}

	log.Printf("Successfully retrieved followers for user: %s", userID)
	return followers, nil
}

func BanFollower(followerID, followedID string) error {
	log.Printf("Banning follower: %s from user: %s", followerID, followedID)
	file, err := os.OpenFile("followers.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil {
		log.Printf("Error decoding follower data: %v", err)
		return err
	}

	for i, relation := range relations {
		if relation.Follower == followerID && relation.Followed == followedID {
			relations = append(relations[:i], relations[i+1:]...)
			break
		}
	}

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(relations); err != nil {
		log.Printf("Error encoding follower data: %v", err)
		return err
	}

	log.Printf("Successfully banned follower: %s from user: %s", followerID, followedID)
	return nil
}
