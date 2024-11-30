package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	database *mongo.Database
}

func NewMongoDB(uri, dbName, collectionName string) (*MongoDB, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	database := client.Database(dbName)

	return &MongoDB{
		client: client,
		database: database,
	}, nil
}

// InsertDocument يقوم بإدخال مستند في مجموعة MongoDB

func (m *MongoDB) InsertDocument(collection string, document interface{}) error {
    coll := m.client.Database(m.database.Name()).Collection(collection)
    _, err := coll.InsertOne(context.TODO(), document)
    return err
}

// FindDocuments يقوم باسترجاع الوثائق من مجموعة MongoDB
func (db *MongoDB) FindDocuments(collection string, filter interface{}) ([]bson.M, error) {
	var results []bson.M
	collectionRef := db.client.Database("videohls").Collection(collection)

	cursor, err := collectionRef.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Close يقوم بإغلاق الاتصال بقاعدة بيانات MongoDB
func (m *MongoDB) Close() error {
	return m.client.Disconnect(context.Background())
}


