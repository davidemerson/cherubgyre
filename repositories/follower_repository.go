package repositories

import (
	"errors"
	"log"
)

type FollowerRelation struct {
	Follower string `json:"follower"`
	Followed string `json:"followed"`
	Status   string `json:"status"` // "pending" or "accepted"
}

var followerStore = newFileStore("followers.json")

func loadFollowers() ([]FollowerRelation, error) {
	var relations []FollowerRelation
	if err := followerStore.load(&relations); err != nil {
		return nil, err
	}
	return relations, nil
}

func AddFollower(followerID, followedID, status string) error {
	log.Printf("Adding follower: %s -> %s [%s]", followerID, followedID, status)
	followerStore.mu.Lock()
	defer followerStore.mu.Unlock()

	var relations []FollowerRelation
	if err := followerStore.loadLocked(&relations); err != nil {
		return err
	}

	for _, relation := range relations {
		if relation.Follower == followerID && relation.Followed == followedID {
			if relation.Status == "pending" && status == "pending" {
				return errors.New("request already pending")
			}
			if relation.Status == "accepted" || relation.Status == "" {
				return errors.New("already following")
			}
		}
	}

	relations = append(relations, FollowerRelation{Follower: followerID, Followed: followedID, Status: status})
	return followerStore.saveLocked(relations)
}

func RemoveFollower(followerID, followedID string) error {
	log.Printf("Removing relationship: %s -> %s", followerID, followedID)
	followerStore.mu.Lock()
	defer followerStore.mu.Unlock()

	var relations []FollowerRelation
	if err := followerStore.loadLocked(&relations); err != nil {
		return err
	}

	newRelations := make([]FollowerRelation, 0, len(relations))
	found := false
	for _, relation := range relations {
		if relation.Follower == followerID && relation.Followed == followedID {
			found = true
			continue
		}
		newRelations = append(newRelations, relation)
	}

	if !found {
		return errors.New("relationship not found")
	}

	return followerStore.saveLocked(newRelations)
}

func AcceptFollower(followerID, followedID string) error {
	log.Printf("Accepting follower: %s -> %s", followerID, followedID)
	followerStore.mu.Lock()
	defer followerStore.mu.Unlock()

	var relations []FollowerRelation
	if err := followerStore.loadLocked(&relations); err != nil {
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

	return followerStore.saveLocked(relations)
}

func GetFollowers(userID string) ([]string, error) {
	relations, err := loadFollowers()
	if err != nil {
		return nil, err
	}

	var followers []string
	for _, relation := range relations {
		if relation.Followed == userID && (relation.Status == "accepted" || relation.Status == "") {
			followers = append(followers, relation.Follower)
		}
	}
	return followers, nil
}

func GetFollowRequests(userID string) ([]string, error) {
	relations, err := loadFollowers()
	if err != nil {
		return nil, err
	}

	var requests []string
	for _, relation := range relations {
		if relation.Followed == userID && relation.Status == "pending" {
			requests = append(requests, relation.Follower)
		}
	}
	return requests, nil
}

func BanFollower(followerID, followedID string) error {
	// Ban is currently implemented as Remove. A future change should add a
	// real banlist that blocks re-requests from the banned follower.
	return RemoveFollower(followerID, followedID)
}

// DeleteUserRelations removes every follower row where the given username
// appears as either the follower or the followed party. Called when a user
// is deregistered so that, if the username is later recycled by a new user
// (per the wordlist space), the new user does not inherit the old user's
// follower graph.
func DeleteUserRelations(username string) error {
	log.Printf("Purging follower relations for: %s", username)
	followerStore.mu.Lock()
	defer followerStore.mu.Unlock()

	var relations []FollowerRelation
	if err := followerStore.loadLocked(&relations); err != nil {
		return err
	}

	kept := make([]FollowerRelation, 0, len(relations))
	for _, r := range relations {
		if r.Follower == username || r.Followed == username {
			continue
		}
		kept = append(kept, r)
	}

	return followerStore.saveLocked(kept)
}

func GetFollowing(userID string) ([]string, error) {
	relations, err := loadFollowers()
	if err != nil {
		return nil, err
	}

	var following []string
	for _, relation := range relations {
		if relation.Follower == userID && (relation.Status == "accepted" || relation.Status == "") {
			following = append(following, relation.Followed)
		}
	}
	return following, nil
}
