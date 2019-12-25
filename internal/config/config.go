package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

const (
	Development = "dev"
	Production  = "production"
)

var (
	GraphqlServerUri    string
	GraphqlServerSecret string
	Port                string
	Env                 string
	RedisHost           string
	EnableCache         bool
)

func Init() {
	err := godotenv.Load()
	GraphqlServerUri = os.Getenv("GRAPHQL_SERVER_URI")
	GraphqlServerSecret = os.Getenv("GRAPHQL_SERVER_SECRET")
	RedisHost = os.Getenv("REDIS_HOST")
	Port = os.Getenv("SERVER_PORT")
	Env = os.Getenv("ENV")
	EnableCache = os.Getenv("ENABLE_CACHE") == "true"
	if Env != Production {
		Env = Development
	}

	fmt.Println(err)
}
