package wuid

import (
	"context"
	"errors"
	"time"

	v2 "github.com/r2b89/wuid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// WUID is an extremely fast unique number generator.
type WUID struct {
	w *v2.WUID
}

// NewWUID creates a new WUID instance.
func NewWUID(name string, opts ...Option) *WUID {
	return &WUID{w: v2.NewWUID(name, opts...)}
}

// Next returns the next unique number.
func (this *WUID) Next() int64 {
	return this.w.Next()
}

type NewClient func() (client *mongo.Client, autoDisconnect bool, err error)

// LoadH28FromMongo adds 1 to a specific number in your MongoDB, fetches its new value,
// and then sets that as the high 28 bits of the unique numbers that Next generates.
func (this *WUID) LoadH28FromMongo(newClient NewClient, dbName, coll, docID string) error {
	if len(dbName) == 0 {
		return errors.New("dbName cannot be empty. name: " + this.w.Name)
	}
	if len(coll) == 0 {
		return errors.New("coll cannot be empty. name: " + this.w.Name)
	}
	if len(docID) == 0 {
		return errors.New("docID cannot be empty. name: " + this.w.Name)
	}

	client, autoDisconnect, err := newClient()
	if err != nil {
		return err
	}
	if autoDisconnect {
		defer func() {
			ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel2()
			_ = client.Disconnect(ctx2)
		}()
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel1()
	if err := client.Ping(ctx1, readpref.Primary()); err != nil {
		return err
	}

	collOpts := &options.CollectionOptions{
		ReadConcern:    readconcern.Majority(),
		WriteConcern:   writeconcern.New(writeconcern.WMajority()),
		ReadPreference: readpref.Primary(),
	}
	c := client.Database(dbName).Collection(coll, collOpts)

	filter := bson.D{{"_id", docID}}
	update := bson.D{{"$inc", bson.D{{"n", int32(1)}}}}
	var findOneAndUpdateOptions options.FindOneAndUpdateOptions
	findOneAndUpdateOptions.SetUpsert(true).SetReturnDocument(options.After)
	var doc struct {
		N int32
	}
	err = c.FindOneAndUpdate(ctx1, filter, update, &findOneAndUpdateOptions).Decode(&doc)
	if err != nil {
		return err
	}
	h28 := int64(doc.N)
	if err = this.w.VerifyH28(h28); err != nil {
		return err
	}

	this.w.Reset(h28 << 36)

	this.w.Lock()
	defer this.w.Unlock()

	if this.w.Renew != nil {
		return nil
	}
	this.w.Renew = func() error {
		return this.LoadH28FromMongo(newClient, dbName, coll, docID)
	}

	return nil
}

// RenewNow reacquires the high 28 bits from your data store immediately
func (this *WUID) RenewNow() error {
	return this.w.RenewNow()
}

type Option = v2.Option

// WithSection adds a section ID to the generated numbers. The section ID must be in between [0, 7].
func WithSection(section int8) Option {
	return v2.WithSection(section)
}

// WithH28Verifier sets your own h28 verifier
func WithH28Verifier(cb func(h28 int64) error) Option {
	return v2.WithH28Verifier(cb)
}

// WithStep sets the step and floor of Next()
func WithStep(step int64, floor int64) Option {
	return v2.WithStep(step, floor)
}
