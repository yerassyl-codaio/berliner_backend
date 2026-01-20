package repository

import (
	"fmt"
	"log"

	"github.com/I1Asyl/berliner_backend/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Database struct {
	*sqlx.DB
}

type Transaction struct {
	*sqlx.Tx
}

// SetupOrm sets up the database connection
func NewDatabase(dsn string) Database {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("init")

	return Database{db}
}

func (db Database) StartTransaction() Transaction {
	tx := db.MustBegin()
	return Transaction{tx}
}

// func (db Database) (tx Transaction) {
// 	tx.Commit()
// }

func (db Database) GetChannelByName(name string) (models.Channel, error) {
	var channel models.Channel
	err := db.Get(&channel, "SELECT * FROM channel WHERE name = $1", name)
	return channel, err
}

func (db Transaction) GetChannelByName(name string) (models.Channel, error) {
	var channel models.Channel
	err := db.Get(&channel, "SELECT * FROM channel WHERE name = $1", name)
	return channel, err
}

func (db Database) GetUserByUserame(username string) (models.User, error) {
	var user models.User
	err := db.Get(&user, `SELECT * FROM "user" WHERE username = $1`, username)
	return user, err
}
func (db Transaction) GetUserByUserame(username string) (models.User, error) {
	var user models.User
	err := db.Get(&user, `SELECT * FROM "user" WHERE username = $1`, username)
	return user, err
}

func (db Database) GetUserChannels(user models.User) ([]models.Channel, error) {
	var channels []models.Channel
	err := db.Select(&channels, "SELECT * FROM channel WHERE leader_id = $1", user.Id)
	return channels, err
}
func (db Transaction) GetUserChannels(user models.User) ([]models.Channel, error) {
	var channels []models.Channel
	err := db.Select(&channels, "SELECT * FROM channel WHERE leader_id = $1", user.Id)
	return channels, err
}

func (db Database) AddUser(user models.User) error {
	_, err := db.Exec(`INSERT INTO "user" (username, first_name, last_name, email, password) VALUES ($1, $2, $3, $4, $5)`, user.Username, user.FirstName, user.LastName, user.Email, user.Password)
	return err
}
func (db Transaction) AddUser(user models.User) error {
	_, err := db.Exec(`INSERT INTO "user" (username, first_name, last_name, email, password) VALUES ($1, $2, $3, $4, $5)`, user.Username, user.FirstName, user.LastName, user.Email, user.Password)
	return err
}

func (db Database) AddMembership(membership models.Membership) error {
	_, err := db.Exec("INSERT INTO membership (channel_id, user_id, is_editor) VALUES ($1, $2, $3)", membership.ChannelId, membership.UserId, membership.IsEditor)
	return err
}
func (db Transaction) AddMembership(membership models.Membership) error {
	_, err := db.Exec("INSERT INTO membership (channel_id, user_id, is_editor) VALUES ($1, $2, $3)", membership.ChannelId, membership.UserId, membership.IsEditor)
	return err
}

func (db Database) AddChannel(channel models.Channel) error {
	_, err := db.Exec("INSERT INTO channel (name, leader_id, description) VALUES ($1, $2, $3)", channel.Name, channel.LeaderId, channel.Description)
	return err
}
func (db Transaction) AddChannel(channel models.Channel) error {
	_, err := db.Exec("INSERT INTO channel (name, leader_id, description) VALUES ($1, $2, $3)", channel.Name, channel.LeaderId, channel.Description)
	return err
}

func (db Database) AddUserPost(post models.UserPost) error {
	_, err := db.Exec("INSERT INTO user_post (author_type, content, updated_at, created_at, user_id, is_public) VALUES ($1, $2, $3, $4, $5, $6);", post.AuthorType, post.Content, post.UpdatedAt, post.CreatedAt, post.UserId, post.IsPublic)
	return err
}
func (db Database) AddChannelPost(post models.ChannelPost) error {
	_, err := db.Exec("INSERT INTO channel_post (author_type, content, updated_at, created_at, channel_id, is_public) VALUES ($1, $2, $3, $4, $5, $6);", post.AuthorType, post.Content, post.UpdatedAt, post.CreatedAt, post.ChannelId, post.IsPublic)
	fmt.Println(post.ChannelId)
	return err
}
func (db Database) DeleteUserPost(post models.UserPost) error {
	_, err := db.Exec("DELETE FROM user_post WHERE id = $1;", post.Id)
	return err
}

func (db Database) DeleteChannelPost(post models.ChannelPost) error {
	_, err := db.Exec("DELETE FROM channel_post WHERE id = $1;", post.Id)
	return err
}

func (db Database) GetUserPosts(user models.User) ([]struct {
	models.User
	models.UserPost
}, error) {
	//var posts []models.UserPost
	users := "SELECT following.user_id FROM following WHERE following.follower_id=$1"
	var newTable []struct {
		models.User
		models.UserPost
	}

	err := db.Select(&newTable, fmt.Sprintf(`SELECT user_post.*, "user".username, "user".first_name, "user".last_name FROM user_post LEFT JOIN "user" on user_post.user_id = "user".id WHERE (user_post.user_id in (%v) AND user_post.is_public) OR user_post.user_id = $2 ORDER BY updated_at DESC`, users), user.Id, user.Id)
	return newTable, err

}

func (db Database) GetChannelPosts(user models.User) ([]struct {
	models.Channel
	models.ChannelPost
}, error) {
	//var posts []models.UserPost
	channels := "SELECT membership.channel_id FROM membership WHERE membership.user_id=$1"
	var newTable []struct {
		models.Channel
		models.ChannelPost
	}
	err := db.Select(&newTable, fmt.Sprintf("SELECT channel_post.*, channel.name, channel.leader_id FROM channel_post LEFT JOIN channel on channel_post.channel_id = channel.id WHERE channel_post.channel_id in (%v) AND (channel_post.is_public OR channel.leader_id = $2) ORDER BY updated_at DESC", channels), user.Id, user.Id)
	return newTable, err

}
func (db Database) GetMyChannelPosts(user models.User) ([]struct {
	models.Channel
	models.ChannelPost
}, error) {
	//var posts []models.UserPost
	var newTable []struct {
		models.Channel
		models.ChannelPost
	}
	err := db.Select(&newTable, "SELECT channel_post.*, channel.name, channel.leader_id FROM channel_post LEFT JOIN channel on channel_post.channel_id = channel.id WHERE channel.leader_id = $1 ORDER BY updated_at DESC", user.Id)
	return newTable, err

}

func (db Database) GetNewUserPosts(user models.User) ([]struct {
	models.User
	models.UserPost
}, error) {
	//var posts []models.UserPost
	users := "SELECT following.user_id FROM following WHERE following.follower_id=$1"
	var newTable []struct {
		models.User
		models.UserPost
	}

	err := db.Select(&newTable, fmt.Sprintf(`SELECT user_post.*, "user".username, "user".first_name, "user".last_name FROM user_post LEFT JOIN "user" on user_post.user_id = "user".id WHERE user_post.user_id NOT in (%v) AND NOT user_post.user_id = $2 AND user_post.is_public = true ORDER BY updated_at DESC`, users), user.Id, user.Id)
	return newTable, err
}

func (db Database) GetNewChannelPosts(user models.User) ([]struct {
	models.Channel
	models.ChannelPost
}, error) {
	//var posts []models.UserPost
	users := "SELECT membership.channel_id FROM membership WHERE membership.user_id=$1"
	var newTable []struct {
		models.Channel
		models.ChannelPost
	}

	err := db.Select(&newTable, fmt.Sprintf("SELECT channel_post.*, channel.name FROM channel_post LEFT JOIN channel on channel_post.channel_id = channel.id WHERE channel_post.channel_id NOT in (%v) AND channel_post.is_public = true ORDER BY updated_at DESC", users), user.Id)
	return newTable, err
}

func (db Database) FollowChannel(user models.User, channel models.Channel) error {
	query := "INSERT INTO membership (channel_id, user_id, is_editor) VALUES ($1, $2, $3)"
	_, err := db.Exec(query, channel.Id, user.Id, false)
	return err
}
func (db Database) FollowUser(follower models.User, user models.User) error {
	query := "INSERT INTO following (user_id, follower_id) VALUES ($1, $2)"
	_, err := db.Exec(query, user.Id, follower.Id)
	return err
}

func (db Database) UnfollowChannel(user models.User, channel models.Channel) error {
	query := "DELETE FROM membership WHERE channel_id = $1 AND user_id = $2"
	_, err := db.Exec(query, channel.Id, user.Id)
	return err
}
func (db Database) UnfollowUser(follower models.User, user models.User) error {
	query := "DELETE FROM following WHERE user_id = $1 AND follower_id = $2"
	_, err := db.Exec(query, user.Id, follower.Id)
	return err
}

func (db Database) GetFollowing(user models.User) ([]models.User, error) {
	var users []models.User
	err := db.Select(&users, "SELECT * FROM following WHERE follower_id = $1", user.Id)
	return users, err
}

func (db Database) DeleteChannel(channel models.Channel) error {
	_, err := db.Exec("DELETE FROM channel WHERE id = $1", channel.Id)
	return err
}

func (db Database) AddFollowing(following models.Following) error {
	_, err := db.Exec("INSERT INTO following (follower_id, user_id) VALUES ($1, $2)", following.FollowerId, following.UserId)
	return err
}

func (db Database) UpdateChannel(channel models.Channel) error {
	if channel.Name != "" {
		_, err := db.Exec("UPDATE channel SET name = $1 WHERE channel_id = $2", channel.Name, channel.Id)
		if err != nil {
			return err
		}
	}
	if channel.Description != "" {
		_, err := db.Exec("UPDATE channel SET description = $1 WHERE channel_id = $2", channel.Name, channel.Id)
		if err != nil {
			return err
		}
	}
	return nil
}
