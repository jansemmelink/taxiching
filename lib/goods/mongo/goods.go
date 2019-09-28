package mongo

import (
	"context"
	"time"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/goods"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
	"github.com/satori/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//e.g. Products("mongodb://localhost:27017", "taxiching")
func Products(mongoURI string, dbName string, users users.IUsers) (goods.IProducts, error) {
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

	// err = client.Ping(ctx, readpref.Primary())
	// if err != nil {
	// 	return nil, log.Wrapf(err, "Failed to check mongo %s", mongoURI)
	// }

	collection := client.Database(dbName).Collection("products")
	return &factory{
		users:      users,
		collection: collection,
	}, nil
} //Products()

func (f *factory) UserGoods(userID string) (map[string]goods.IProduct, bool) {
	return nil, true //todo
} //factory.UserGoods()

//mongoProduct implements IUser
type mongoProduct struct {
	id    string
	owner users.IUser
	name  string
	cost  wallets.Amount
}

func (u mongoProduct) ID() string {
	return u.id
}

func (u mongoProduct) Owner() users.IUser {
	return u.owner
}

func (u mongoProduct) Name() string {
	return u.name
}

func (u mongoProduct) Cost() wallets.Amount {
	return u.cost
}

type factory struct {
	users      users.IUsers
	collection *mongo.Collection
}

func (f factory) New(userID string, name string, cost wallets.Amount) (goods.IProduct, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u := f.users.GetID(userID)
	if u == nil {
		return nil, log.Wrapf(nil, "unknown user id")
	}

	if p := f.GetUserProduct(u, name); p != nil {
		return nil, log.Wrapf(nil, "User(%s).Product(%s) already exists", userID, name)
	}

	id := uuid.NewV1().String()
	_, err := f.collection.InsertOne(
		ctx,
		bson.M{
			"id":    id,
			"owner": userID,
			"name":  name,
			"cost":  cost,
		})
	if err != nil {
		return nil, log.Wrapf(err, "failed to insert product into db")
	}

	p := &mongoProduct{
		id:    id,
		owner: u,
		name:  name,
		cost:  cost,
	}
	return p, nil
} //factory.New()

func (f factory) GetUserProduct(user users.IUser, productName string) goods.IProduct {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cur, err := f.collection.Find(ctx, bson.M{"owner": user.ID(), "name": productName})
	if err != nil {
		log.Errorf("Failed to find user product: %v", err)
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

		p := mongoProduct{
			id:    result["id"].(string),
			owner: user,
			name:  result["name"].(string),
			cost:  wallets.Amount(result["cost"].(int32)),
		}
		return &p
	}

	if err := cur.Err(); err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}
	return nil
} //factory.GetUserProduct()

func (f factory) GetID(id string) goods.IProduct {
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

		userID := result["owner"].(string)
		user := f.users.GetID(userID)
		if user == nil {
			log.Errorf("Failed to get user from id=%s", result["owner"].(string))
			return nil
		}

		p := mongoProduct{
			id:    result["id"].(string),
			owner: user,
			name:  result["name"].(string),
			cost:  wallets.Amount(result["cost"].(int32)),
		}
		return &p
	}

	if err := cur.Err(); err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}
	return nil
} //factory.GetID()

func (f *factory) DelID(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := f.collection.FindOneAndDelete(ctx, bson.M{"id": id})
	if err != nil {
		log.Errorf("Failed to delete id: %v", err)
	}
}
