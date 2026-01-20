package repository

import (
	"github.com/I1Asyl/berliner_backend/models"
)

type SqlQueries interface {
	FollowChannel(user models.User, channel models.Channel) error
	FollowUser(follower models.User, user models.User) error
	UnfollowChannel(user models.User, channel models.Channel) error
	UnfollowUser(follower models.User, user models.User) error
	GetChannelByName(name string) (models.Channel, error)
	GetUserByUserame(name string) (models.User, error)
	GetUserChannels(user models.User) ([]models.Channel, error)
	AddMembership(models.Membership) error
	AddUser(models.User) error
	AddChannel(channel models.Channel) error
	AddUserPost(post models.UserPost) error
	AddChannelPost(post models.ChannelPost) error
	DeleteUserPost(post models.UserPost) error
	DeleteChannelPost(post models.ChannelPost) error
	AddFollowing(following models.Following) error
	StartTransaction() Transaction
	GetMyChannelPosts(user models.User) ([]struct {
		models.Channel
		models.ChannelPost
	}, error)
	GetChannelPosts(user models.User) ([]struct {
		models.Channel
		models.ChannelPost
	}, error)
	GetUserPosts(user models.User) ([]struct {
		models.User
		models.UserPost
	}, error)
	GetNewUserPosts(user models.User) ([]struct {
		models.User
		models.UserPost
	}, error)
	GetNewChannelPosts(user models.User) ([]struct {
		models.Channel
		models.ChannelPost
	}, error)
	GetFollowing(user models.User) ([]models.User, error)
	UpdateChannel(channel models.Channel) error
	DeleteChannel(channel models.Channel) error
}

type Repository struct {
	SqlQueries
}

//go:generate mockgen -source=repository.go -destination=mocks/repository.go

func NewRepository(dsn string) *Repository {
	return &Repository{SqlQueries: NewDatabase(dsn)}
}
