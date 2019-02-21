package returners

import (
	"../config"
	"../utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	mongoInstance *mongo.Client
	mongoOnce     sync.Once
)

func getMongoConnectArgs(opts *config.MasterOptions) string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
		opts.Returner.Mongo.User,
		opts.Returner.Mongo.Passwd,
		opts.Returner.Mongo.Ip, opts.Returner.Mongo.Port,
		opts.Returner.Mongo.DB,
		opts.Returner.Mongo.AuthSource)
}

func MongoConnect(opts *config.MasterOptions) *mongo.Client {
	created := false
	mongoOnce.Do(func() {
		connUri := getMongoConnectArgs(opts)
		client, err := mongo.NewClient(connUri)
		if !utils.CheckError(err) {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = client.Connect(ctx)
			mongoInstance = client
			created = true
		}

	})
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	if err := mongoInstance.Ping(ctx, readpref.Primary()); err != nil {
		log.Debug("db instance lost, try to reconnect")
		connUri := getMongoConnectArgs(opts)
		client, err := mongo.NewClient(connUri)
		if !utils.CheckError(err) {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = client.Connect(ctx)
			mongoInstance = client
			created = true
		}
	}

	if created {
		log.Debug("create a db instance")

	} else {
		log.Debug("reuse a db instance")
	}
	return mongoInstance
}

func UpdateTask(
	opts *config.MasterOptions,
	task *utils.Task,
	startTime int64, endTime int64, isFinished bool,
	status int, upsert bool) {
	taskKwargs, err := json.Marshal(&task)
	if !utils.CheckError(err) {
		db := MongoConnect(opts)
		collection := db.Database(opts.Returner.Mongo.DB).Collection("task_record")
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := collection.UpdateMany(ctx,
			bson.M{"task_instance_id": task.InstanceID},
			bson.D{{"$set",
				bson.M{
					"task_id": task.ID, "project_id": task.ProjectID,
					"task_name": task.Name, "task_kwargs": taskKwargs, "operator": task.Operator,
					"is_schedule": task.IsSchedule, "start_time": startTime,
					"end_time": endTime, "time_consuming": endTime - startTime, "is_finished": isFinished,
					"status": status,
				}}},
			&options.UpdateOptions{Upsert: &upsert})
		utils.CheckError(err)
	}
}

func UpdateStep(
	opts *config.MasterOptions, taskInstanceID string, jid string,
	step *utils.Step, startTime int64, endTime int64, isFinished bool,
	status int, upsert bool) {
	db := MongoConnect(opts)
	collection := db.Database(opts.Returner.Mongo.DB).Collection("step_record")
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	_, err := collection.UpdateMany(ctx,
		bson.M{"jid": jid},
		bson.D{{"$set",
			bson.M{
				"task_instance_id": taskInstanceID, "jid": jid, "step_id": step.ID,
				"account": step.Account, "type": step.Type, "name": step.Name,
				"block_name": step.BlockName, "description": step.Text, "start_time": startTime,
				"end_time": endTime, "time_consuming": endTime - startTime, "is_finished": isFinished,
				"status": status,
			}}, {"$inc", bson.M{"retry_count": 1}}},
		&options.UpdateOptions{Upsert: &upsert})
	utils.CheckError(err)
}

func UpdateMinion(opts *config.MasterOptions, events []*utils.Event, upsert bool) {
	db := MongoConnect(opts)
	collection := db.Database(opts.Returner.Mongo.DB).Collection("minion_result")
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	for _, event := range events {
		cursor, err := collection.Aggregate(ctx,
			[]bson.M{
				{"$match": bson.M{"jid": event.JID, "minion_id": event.MinionId}},
				{"$project": bson.D{
					{"result", bson.D{{"$ifNull", []string{"$result", ""}}}},
					{"job_type", 1},
					{"retcode", 1},
				}}},
		)
		if !utils.CheckError(err) {
			doc := bsonx.Doc{}
			cursor.Next(context.Background())
			cursor.Decode(&doc)
			doc = doc.Set("retcode", bsonx.Int32(int32(event.Retcode)))

			if doc.LookupElement("job_type").Value.NullOK() {
				doc = doc.Append("job_type", bsonx.Int32(int32(event.JobType)))
			}
			if doc.LookupElement("start_time").Value.NullOK() {
				doc = doc.Append("start_time", bsonx.Int32(int32(event.StartTime)))
			}
			doc = doc.Append("end_time", bsonx.Int32(int32(event.EndTime)))

			log.Debug(doc)
			doc.Set("result",
				bsonx.String(fmt.Sprintf("%v%s", doc.LookupElement("result").Value, event.Result)))
			_, err = collection.UpdateOne(ctx,
				bson.M{"minion_id": event.MinionId, "jid": event.JID},
				bson.D{{"$set", doc}}, &options.UpdateOptions{Upsert: &upsert})
			utils.CheckError(err)
		}
	}
}
