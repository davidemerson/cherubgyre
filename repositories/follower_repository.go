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
	Status   string `json:"status"` // "pending" or "accepted"
}

func AddFollower(followerID, followedID, status string) error {
	log.Printf("Adding follower: %s to user: %s with status: %s", followerID, followedID, status)
	file, err := os.OpenFile("followers.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding follower data: %v", err)
		return err
	}

	for _, relation := range relations {
		if relation.Follower == followerID && relation.Followed == followedID {
			if relation.Status == "pending" && status == "pending" {
				log.Printf("User: %s already has a pending request for: %s", followerID, followedID)
				return errors.New("request already pending")
			}
			if relation.Status == "accepted" || relation.Status == "" {
				log.Printf("User: %s is already following: %s", followerID, followedID)
				return errors.New("already following")
			}
			// If status needs update (e.g. was pending, now accepted), we handle that in AcceptFollower
			// But if we want to support direct overwrite here, we could. For now, strict separation.
		}
	}

	relations = append(relations, FollowerRelation{Follower: followerID, Followed: followedID, Status: status})

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(relations); err != nil {
		log.Printf("Error encoding follower data: %v", err)
		return err
	}

	log.Printf("Successfully added relationship: %s -> %s [%s]", followerID, followedID, status)
	return nil
}

func RemoveFollower(followerID, followedID string) error {
	log.Printf("Removing relationship between follower: %s and user: %s", followerID, followedID)
	file, err := os.OpenFile("followers.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding follower data: %v", err)
		return err
	}

	newRelations := []FollowerRelation{}
	found := false
	for _, relation := range relations {
		if relation.Follower == followerID && relation.Followed == followedID {
			found = true
			continue // Skip adding this to new list (delete)
		}
		newRelations = append(newRelations, relation)
	}

	if !found {
		return errors.New("relationship not found")
	}

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(newRelations); err != nil {
		log.Printf("Error encoding follower data: %v", err)
		return err
	}

	log.Printf("Successfully removed relationship: %s -> %s", followerID, followedID)
	return nil
}

func AcceptFollower(followerID, followedID string) error {
	log.Printf("Accepting follower: %s for user: %s", followerID, followedID)
	file, err := os.OpenFile("followers.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding follower data: %v", err)
		return err
	}

	found := false
	for i, relation := range relations {
		if relation.Follower == followerID && relation.Followed == followedID {
			if relation.Status == "accepted" || relation.Status == "" {
				return errors.New("already accepted")
			}
			relations[i].Status = "accepted"
			found = true
			break
		}
	}

	if !found {
		return errors.New("request not found")
	}

	file.Seek(0, 0)
	file.Truncate(0)

	if err := json.NewEncoder(file).Encode(relations); err != nil {
		log.Printf("Error encoding follower data: %v", err)
		return err
	}

	log.Printf("Successfully accepted follower: %s for user: %s", followerID, followedID)
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
	if err := json.NewDecoder(file).Decode(&relations); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding follower data: %v", err)
		return nil, err
	}

	var followers []string
	for _, relation := range relations {
		if relation.Followed == userID && (relation.Status == "accepted" || relation.Status == "") {
			followers = append(followers, relation.Follower)
		}
	}

	log.Printf("Successfully retrieved followers for user: %s", userID)
	return followers, nil
}

func GetFollowRequests(userID string) ([]string, error) {
	log.Printf("Getting follow requests for user: %s", userID)
	file, err := os.OpenFile("followers.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding follower data: %v", err)
		return nil, err
	}

	var requests []string
	for _, relation := range relations {
		if relation.Followed == userID && relation.Status == "pending" {
			requests = append(requests, relation.Follower)
		}
	}

	log.Printf("Successfully retrieved follow requests for user: %s", userID)
	return requests, nil
}

func BanFollower(followerID, followedID string) error {
	// Banning essentially works same as remove for now in terms of separating relation
	// But logically it might be different in future. Keeping it wrapper for now.
	return RemoveFollower(followerID, followedID)
}

func GetFollowing(userID string) ([]string, error) {
	log.Printf("Getting following list for user: %s", userID)
	file, err := os.OpenFile("followers.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	var relations []FollowerRelation
	if err := json.NewDecoder(file).Decode(&relations); err != nil && err.Error() != "EOF" {
		log.Printf("Error decoding follower data: %v", err)
		return nil, err
	}

	var following []string
	for _, relation := range relations {
		if relation.Follower == userID && (relation.Status == "accepted" || relation.Status == "") {
			following = append(following, relation.Followed)
		}
	}

	log.Printf("Successfully retrieved following list for user: %s", userID)
	return following, nil
}
