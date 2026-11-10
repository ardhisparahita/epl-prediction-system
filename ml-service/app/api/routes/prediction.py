from fastapi import APIRouter, HTTPException, Request
import pandas as pd

from app.schemas.prediction import (
    PredictionRequest,
    PredictionResponse,
)

router = APIRouter(tags=["Predictions"])


@router.post(
    "",
    response_model=PredictionResponse,
    summary="Predict football match result",
)
def predict_match(
    payload: PredictionRequest,
    request: Request,
):
    prediction_service = request.app.state.prediction_service

    # Konversi request menjadi format yang digunakan Feature Service.
    historical_matches = pd.DataFrame([
        {
            "Date": match.date,
            "HomeTeam": match.home_team,
            "AwayTeam": match.away_team,
            "FTHG": match.home_goals,
            "FTAG": match.away_goals,
            "HST": match.home_shots_on_target,
            "AST": match.away_shots_on_target,
        }
        for match in payload.historical_matches
    ])

    try:
        result = prediction_service.predict_fixture(
            home_team=payload.home_team,
            away_team=payload.away_team,
            kickoff_date=str(payload.kickoff_date),
            historical_matches=historical_matches,
        )

        return result

    except ValueError as exc:
        raise HTTPException(
            status_code=422,
            detail=str(exc),
        ) from exc