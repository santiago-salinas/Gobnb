package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pocketbase_go/logger"
	"pocketbase_go/my_models"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"

	"github.com/go-redis/redis/v8"
)

const (
	usersCollection = "users"
	rolesCollection = "roles"

	test_user_one   = "User1"
	test_user_two   = "User2"
	test_user_three = "User3"

	test_token_one   = "test_token_one"
	test_token_two   = "test_token_two"
	test_token_three = "test_token_three"
)

type PocketUserRepo struct {
	Db           pocketbase.PocketBase
	Cache        *redis.Client
	clientSecret string
	clientId     string
	tokenUrl     string
}

type Response struct {
	User struct {
		Login string `json:"login"`
	} `json:"user"`
}

func (r *PocketUserRepo) SetConfigVariables(clientSecret, clientId, tokenUrl string) {
	r.clientSecret = clientSecret
	r.clientId = clientId
	r.tokenUrl = tokenUrl
}

func getTestUser(token string) string {
	tokens := map[string]string{
		test_token_one:   test_user_one,
		test_token_two:   test_user_two,
		test_token_three: test_user_three,
	}
	return tokens[token]
}

func (r *PocketUserRepo) AddUser(token string) error {
	logger.Info("Repo: Adding user with token: ", token)

	username, err := sendPostRequest(token, r.clientId, r.clientSecret)
	if err != nil {
		return err
	}

	collection, err := r.Db.Dao().FindCollectionByNameOrId(usersCollection)
	if err != nil {
		return err
	}

	record, err := r.Db.Dao().FindFirstRecordByData(usersCollection, "username", username)
	if record != nil {
		logger.Error("Repo: User already exists")
		return fmt.Errorf("error: user already exists")
	}

	if err != nil && err.Error() == "sql: no rows in result set" {
		record := models.NewRecord(collection)
		record.Set("username", username)
		record.Set("email", username+"@pocketbase.com")
		err = r.Db.Dao().SaveRecord(record)
		if err != nil {
			logger.Error("Repo: ", err)
			return err
		}

		return nil
	}

	if err != nil {
		logger.Error("Repo: ", err)
		return err
	}

	logger.Info("Repo: User added successfully")
	return nil
}

func (r *PocketUserRepo) Login(token string) (roles []string, id string, err error) {
	logger.Info("Repo: Logging in user with token: ", token)
	if r.Cache != nil {
		val, err := r.Cache.Get(ctx, token).Result()
		if err == redis.Nil || err != nil {
			roles, id, err = r.userFromDB(token)
			if err != nil {
				return nil, "", err
			}

			err = r.storeUserInCache(token, roles, id)
			if err != nil {
				logger.Warn("Could not store login in cache: ", err)
			}

			return roles, id, nil
		} else {
			var cachedData struct {
				Roles []string `json:"roles"`
				ID    string   `json:"id"`
			}

			err := json.Unmarshal([]byte(val), &cachedData)
			if err != nil {
				return nil, "", err
			}

			return cachedData.Roles, cachedData.ID, nil
		}
	} else {
		return r.userFromDB(token)
	}
}

func (r *PocketUserRepo) userFromDB(token string) (roles []string, id string, err error) {
	username := getTestUser(token)
	if username == "" {
		username, err = sendPostRequest(token, r.clientId, r.clientSecret)
		if err != nil {
			logger.Error("Repo: Error trying to login user with token : ", token, err)
			err = fmt.Errorf("invalid token")
			return nil, "", err
		}
	}

	record, err := r.Db.Dao().FindFirstRecordByData(usersCollection, "username", username)
	if err != nil {
		logger.Error("Repo: Error trying to login user with token : ", token, err)
		return nil, "", err
	}

	roles, ok := record.Get("roles").([]string) // Perform type assertion here
	if !ok {
		logger.Error("Repo: Error converting roles to []string")
		return nil, "", fmt.Errorf("error converting roles to []string")
	}

	userId := record.GetString("id")

	fmt.Printf("Roles: %v, id: %v\n", roles, userId)

	logger.Info("User logged in: ", username)
	return roles, userId, nil
}

func (r *PocketUserRepo) storeUserInCache(token string, roles []string, id string) error {
	ttlInMinutes := 2 * time.Minute

	cachedData := struct {
		Roles []string `json:"roles"`
		ID    string   `json:"id"`
	}{
		Roles: roles,
		ID:    id,
	}

	cachedDataJSON, err := json.Marshal(cachedData)
	if err != nil {
		return err
	}

	if r.Cache != nil {
		err = r.Cache.Set(ctx, token, cachedDataJSON, ttlInMinutes).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PocketUserRepo) GetUsersByRole(role string) ([]string, error) {
	records, err := r.Db.Dao().FindRecordsByFilter(
		usersCollection, "roles ~ {:role}", "", 0, 0, dbx.Params{"role": role})
	if err != nil {
		logger.Error("Repo: %v", err)
		return nil, err
	}

	var ret []string

	for _, record := range records {
		ret = append(ret, record.GetString("email"))
	}

	logger.Info("Repo: Users with role : %v", role, ret)
	return ret, nil
}

func (r *PocketUserRepo) GetPropertyOwner(propertyId string) (string, error) {
	logger.Info("Repo: Getting owner of property ", propertyId)
	propertyRecord, err := r.Db.Dao().FindRecordById("properties", propertyId)
	if err != nil {
		logger.Error("Repo: %v", err)
		return "", err
	}

	propertyOwner := propertyRecord.Get("owner").(string)
	ownerRecord, err := r.Db.Dao().FindRecordById(usersCollectionName, propertyOwner)
	if err != nil {
		logger.Error("Repo: %v", err)
		return "", err
	}

	logger.Info("Repo: Owner of property : ", propertyId, ownerRecord.GetString("email"))
	return ownerRecord.GetString("email"), nil
}

func (r *PocketUserRepo) GetUserById(userId string) (my_models.User, error) {
	logger.Info("Repo: Getting user with id ", userId)
	record, err := r.Db.Dao().FindRecordById(usersCollection, userId)
	if err != nil {
		logger.Error("Repo: %v", err)
		return my_models.User{}, err
	}

	user := my_models.User{
		ID:       record.GetString("id"),
		Email:    record.GetString("email"),
		Username: record.GetString("username"),
		Roles:    record.Get("roles").([]string),
	}

	logger.Info("Repo: User with id : %v", userId, user)
	return user, nil
}

func sendPostRequest(token, client_id, client_secret string) (username string, err error) {
	logger.Info("Verifying token: ", token)
	url := fmt.Sprintf("https://api.github.com/applications/%s/token", client_id)
	tokenData := map[string]string{
		"access_token": token,
	}

	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		logger.Error("Repo: Error marshalling JSON: %v", err)
		return "", fmt.Errorf("error marshalling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Repo: Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.SetBasicAuth(client_id, client_secret)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Repo: Error sending request: %v", err)
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Repo: Unexpected status code: %v", resp.StatusCode)
		return "", fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Repo: Error reading response body: %v", err)
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Error(err, "Error unmarshalling JSON: %v", err)
		return "", fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	logger.Info("Repo: Username: ", response.User.Login)
	return response.User.Login, nil
}
