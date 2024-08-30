package model

import (
	"context"
	"playGround/config"
	"playGround/pbs"
	"playGround/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Article struct {
	pbs.Article
	Model
}

var ArticleColl *mongo.Collection //集合

func init() {
	ArticleColl = config.Db.Collection("article")
}

func (a *Article) Create(data *pbs.Article) error {
	data.Id = primitive.NewObjectID().Hex()
	data.CreatedAt = time.Now().Unix()
	return a.SetColl(ArticleColl).Add(data)
}

func (a *Article) GetArticleList(page, pageSize int64) (rs []*pbs.Article, count int64, err error) {
	var filter = bson.M{"deleted_at": 0}
	opt := &options.FindOptions{}
	if page != -1 && page > 0 { //page等于-1时不分页
		var offset = (page - 1) * pageSize
		opt.SetLimit(pageSize)
		opt.SetSkip(offset)
	}
	opt.SetSort(bson.M{"_id": -1})
	count, _ = ArticleColl.CountDocuments(Context, filter)
	query, err := ArticleColl.Find(Context, filter, opt)
	if err != nil {
		return
	}
	err = query.All(Context, &rs)
	return
}

func (a *Article) View(articleId string) (rs *pbs.Article, err error) {
	rs = new(pbs.Article)
	err = ArticleColl.FindOne(context.Background(), bson.M{"_id": articleId}).Decode(&rs)
	return
}

func (c *Article) Edit(data *pbs.Article) error {
	data.UpdatedAt = time.Now().Unix()
	update := utils.Struct2Map(*data)
	delete(update, "_id")
	delete(update, "created_at") //不能修改ID
	delete(update, "_id,omitempty")
	_, err := ArticleColl.UpdateOne(Context, bson.M{"_id": data.Id}, bson.M{"$set": update})
	return err
}
