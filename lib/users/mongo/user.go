package mongo

import (
	"context"
	"time"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/satori/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//e.g. Users("mongodb://localhost:27017", "taxiching")
func Users(mongoURI string, dbName string) (users.IUsers, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, log.Wrapf(err, "Failed to create mongo client to %s", mongoURI)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, log.Wrapf(err, "Failed to connect to mongo %s", mongoURI)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, log.Wrapf(err, "Failed to check mongo %s", mongoURI)
	}

	collection := client.Database(dbName).Collection("users")
	return &factory{
		collection: collection,
	}, nil
} //Users()

//mongoUser implements IUser
type mongoUser struct {
	id       string
	msisdn   string
	name     string
	password string
}

func (u mongoUser) ID() string {
	return u.id
}

func (u mongoUser) Msisdn() string {
	return u.msisdn
}

func (u mongoUser) Name() string {
	return u.name
}

func (u mongoUser) Auth(password string) bool {
	if u.password == password {
		return true
	}
	return false
}

func (u *mongoUser) SetPassword(oldPassword, newPassword string) error {
	if u.password != oldPassword {
		return log.Wrapf(nil, "Incorrect old password")
	}
	p, err := users.ValidatePassword(newPassword)
	if err != nil {
		return log.Wrapf(nil, "Cannot set invalid password")
	}
	u.password = p
	return nil
}

type factory struct {
	collection *mongo.Collection
}

func (f factory) New(msisdn, name, password string) (users.IUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m, err := users.ValidateMsisdn(msisdn)
	if err != nil {
		return nil, log.Wrapf(err, "cannot create user with invalid msisdn")
	}
	n, err := users.ValidateName(name)
	if err != nil {
		return nil, log.Wrapf(err, "cannot create user with invalid name")
	}
	p, err := users.ValidatePassword(password)
	if err != nil {
		return nil, log.Wrapf(err, "cannot create user with invalid password")
	}

	//make sure msisdn is uniq
	if f.GetMsisdn(m) != nil {
		return nil, log.Wrapf(nil, "User already exists with msisdn=%s", m)
	}

	id := uuid.NewV1().String()
	_, err = f.collection.InsertOne(
		ctx,
		bson.M{
			"id":       id,
			"msisdn":   m,
			"name":     n,
			"password": p,
		})
	if err != nil {
		return nil, log.Wrapf(err, "failed to insert user into db")
	}
	u := &mongoUser{
		id:       id,
		msisdn:   msisdn,
		name:     name,
		password: password,
	}
	return u, nil
} //factory.New()

func (f factory) GetMsisdn(msisdn string) users.IUser {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := f.collection.Find(ctx, bson.M{"msisdn": msisdn})
	if err != nil {
		log.Errorf("Failed to find msisdn: %v", err)
		return nil
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Failed to get data: %v", err)
			return nil
		}
		// do something with result....
		log.Debugf("GOT (%T): %+v", result, result)
		u := mongoUser{
			id:       result["id"].(string),
			name:     result["name"].(string),
			msisdn:   result["msisdn"].(string),
			password: result["password"].(string),
		}
		return &u
	}
	if err := cur.Err(); err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}
	return nil
} //factory.GetMsisdn()

func (f factory) GetID(id string) users.IUser {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := f.collection.Find(ctx, bson.M{"id": id})
	if err != nil {
		log.Errorf("Failed to find id: %v", err)
		return nil
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Failed to get data: %v", err)
			return nil
		}
		// do something with result....
		log.Debugf("GOT (%T): %+v", result, result)
		u := mongoUser{
			id:       result["id"].(string),
			name:     result["name"].(string),
			msisdn:   result["msisdn"].(string),
			password: result["password"].(string),
		}
		return &u
	}
	if err := cur.Err(); err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}
	return nil
} //factory.GetID()
