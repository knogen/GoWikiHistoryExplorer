package matchArxiv

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

type mongoRefDataBase struct {
	ctx            context.Context
	client         *mongo.Client
	collection_ref *mongo.Collection
}

func (c *mongoRefDataBase) initIndex() {
	mods := []mongo.IndexModel{
		{Keys: bson.M{"ref.doi": 1}},
		{Keys: bson.M{"ref.title": 1}},
		{Keys: bson.M{"match.magID": 1}},
		{Keys: bson.M{"match.mode": 1}},
	}
	_, err := c.collection_ref.Indexes().CreateMany(c.ctx, mods)
	if err != nil {
		log.Warn("collection_users index err:", err)
	}
}

func newMongoRefDataBase(MongoUri string) *mongoRefDataBase {
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoUri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client.Connect(ctx)
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database("wikipedia_ref")
	ctx = context.Background()

	mongodb := mongoRefDataBase{
		ctx:            ctx,
		client:         client,
		collection_ref: database.Collection("history_ref_20220201_v2_arxiv"),
	}
	mongodb.initIndex()
	return &mongodb
}

// 在 pages_articles 更新后， 对比 schedule_revision_info 和 pages_articles，找出过时了的 schedule_revision_info
func (c *mongoRefDataBase) GetUnlink() chan wikipediaRefObj {
	// 忽略重定向的 page
	cur, err := c.collection_ref.Find(c.ctx, bson.M{"match.mode": 0})
	failOnError(err, "查询失败了")

	outChan := make(chan wikipediaRefObj, 10000000)
	go func() {
		ct := 0
		for cur.Next(c.ctx) {
			ct += 1
			var doc wikipediaRefObj
			err := cur.Decode(&doc)
			failOnError(err, "decode 失败了")
			outChan <- doc
			if ct%100000 == 0 {
				log.Info("current count:", ct)
			}
		}
		close(outChan)
	}()

	return outChan
}

func (c *mongoRefDataBase) UpdateMatch(ID string, data refMatch) {

	_, err := c.collection_ref.UpdateOne(c.ctx, bson.M{"_id": ID}, bson.M{"$set": bson.M{"match": data}})
	if err != nil {
		log.Warn("wait 1s,update one err:", err)
		<-time.After(time.Second)
	}
}
