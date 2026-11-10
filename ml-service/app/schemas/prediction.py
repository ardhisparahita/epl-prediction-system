
from datetime import date
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field, model_validator


class HistoricalMatchInput(BaseModel):
    model_config = ConfigDict(
        str_strip_whitespace=True,
        extra="forbid",
    )

    date: date
    home_team: str = Field(min_length=1)
    away_team: str = Field(min_length=1)

    home_goals: int = Field(ge=0)
    away_goals: int = Field(ge=0)

    home_shots_on_target: float | None = Field(
        default=None,
        ge=0,
    )
    away_shots_on_target: float | None = Field(
        default=None,
        ge=0,
    )


class PredictionRequest(BaseModel):
    model_config = ConfigDict(
        str_strip_whitespace=True,
        extra="forbid",
    )

    home_team: str = Field(min_length=1)
    away_team: str = Field(min_length=1)
    kickoff_date: date

    historical_matches: list[HistoricalMatchInput] = Field(
        min_length=1,
    )

    @model_validator(mode="after")
    def validate_teams(self):
        if self.home_team.casefold() == self.away_team.casefold():
            raise ValueError(
                "Home team and away team must be different"
            )
        return self


class ClassProbabilities(BaseModel):
    H: float = Field(ge=0, le=1)
    D: float = Field(ge=0, le=1)
    A: float = Field(ge=0, le=1)


class PredictionProbabilities(BaseModel):
    logistic_regression: ClassProbabilities
    random_forest: ClassProbabilities


class PredictionResponse(BaseModel):
    home_team: str
    away_team: str
    kickoff_date: date

    feature_count: int
    history_window: int
    minimum_history_matches: int

    probabilities: PredictionProbabilities

    model_version: str
