package mongodb

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/omareloui/skrewscore/internal/game"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	gamesCol    *mongo.Collection
)

func Init() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping: %v", err)
	}
	mongoClient = client
	gamesCol = client.Database("skrew").Collection("games")
	log.Println("Connected to MongoDB")
}

func Disconnect(ctx context.Context) error {
	return mongoClient.Disconnect(ctx)
}

func SaveGame(g *game.Game) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.Replace().SetUpsert(true)
	_, err := gamesCol.ReplaceOne(ctx, bson.M{"_id": g.ID}, g, opts)
	return err
}

func LoadGame(id string) (*game.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var g game.Game
	err := gamesCol.FindOne(ctx, bson.M{"_id": id}).Decode(&g)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &g, err
}
