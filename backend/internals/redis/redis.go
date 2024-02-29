package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/Radictionary/kahoot/backend/internals/game"
	"github.com/Radictionary/kahoot/backend/internals/models"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

type RedisConn struct {
	Rdb *redis.Client
}

// InitRedisConnection initializes the Redis instance and Redis connection
func InitRedisConnection() *RedisConn {
	err := godotenv.Load("../../.env")
	if err != nil {
		fmt.Println("Could not load local environment(.env file)")
		return nil
	}
	redisAddr := os.Getenv("REDIS_SERVER_ADDR")
	redisPassword := os.Getenv("REDIS_SERVER_PASSWORD")

	clientConn := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       4,
	})

	Redis := &RedisConn{
		Rdb: clientConn,
	}

	_, err = Redis.Rdb.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("Could not init redis connection")
		os.Exit(1)
	}
	return Redis
}

// Key Value Functions
func (r *RedisConn) StoreData(key string, value string) error {
	ctx := context.Background()
	err := r.Rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to store data: %s", err.Error())
	}
	return nil
}
func (r *RedisConn) RetrieveData(key string) (string, error) {
	ctx := context.Background()
	value, err := r.Rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key '%s' does not exist", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to retrieve data: %s", err.Error())
	}
	return value, nil
}

// Store and Retrieve Map
func (r *RedisConn) RetrieveMap(key string) (map[string]string, error) {
	result, err := r.Rdb.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (r *RedisConn) StoreMap(key string, data map[string]string) error {
	// Convert the map to Redis hash format
	hashData := make([]interface{}, 0, len(data)*2)
	for k, v := range data {
		hashData = append(hashData, k, v)
	}

	// Set the hash data in Redis
	err := r.Rdb.HSet(context.Background(), key, hashData...).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash data in Redis: %v", err)
	}

	return nil
}

// Store and Retrieve User Account
func (r *RedisConn) StoreUserAccount(user models.Account) error {
	// Convert the games slice to JSON
	gamesJSON, err := json.Marshal(user.Games)
	if err != nil {
		return err
	}

	sharedGamesJSON, err := json.Marshal(user.SharedGames)
	if err != nil {
		return err
	}

	redisFields := map[string]interface{}{
		"name":                    user.Name,
		"password":                user.Password,
		"profilePicture": user.ProfilePicture,
		"statistics-lastLoggedIn": user.UserStatistics.LastLoggedIn,
		"games":                   gamesJSON, // Store the JSON representation of games
		"sharedGames":             sharedGamesJSON,
	}
	key := "account:" + user.Name
	err = r.Rdb.HMSet(context.Background(), key, redisFields).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisConn) RetrieveUserAccount(key string) (models.Account, error) {
	result, err := r.Rdb.HGetAll(context.Background(), "account:"+key).Result()
	if err != nil {
		return models.Account{}, err
	}

	user := models.Account{
		Name:     result["name"],
		Password: result["password"],
		ProfilePicture: result["profilePicture"],
		UserStatistics: models.UserStatistics{
			LastLoggedIn: result["statistics-lastLoggedIn"],
		},
	}

	gamesJSON := result["games"]
	var userGames []string
	if len(gamesJSON) > 0 {
		if err := json.Unmarshal([]byte(gamesJSON), &userGames); err != nil {
			return models.Account{}, err
		}
	}
	user.Games = userGames

	sharedGamesJSON := result["sharedGames"]
	var sharedGames []string
	if len(sharedGamesJSON) > 0 {
		if err := json.Unmarshal([]byte(sharedGamesJSON), &sharedGames); err != nil {
			return models.Account{}, err
		}
	}
	user.SharedGames = sharedGames

	return user, nil
}

func (r *RedisConn) ClearUsers(identifier string) error {
	cursor := uint64(0)
	var batchSize int64 = 1000
	for {
		keys, nextCursor, err := r.Rdb.Scan(context.Background(), cursor, identifier+":*", batchSize).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			err := r.Rdb.Del(context.Background(), keys...).Err()
			if err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (r *RedisConn) Remove(identifier string) error {
	err := r.Rdb.Del(context.Background(), identifier)
	if err.Err() != nil {
		return err.Err()
	}
	return nil
}

func (r *RedisConn) GetUserCount() int {
	var userCount int
	var wg sync.WaitGroup
	keys, err := r.Rdb.Keys(context.Background(), "account:*").Result()
	if err != nil {
		fmt.Println("Error retrieving keys:", err)
		return userCount
	}

	countCh := make(chan int, len(keys))

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			count, err := r.Rdb.HLen(context.Background(), k).Result()
			if err != nil {
				fmt.Println("Error retrieving user count:", err)
				countCh <- 0
			}
			countCh <- int(count)
		}(key)
	}

	// Close the countCh channel after all Goroutines have finished
	go func() {
		wg.Wait()
		close(countCh)
	}()

	for count := range countCh {
		userCount += count
	}

	return userCount
}

// StoreGame stores a game in the Redis database.
func (r *RedisConn) StoreGame(game game.Game) error {
	// Convert the game to JSON
	gameJSON, err := json.Marshal(game)
	if err != nil {
		return err
	}

	// Generate a key for the game (e.g., "game:gameName")
	key := fmt.Sprintf("game:%s", game.Name)

	// Store the game data in Redis
	err = r.Rdb.Set(context.Background(), key, gameJSON, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// RetrieveGame retrieves a game from the Redis database by its name.
func (r *RedisConn) RetrieveGame(gameName string) (game.Game, error) {
	// Generate the key for the game
	key := fmt.Sprintf("game:%s", gameName)

	// Retrieve the game data from Redis
	gameJSON, err := r.Rdb.Get(context.Background(), key).Result()
	if err != nil {
		return game.Game{}, err
	}

	// Unmarshal the JSON data into a Game struct
	var Game game.Game
	err = json.Unmarshal([]byte(gameJSON), &Game)
	if err != nil {
		return game.Game{}, err
	}

	return Game, nil
}
