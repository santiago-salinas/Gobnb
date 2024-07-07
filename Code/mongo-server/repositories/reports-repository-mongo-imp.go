package repositories

import (
	"context"
	"mongo-server/mongo_models"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collection = "reports"
)

type reportsRepoMongoImp struct {
	mongoClient *mongo.Client
	database    string
}

func NewReportsMongoRepo(mongoClient *mongo.Client, database string) *reportsRepoMongoImp {
	return &reportsRepoMongoImp{
		mongoClient: mongoClient,
		database:    database,
	}
}

func (reportsRepo *reportsRepoMongoImp) GetLatestSensorReport(sensorId string) (mongo_models.SensorReport, error) {
	var report mongo_models.SensorReportDBO

	// Create a filter for the sensorId
	filter := bson.M{"sensorId": sensorId}

	// Sort by date in descending order to get the latest report first
	opts := options.FindOne().SetSort(bson.D{{"date", -1}})

	err := reportsRepo.mongoClient.Database(reportsRepo.database).
		Collection(collection).
		FindOne(context.TODO(), filter, opts).
		Decode(&report)

	if err != nil {
		return mongo_models.SensorReport{}, err
	}

	return report.ToObject(), nil
}

func (reportsRepo *reportsRepoMongoImp) AddAppReport(report mongo_models.AppReport) error {
	reportDBO := report.ToDBO()
	_, err := reportsRepo.mongoClient.Database(reportsRepo.database).Collection(collection).InsertOne(context.TODO(), reportDBO)
	return err
}

func (reportsRepo *reportsRepoMongoImp) AddSensorReport(report mongo_models.SensorReport) error {
	reportDBO := report.ToDBO()
	_, err := reportsRepo.mongoClient.Database(reportsRepo.database).Collection(collection).InsertOne(context.TODO(), reportDBO)
	return err
}

func (reportsRepo *reportsRepoMongoImp) GetAllReports() ([]mongo_models.AppReport, error) {
	cursorReports, err := reportsRepo.mongoClient.Database(reportsRepo.database).Collection(collection).Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	var reports []mongo_models.AppReport
	err = cursorReports.All(context.TODO(), &reports)
	return reports, err
}

func (reportsRepo *reportsRepoMongoImp) GetAllAppReports(startDate time.Time, endDate time.Time) ([]mongo_models.RankingReportItem, error) {
	filter := bson.D{
		{Key: "sensorId", Value: bson.D{{"$regex", "^APP"}}},
		{Key: "date", Value: bson.D{
			{Key: "$gte", Value: startDate},
			{Key: "$lte", Value: endDate},
		}},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "propertyId", Value: "$propertyId"},
					{Key: "type", Value: "$type"},
				}},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			}},
		},
		{
			{Key: "$project", Value: bson.D{
				{Key: "propertyId", Value: "$_id.propertyId"},
				{Key: "type", Value: "$_id.type"},
				{Key: "count", Value: 1},
			}},
		},
	}

	cursorReports, err := reportsRepo.mongoClient.Database(reportsRepo.database).Collection(collection).Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var problemItems []mongo_models.ProblemItem
	err = cursorReports.All(context.TODO(), &problemItems)

	propertyProblemMap := make(map[string][]mongo_models.ProblemItem)
	for _, problemItem := range problemItems {
		if _, ok := propertyProblemMap[problemItem.PropertyId]; !ok {
			propertyProblemMap[problemItem.PropertyId] = []mongo_models.ProblemItem{}
		}
		propertyProblemMap[problemItem.PropertyId] = append(propertyProblemMap[problemItem.PropertyId], problemItem)
	}

	rankingReportItems := []mongo_models.RankingReportItem{}
	for propertyId, problemItems := range propertyProblemMap {
		totalProblems := 0
		for _, problemItem := range problemItems {
			totalProblems += problemItem.Count
		}
		sort.Slice(problemItems, func(i, j int) bool {
			return problemItems[i].Count > problemItems[j].Count
		})
		problemsLength := len(problemItems)
		if problemsLength >= 2 {
			problemsLength = 2
		}
		twoMostFrequentProblems := problemItems[:problemsLength]
		rankingReportItem := mongo_models.RankingReportItem{
			Id:            propertyId,
			Problems:      twoMostFrequentProblems,
			TotalProblems: totalProblems,
		}
		rankingReportItems = append(rankingReportItems, rankingReportItem)
	}

	sort.Slice(rankingReportItems, func(i, j int) bool {
		return rankingReportItems[i].TotalProblems > rankingReportItems[j].TotalProblems
	})

	sliceSize := 15
	itemsLength := len(rankingReportItems)
	if itemsLength < 15 {
		sliceSize = itemsLength
	}

	return rankingReportItems[:sliceSize], err
}
