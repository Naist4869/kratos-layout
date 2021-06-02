package nosql

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Uri  string
	Port string
	User string
	Pass string
	DB   string
}

func NewMongo(c *Config) (client *mongo.Client, err error) {
	auth := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		Username:      c.User,
		Password:      c.Pass,
		AuthSource:    c.DB,
	}
	client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", c.Uri)).SetAuth(auth).SetConnectTimeout(time.Second*10))
	return
}
