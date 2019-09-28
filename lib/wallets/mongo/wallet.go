package mongo

import (
	"context"
	"math/rand"
	"time"

	"github.com/jansemmelink/log"
	"github.com/jansemmelink/taxiching/lib/users"
	"github.com/jansemmelink/taxiching/lib/wallets"
	"github.com/satori/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//e.g. install("mongodb://localhost:27017")
func Wallets(mongoURI string, dbName string, users users.IUsers) (wallets.IWallets, error) {
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

	collection := client.Database(dbName).Collection("wallets")
	return &factory{
		users:      users,
		collection: collection,
	}, nil
} //Wallets()

type factory struct {
	users      users.IUsers
	collection *mongo.Collection
}

func (f factory) New(u users.IUser, walletName string, minBalance wallets.Amount) (wallets.IWallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := uuid.NewV1().String()
	res, err := f.collection.InsertOne(
		ctx,
		bson.M{
			"id":         id,
			"owner":      u.ID(),
			"name":       walletName,
			"minBalance": minBalance,
		})
	if err != nil {
		return nil, log.Wrapf(err, "failed to insert wallet into db")
	}

	log.Debugf("inserted %T: %+v", res, res)

	w := &mongoWallet{
		id:         id,
		owner:      u,
		name:       walletName,
		minBalance: minBalance,
	}
	return w, nil
} //factory.New()

func (f *factory) UserWallet(userID string, walletName string) wallets.IWallet {
	if f == nil {
		return nil
	}
	return nil
} //factory.GetMsisdn()

func (f factory) GetID(id string) wallets.IWallet {
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

		w := mongoWallet{
			id:         result["id"].(string),
			owner:      user,
			name:       result["name"].(string),
			minBalance: wallets.Amount(result["minBalance"].(int32)),
		}
		return &w
	}
	if err := cur.Err(); err != nil {
		log.Errorf("Error: %v", err)
		return nil
	}
	return nil
} //factory.GetID()

func (f *factory) GetByDepRef(ref string) wallets.IWallet {
	return nil
} //factory.GetByDepRef()

type mongoWallet struct {
	id         string
	owner      users.IUser
	name       string
	depRef     string
	balance    wallets.Amount
	minBalance wallets.Amount
}

func (w mongoWallet) ID() string {
	return w.id
}

func (w mongoWallet) Owner() users.IUser {
	return w.owner
}

func (w mongoWallet) Name() string {
	return w.name
}

func (w mongoWallet) Balance() wallets.Amount {
	return w.balance
}

func (w mongoWallet) MinBalance() wallets.Amount {
	return w.minBalance
}

func (w *mongoWallet) Debit(amount wallets.Amount) {
	w.balance -= amount
}

func (w *mongoWallet) Credit(amount wallets.Amount) {
	w.balance += amount
}

func (w *mongoWallet) DepositReference() string {
	return w.depRef
}

//wallet id is considered secret, and its long and tedious to type
//this function allocates a unique deposit reference for a wallet
//that is easier to type
//to make it simple, the range is limited, and it should only be
//used when user can make EFT payments into the wallet
func (f *factory) NewDepRef(w wallets.IWallet) (string, error) {
	// f.depRefMutex.Lock()
	// defer f.depRefMutex.Unlock()
	for attempt := 0; attempt < 10; attempt++ {
		ref := "W-"
		ref += string('0' + rand.Intn(10))
		ref += string('A' + rand.Intn(26))
		ref += string('A' + rand.Intn(26))
		ref += string('A' + rand.Intn(26))
		ref += "-"
		ref += string('0' + rand.Intn(10))
		ref += string('A' + rand.Intn(26))
		ref += string('A' + rand.Intn(26))
		ref += string('A' + rand.Intn(26))

		return "", log.Wrapf(nil, "todo: dep ref not yet stored in db")
	} //for each attempt
	return "", log.Wrapf(nil, "Unable to generate deposit reference")
} //factory.NewDepRef()
