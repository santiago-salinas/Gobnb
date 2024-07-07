package mongo_models

type ProblemItem struct {
	PropertyId string `json:"propertyId" bson:"propertyId"`
	Problem string `json:"type" bson:"type"`
	Count   int    `json:"count" bson:"count"`
}

type RankingReportItem struct {
	Id           string        `json:"id" bson:"id"`
	Name         string        `json:"name" bson:"name"`
	Neighborhood string        `json:"neighborhood" bson:"neighborhood"`
	State        string        `json:"state" bson:"state"`
	Problems     []ProblemItem `json:"problems" bson:"problems"`
	TotalProblems int           `json:"totalProblems" bson:"totalProblems"`
}

type RankingReport struct {
	Items []RankingReportItem `json:"items" bson:"items"`
}
