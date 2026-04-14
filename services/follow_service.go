package services

import (
	"cherubgyre/dtos"
	"cherubgyre/repositories"
	"log/slog"
)

// The follow service now takes already-authenticated usernames from the
// controllers (via the RequireAuth middleware) instead of raw tokens.
// This collapses the layer of per-function ValidateToken + GetUsernameFromToken
// calls that were being done inconsistently at the service layer.

func FollowUser(followerUsername, targetUsername string) error {
	if err := repositories.AddFollower(followerUsername, targetUsername, "pending"); err != nil {
		slog.Error("add follow request failed",
			slog.String("follower", followerUsername),
			slog.String("target", targetUsername),
			slog.Any("err", err))
		return err
	}
	return nil
}

func AcceptFollow(currentUsername, requesterUsername string) error {
	if err := repositories.AcceptFollower(requesterUsername, currentUsername); err != nil {
		slog.Error("accept follower failed",
			slog.String("user", currentUsername),
			slog.String("requester", requesterUsername),
			slog.Any("err", err))
		return err
	}
	return nil
}

func DeclineFollow(currentUsername, requesterUsername string) error {
	if err := repositories.RemoveFollower(requesterUsername, currentUsername); err != nil {
		slog.Error("decline follower failed",
			slog.String("user", currentUsername),
			slog.String("requester", requesterUsername),
			slog.Any("err", err))
		return err
	}
	return nil
}

func UnfollowUser(followerUsername, targetUsername string) error {
	if err := repositories.RemoveFollower(followerUsername, targetUsername); err != nil {
		slog.Error("remove follower failed",
			slog.String("follower", followerUsername),
			slog.String("target", targetUsername),
			slog.Any("err", err))
		return err
	}
	return nil
}

func GetFollowRequests(username string) ([]dtos.UserResponseDTO, error) {
	requestUsernames, err := repositories.GetFollowRequests(username)
	if err != nil {
		slog.Error("get follow requests failed", slog.String("user", username), slog.Any("err", err))
		return nil, err
	}
	return lookupUsers(requestUsernames), nil
}

func GetFollowers(username string) ([]dtos.UserResponseDTO, error) {
	followerUsernames, err := repositories.GetFollowers(username)
	if err != nil {
		slog.Error("get followers failed", slog.String("user", username), slog.Any("err", err))
		return nil, err
	}
	return lookupUsers(followerUsernames), nil
}

func BanFollower(currentUsername, targetUsername string) error {
	if err := repositories.BanFollower(targetUsername, currentUsername); err != nil {
		slog.Error("ban follower failed",
			slog.String("user", currentUsername),
			slog.String("target", targetUsername),
			slog.Any("err", err))
		return err
	}
	return nil
}

func GetFollowing(username string) ([]dtos.UserResponseDTO, error) {
	followingUsernames, err := repositories.GetFollowing(username)
	if err != nil {
		slog.Error("get following failed", slog.String("user", username), slog.Any("err", err))
		return nil, err
	}
	return lookupUsers(followingUsernames), nil
}

// lookupUsers enriches a list of usernames with avatar info. Missing
// users are skipped rather than errored — a deleted account in someone's
// follower list shouldn't break the whole response.
func lookupUsers(usernames []string) []dtos.UserResponseDTO {
	out := make([]dtos.UserResponseDTO, 0, len(usernames))
	for _, name := range usernames {
		user, err := repositories.GetUserByID(name)
		if err != nil {
			continue
		}
		out = append(out, dtos.UserResponseDTO{
			Username: user.Username,
			Avatar:   user.Avatar,
		})
	}
	return out
}
