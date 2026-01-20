package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/I1Asyl/berliner_backend/models"
	"github.com/I1Asyl/berliner_backend/pkg/repository"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var services *Services
var repo *repository.Repository
var testUser models.User

// setupSchema creates all database tables needed for testing
func setupSchema(db *sql.DB) error {
	schema := `
		CREATE TYPE author_type AS ENUM ('user', 'channel');

		CREATE TABLE IF NOT EXISTS "user" (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) NOT NULL,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS channel (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			leader_id INT DEFAULT NULL,
			description TEXT NOT NULL,
			FOREIGN KEY (leader_id) REFERENCES "user"(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS membership (
			id SERIAL PRIMARY KEY,
			channel_id INT NOT NULL,
			user_id INT NOT NULL,
			is_editor BOOLEAN NOT NULL,
			FOREIGN KEY (channel_id) REFERENCES channel(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS request (
			id SERIAL PRIMARY KEY,
			channel_id INT NOT NULL,
			user_id INT NOT NULL,
			is_accepted BOOLEAN NOT NULL,
			FOREIGN KEY (channel_id) REFERENCES channel(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS following (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			follower_id INT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
			FOREIGN KEY (follower_id) REFERENCES "user"(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS user_post (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			author_type author_type NOT NULL,
			is_public BOOLEAN NOT NULL,
			user_id INT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS channel_post (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			author_type author_type NOT NULL,
			is_public BOOLEAN NOT NULL,
			channel_id INT NOT NULL,
			FOREIGN KEY (channel_id) REFERENCES channel(id) ON DELETE CASCADE
		);
	`
	_, err := db.Exec(schema)
	return err
}

// teardownSchema drops all database tables after testing
func teardownSchema(db *sql.DB) error {
	schema := `
		DROP TABLE IF EXISTS membership CASCADE;
		DROP TABLE IF EXISTS request CASCADE;
		DROP TABLE IF EXISTS user_post CASCADE;
		DROP TABLE IF EXISTS channel_post CASCADE;
		DROP TABLE IF EXISTS channel CASCADE;
		DROP TABLE IF EXISTS following CASCADE;
		DROP TABLE IF EXISTS "user" CASCADE;
		DROP TYPE IF EXISTS author_type CASCADE;
	`
	_, err := db.Exec(schema)
	return err
}

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "16-alpine", []string{
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_USER=postgres",
		"POSTGRES_DB=berliner",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	testUser = models.User{
		Id:        1,
		Username:  "asyl",
		FirstName: "Yerassyl",
		LastName:  "Altay",
		Email:     "altayerasyl@gmail.com",
		Password:  "Qqwerty1!.",
	}
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	dsn := fmt.Sprintf("host=localhost port=%s user=postgres password=secret dbname=berliner sslmode=disable", resource.GetPort("5432/tcp"))
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}
	repo = repository.NewRepository(dsn)
	services = NewService(repo)

	// Setup database schema
	if err := setupSchema(db); err != nil {
		log.Fatalf("Could not setup database schema: %s", err)
	}

	code := m.Run()

	// Teardown database schema
	if err := teardownSchema(db); err != nil {
		log.Printf("Could not teardown database schema: %s", err)
	}

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestAddUser(t *testing.T) {
	testTable := []struct {
		name      string
		inputUser models.User
		expected  map[string]string
	}{
		{
			name: "success",
			inputUser: models.User{
				Username:  "test",
				Email:     "email@som.com",
				Password:  "Qqwerty1!.",
				LastName:  "Yerassyl",
				FirstName: "Altay",
			},
			expected: map[string]string{},
		},
		{
			name: "error username",
			inputUser: models.User{
				Username:  "t",
				Email:     "email@som.com",
				Password:  "Qqwerty1!.",
				LastName:  "Yerassyl",
				FirstName: "Altay",
			},
			expected: map[string]string{
				"username": "Invalid username",
			},
		},
		{
			name: "error email",
			inputUser: models.User{
				Username:  "test",
				Email:     "emailsom.com",
				Password:  "Qqwerty1!.",
				LastName:  "Yerassyl",
				FirstName: "Altay",
			},
			expected: map[string]string{
				"email": "Invalid email",
			},
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			err := services.AddUser(testCase.inputUser)
			ans := reflect.DeepEqual(err, testCase.expected)
			if !ans {
				t.Errorf("Expected %v, got %v", testCase.expected, err)
			}
		})
	}
}

func TestCheckUserAndPassword(t *testing.T) {
	testTable := []struct {
		name      string
		inputUser models.AuthorizationForm
		expected  bool
	}{
		{
			name: "success",
			inputUser: models.AuthorizationForm{
				Username: "asyl",
				Password: "Qqwerty1!.",
			},
			expected: true,
		},
		{
			name: "error",
			inputUser: models.AuthorizationForm{
				Username: "asyl",
				Password: "Qqwerty1!",
			},
			expected: false,
		},
	}

	services.AddUser(testUser)
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			ans, _ := services.CheckUserAndPassword(testCase.inputUser)
			if ans != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, ans)
			}
		})
	}
}

func TestGenarateToken(t *testing.T) {
	testTable := []struct {
		name       string
		issueTime  time.Time
		expireTime time.Time
		inputUser  models.AuthorizationForm
		expected   string
		jwt_secret string
	}{
		{
			name:       "success",
			issueTime:  time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			expireTime: time.Date(2009, time.November, 10, 23, 30, 0, 0, time.UTC),
			inputUser: models.AuthorizationForm{
				Username: "asyl",
				Password: "Qqwerty1!.",
			},
			expected:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6ImFzeWwiLCJpc3MiOiJ0ZXN0Iiwic3ViIjoic29tZWJvZHkiLCJleHAiOjEyNTc4OTU4MDAsImlhdCI6MTI1Nzg5NDAwMH0.cWHFSBmmpznRvLw56mokDKpa1Olv4Wy7Pf5YGp3gKFw",
			jwt_secret: "randomJWTSecret",
		},
		{
			name:       "success2",
			issueTime:  time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			expireTime: time.Date(3000, time.November, 10, 23, 30, 0, 0, time.UTC),
			inputUser: models.AuthorizationForm{
				Username: "asyl",
				Password: "Qqwerty1!.",
			},
			expected:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6ImFzeWwiLCJpc3MiOiJ0ZXN0Iiwic3ViIjoic29tZWJvZHkiLCJleHAiOjMyNTMwODA3ODAwLCJpYXQiOjEyNTc4OTQwMDB9.yGe-6MApCd8jvvsuwZH4O9tc3AB-ISBDMYx3xSP_Ork",
			jwt_secret: "randomJWTSecret",
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", testCase.jwt_secret)
			ans, err := services.GenerateToken(testCase.inputUser, testCase.issueTime, testCase.expireTime)
			if ans != testCase.expected {
				t.Errorf("Expected %v, got %v, error: %s", testCase.expected, ans, err)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	testTable := []struct {
		name             string
		token            string
		jwt_secret       string
		expectedUsername string
	}{
		{
			name:             "success",
			token:            "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6ImFzeWwiLCJpc3MiOiJ0ZXN0Iiwic3ViIjoic29tZWJvZHkiLCJleHAiOjMyNTMwODA3ODAwLCJpYXQiOjEyNTc4OTQwMDB9.yGe-6MApCd8jvvsuwZH4O9tc3AB-ISBDMYx3xSP_Ork",
			jwt_secret:       "randomJWTSecret",
			expectedUsername: "asyl",
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", testCase.jwt_secret)
			ans, err := services.ParseToken(testCase.token)
			if ans != testCase.expectedUsername {
				t.Errorf("Expected %v, got %v, error: %s", testCase.expectedUsername, ans, err)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	testTable := []struct {
		name     string
		password string
		hash     string
	}{
		{
			name:     "success",
			password: "Qqwerty1!.",
			hash:     "$2a$11$IZph6sLg28fsOA2qD6xhsO2pWvnL9ihKkalZgpAeG.Nl6I8QN.Y4m",
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			err := bcrypt.CompareHashAndPassword([]byte(testCase.hash), []byte(testCase.password))
			if err != nil {
				t.Errorf("Expected %v, is not a hash of %v", testCase.hash, testCase.password)
			}
		})
	}
}

func TestCreateChannel(t *testing.T) {
	testTable := []struct {
		name          string
		channel       models.Channel
		channelLeader models.User
		expected      map[string]string
	}{
		{
			name: "success",
			channel: models.Channel{
				Name:        "Channel",
				LeaderId:    1,
				Description: "hoho",
			},
			channelLeader: testUser,
			expected:      map[string]string{},
		},
		{
			name: "error",
			channel: models.Channel{
				Name:        "Channel",
				LeaderId:    1,
				Description: "",
			},
			channelLeader: testUser,
			expected: map[string]string{
				"description": "Channel description can not be empty",
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			services.AddUser(testUser)
			err := services.CreateChannel(testCase.channel, testUser)
			if !reflect.DeepEqual(err, testCase.expected) {
				t.Errorf("Expected %v, got %v", testCase.expected, err)
			}
		})
	}

}

// gets User model by username in the transaction
func TestGetUserByUsername(t *testing.T) {
	testTable := []struct {
		name     string
		username string
		expected models.User
	}{
		{
			name:     "success",
			username: "asyl",
			expected: testUser,
		},
		{
			name:     "error",
			username: "x",
			expected: models.User{},
		},
	}
	services.AddUser(testUser)
	for _, testCase := range testTable {
		user, _ := services.GetUserByUsername(testCase.username)

		// user's password is hashed, so we don't need to compare it
		user.Password = testCase.expected.Password

		if !reflect.DeepEqual(user, testCase.expected) {
			t.Errorf("Expected %v, got %v", testCase.expected, user)
		}
	}
}

// // create a new channel in the database for the given user
// func (a ApiService) CreateChannel(channel models.Channel, user models.User) map[string]string {

// 	invalid := channel.IsValid()
// 	channel.LeaderId = user.Id
// 	tx := a.repo.SqlQueries.StartTransaction()

// 	if len(invalid) == 0 {
// 		if err := tx.AddChannel(channel); err != nil {
// 			invalid["common"] = err.Error()
// 		} else {
// 			channel, _ = tx.GetChannelByName(channel.Name)
// 			membership := models.Membership{UserId: channel.LeaderId, ChannelId: channel.Id, IsEditor: true}
// 			tx.AddMembership(membership)

// 		}
// 	}
// 	err := tx.Commit()
// 	if err != nil {
// 		tx.Rollback()
// 	}

// 	return invalid
// }

// // create a new post in the database for the given user or channel
// func (a ApiService) CreatePost(post models.Post) map[string]string {

// 	invalid := post.IsValid()

// 	if len(invalid) == 0 {
// 		err := a.repo.SqlQueries.AddPost(post)
// 		if err != nil {
// 			invalid["common"] = err.Error()
// 		}
// 	}

// 	return invalid
// }

// // get all user's channel posts from the database
// func (a ApiService) GetPostsFromChannels(user models.User) ([]models.Post, error) {
// 	posts, err := a.repo.SqlQueries.GetChannelPosts(user)
// 	return posts, err
// }

// // get all user's following's posts from the database
// func (a ApiService) GetPostsFromUsers(user models.User) ([]models.Post, error) {
// 	posts, err := a.repo.SqlQueries.GetUserPosts(user)
// 	return posts, err
// }

// // get all posts available for the given user from the database
// func (a ApiService) GetAllPosts(user models.User) ([]models.Post, error) {
// 	channelPosts, _ := a.GetPostsFromChannels(user)
// 	userPosts, _ := a.GetPostsFromUsers(user)
// 	posts := channelPosts
// 	posts = append(posts, userPosts...)
// 	return posts, nil
// }

// func (a ApiService) GetFollowing(user models.User) ([]models.User, error) {
// 	users, err := a.repo.SqlQueries.GetFollowing(user)
// 	return users, err
// }

// func (a ApiService) DeleteChannel(channel models.Channel) error {
// 	err := a.repo.SqlQueries.DeleteChannel(channel)
// 	return err
// }

// func (a ApiService) UpdateChannel(channel models.Channel) error {
// 	err := a.repo.SqlQueries.UpdateChannel(channel)
// 	return err
// }
