package clerkservice

import (
	"context"
	"errors"

	"github.com/clerkinc/clerk-sdk-go/clerk"
)

type ClerkServiceInterface interface {
	GetUser(ctx context.Context, clerkUserId string) (clerk.User, error)
}

type ClerkService struct {
	client *clerk.Client
}

func NewClerkService(client *clerk.Client) *ClerkService {
	return &ClerkService{client: client}
}

func (cs *ClerkService) GetUser(clerkUserId string) (clerk.User, error) {
	users, err := (*cs.client).Users().ListAll(clerk.ListAllUsersParams{UserIDs: []string{clerkUserId}})

	if err != nil {
		return clerk.User{}, err
	}

	if len(users) == 0 {
		return clerk.User{}, errors.New("user not found")
	}

	return users[0], nil
}
