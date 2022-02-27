package dbstore

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/mgo.v2/bson"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	envPath = "env/mongo.env"
)

// Create singleton DB
var db *DB

type DB struct {
	ctx       context.Context
	logger    *log.Logger
	client    *mongo.Client
	snapshots *mongo.Collection
}

func New(ctx context.Context) error {

	logger := log.New(os.Stderr, "[mongo] ", log.LstdFlags)

	// Get environment variables and format uri
	if err := godotenv.Load(envPath); err != nil {
		return err
	}

	// Write db uri
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/%s",
		os.Getenv("MONGOUSER"),
		os.Getenv("MONGOPASS"),
		os.Getenv("MONGOHOST"),
		os.Getenv("MONGOPORT"),
		os.Getenv("MONGONAME"),
	)

	// Connect to mongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	// Get shorthand for snapshots collection
	snapshots := client.Database(os.Getenv("MONGONAME")).
		Collection(os.Getenv("MONGOCOLL"))

	// Ping connection to make sure that the database is on
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	// Insert variables inside db object
	db = &DB{
		ctx:       ctx,
		client:    client,
		logger:    logger,
		snapshots: snapshots,
	}

	return nil
}

// Close closes the mongoDB connection.
func Close() error {
	return db.client.Disconnect(db.ctx)
}

func InsertSnapshot(ctx context.Context, snapshot []byte) error {
	var snapDoc bson.D

	err := bson.UnmarshalJSON(snapshot, &snapDoc)
	if err != nil {
		return err
	}

	result, err := db.snapshots.InsertOne(ctx, snapDoc)
	if err != nil {
		db.logger.Printf("Error inserting snapshot: %s", err.Error())
		return err
	}
	db.logger.Printf("Inserted document with _id: %d", result.InsertedID)

	return nil
}
