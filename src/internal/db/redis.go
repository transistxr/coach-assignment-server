package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisClient struct {
	Client *redis.Client
}

func InitRDB(db *sql.DB) *RedisClient {

	log.Println("Initializing Redis Database Connection")
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("could not connect to redis: %v", err)
	}

	log.Println("Connected to Redis")

	log.Println("Loading API Keys into Redis..")

	rows, err := db.Query(`SELECT key_hash, name, rate_limit FROM api_keys`)
	if err != nil {
	}
	defer rows.Close()

	for rows.Next() {
		var key_hash string
		var name string
		var rate_limit float64

		if err := rows.Scan(&key_hash, &name, &rate_limit); err != nil {
			log.Fatal(err)
		}

		switch name{
			case "Test API Key":
			 rdb.Set(ctx, "TEST_API_KEY",key_hash, 0)
			 rdb.Set(ctx, "TEST_ENV_RATE_LIMIT",rate_limit, 0)

			case "Development Key":
			 rdb.Set(ctx, "DEVELOPMENT_KEY",key_hash, 0)
			 rdb.Set(ctx, "DEV_ENV_RATE_LIMIT",rate_limit, 0)

			case "Production Key":
			 rdb.Set(ctx, "PRODUCTION_KEY",key_hash, 0)
			 rdb.Set(ctx, "PROD_ENV_RATE_LIMIT",rate_limit, 0)

		}
	}

	return &RedisClient{Client: rdb}

}

func GetValueString(rdb *RedisClient, key string) string {

	log.Printf("Reading Redis key: %s \n", key)

    val, err := rdb.Client.Get(ctx, key).Result()
    if err != nil {
        panic(err)
    }

	return val
}

func SetIdempotencyKey(rdb *RedisClient, idemKey string, response any) error {

	log.Printf("Setting idempotency key %s with response: %s \n", idemKey, response)
	b, err := json.Marshal(response)
	if err != nil {
		return err
	}
	rdb.Client.Set(ctx, idemKey, b, 24*time.Hour)
	return nil
}

func CheckIdempotency(rdb *RedisClient, idemKey string, response any) (bool, any) {

	log.Printf("Checking if idempotency key %s is in DB \n", idemKey)

	val, err := rdb.Client.Get(ctx, idemKey).Result()
	if err != nil {
		return false, nil
	}

	log.Printf("Val %s", val)

	if err := json.Unmarshal([]byte(val), response); err != nil {
		return false, nil
	}
	return true, response
}
