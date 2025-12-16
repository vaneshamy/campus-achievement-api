package model

// ===== AGGREGATION RESULT MODELS =====

type TypeCount struct {
	Type  string `bson:"_id" json:"type"`
	Count int    `bson:"count" json:"count"`
}

type KeyCount struct {
	Key   string `bson:"_id" json:"key"`
	Count int    `bson:"count" json:"count"`
}

type StudentPoints struct {
	StudentID string `bson:"_id" json:"studentId"`
	Points    int    `bson:"points" json:"points"`
}

// ===== RAW RESULT FROM REPOSITORY =====

type StatusCountRow struct {
	Status string
	Count  int
}

type StudentCountRow struct {
	StudentID string
	Total     int
}

// ===== API RESPONSE =====

type ReportStatisticsResponse struct {
	ByStatus   map[string]int `json:"by_status"`
	TopStudent map[string]int `json:"top_student"`
	ByMonth    map[string]int `json:"by_month"`
}
