package dto

// MatchListInput represents match history query parameters.
type MatchListInput struct {
	Team  string
	From  string
	To    string
	Page  int
	Limit int
}

// MatchListItem represents one match in the API response.
type MatchListItem struct {
	ID                uint64 `json:"id"`
	Date              string `json:"date"`
	HomeTeam          string `json:"home_team"`
	AwayTeam          string `json:"away_team"`
	HomeGoals         int    `json:"home_goals"`
	AwayGoals         int    `json:"away_goals"`
	HomeShotsOnTarget *int   `json:"home_shots_on_target"`
	AwayShotsOnTarget *int   `json:"away_shots_on_target"`
}

// MatchPagination contains pagination metadata.
type MatchPagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// MatchListResponse represents the paginated match history response.
type MatchListResponse struct {
	Data       []MatchListItem `json:"data"`
	Pagination MatchPagination `json:"pagination"`
}
