package dto

// HistoricalMatch represents one completed historical match.
type HistoricalMatch struct {
	Date              string   `json:"date"`
	HomeTeam          string   `json:"home_team"`
	AwayTeam          string   `json:"away_team"`
	HomeGoals         int      `json:"home_goals"`
	AwayGoals         int      `json:"away_goals"`
	HomeShotsOnTarget *float64 `json:"home_shots_on_target"`
	AwayShotsOnTarget *float64 `json:"away_shots_on_target"`
}

// PredictionRequest represents the request sent to the ML Service.
type PredictionRequest struct {
	HomeTeam          string            `json:"home_team"`
	AwayTeam          string            `json:"away_team"`
	KickoffDate       string            `json:"kickoff_date"`
	HistoricalMatches []HistoricalMatch `json:"historical_matches"`
}

// ClassProbabilities contains probabilities for H, D, and A.
type ClassProbabilities struct {
	H float64 `json:"H"`
	D float64 `json:"D"`
	A float64 `json:"A"`
}

// PredictionProbabilities contains the output of both models.
type PredictionProbabilities struct {
	LogisticRegression ClassProbabilities `json:"logistic_regression"`
	RandomForest       ClassProbabilities `json:"random_forest"`
}

// PredictionResponse represents the response returned by ML Service.
type PredictionResponse struct {
	HomeTeam              string                  `json:"home_team"`
	AwayTeam              string                  `json:"away_team"`
	KickoffDate           string                  `json:"kickoff_date"`
	FeatureCount          int                     `json:"feature_count"`
	HistoryWindow         int                     `json:"history_window"`
	MinimumHistoryMatches int                     `json:"minimum_history_matches"`
	Probabilities         PredictionProbabilities `json:"probabilities"`
	ModelVersion          string                  `json:"model_version"`
}

// MLServiceError represents an error returned by FastAPI.
type MLServiceError struct {
	Detail any `json:"detail"`
}
