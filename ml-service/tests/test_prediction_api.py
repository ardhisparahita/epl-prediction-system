
from copy import deepcopy

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.main import app
from app.api.routes.health import router as health_router


@pytest.fixture(scope="module")
def client():
    """
    Create TestClient and trigger the application lifespan.

    Production model artifacts are loaded when the client starts.
    """
    with TestClient(app) as test_client:
        yield test_client


def make_match(
    date,
    home_team,
    away_team,
    home_goals,
    away_goals,
    home_sot=None,
    away_sot=None,
):
    return {
        "date": date,
        "home_team": home_team,
        "away_team": away_team,
        "home_goals": home_goals,
        "away_goals": away_goals,
        "home_shots_on_target": home_sot,
        "away_shots_on_target": away_sot,
    }


@pytest.fixture
def valid_prediction_payload():
    """
    Valid prediction request for Arsenal vs Chelsea.
    Contains five previous matches for each team.
    """
    return {
        "home_team": "Arsenal",
        "away_team": "Chelsea",
        "kickoff_date": "2026-02-01",
        "historical_matches": [
            # Arsenal history.
            make_match(
                "2026-01-01", "Arsenal", "Opponent 1",
                2, 0, 5, 2,
            ),
            make_match(
                "2026-01-05", "Opponent 2", "Arsenal",
                1, 1, 4, 6,
            ),
            make_match(
                "2026-01-10", "Arsenal", "Opponent 3",
                0, 2, 3, 7,
            ),
            make_match(
                "2026-01-15", "Opponent 4", "Arsenal",
                0, 3, 2, 8,
            ),
            make_match(
                "2026-01-20", "Arsenal", "Opponent 5",
                1, 0, 4, 1,
            ),

            # Chelsea history.
            make_match(
                "2026-01-02", "Opponent 6", "Chelsea",
                1, 2, 2, 3,
            ),
            make_match(
                "2026-01-06", "Chelsea", "Opponent 7",
                1, 1, 4, 4,
            ),
            make_match(
                "2026-01-11", "Opponent 8", "Chelsea",
                2, 0, 5, 2,
            ),
            make_match(
                "2026-01-16", "Chelsea", "Opponent 9",
                3, 1, 7, 5,
            ),
            make_match(
                "2026-01-21", "Opponent 10", "Chelsea",
                0, 1, 1, 6,
            ),
        ],
    }


def test_prediction_endpoint_returns_successful_response(
    client,
    valid_prediction_payload,
):
    response = client.post(
        "/api/v1/predictions",
        json=valid_prediction_payload,
    )

    assert response.status_code == 200, response.text

    data = response.json()

    assert data["home_team"] == "Arsenal"
    assert data["away_team"] == "Chelsea"
    assert data["kickoff_date"] == "2026-02-01"

    assert data["feature_count"] == 12
    assert data["history_window"] == 5
    assert data["minimum_history_matches"] == 3

    assert data["model_version"] == (
        "epl-result-live-2026-2027-v1"
    )

    assert "logistic_regression" in data["probabilities"]
    assert "random_forest" in data["probabilities"]

    for model_name in ["logistic_regression", "random_forest"]:
        probabilities = data["probabilities"][model_name]

        assert set(probabilities.keys()) == {"H", "D", "A"}

        assert sum(probabilities.values()) == pytest.approx(
            1.0,
            abs=1e-5,
        )


def test_prediction_rejects_unregistered_home_team(
    client,
    valid_prediction_payload,
):
    payload = deepcopy(valid_prediction_payload)
    payload["home_team"] = "Unknown Team"

    response = client.post(
        "/api/v1/predictions",
        json=payload,
    )

    assert response.status_code == 422
    assert "Home team is not registered" in response.json()["detail"]


def test_prediction_rejects_unregistered_away_team(
    client,
    valid_prediction_payload,
):
    payload = deepcopy(valid_prediction_payload)
    payload["away_team"] = "Unknown Team"

    response = client.post(
        "/api/v1/predictions",
        json=payload,
    )

    assert response.status_code == 422
    assert "Away team is not registered" in response.json()["detail"]


def test_prediction_rejects_same_home_and_away_team(
    client,
    valid_prediction_payload,
):
    payload = deepcopy(valid_prediction_payload)
    payload["away_team"] = payload["home_team"]

    response = client.post(
        "/api/v1/predictions",
        json=payload,
    )

    assert response.status_code == 422

    error_text = str(response.json()["detail"])

    assert "Home team and away team must be different" in error_text


def test_prediction_rejects_insufficient_history(
    client,
    valid_prediction_payload,
):
    payload = deepcopy(valid_prediction_payload)

    arsenal_matches = [
        match
        for match in payload["historical_matches"]
        if (
            match["home_team"] == "Arsenal"
            or match["away_team"] == "Arsenal"
        )
    ][:2]

    chelsea_matches = [
        match
        for match in payload["historical_matches"]
        if (
            match["home_team"] == "Chelsea"
            or match["away_team"] == "Chelsea"
        )
    ]

    payload["historical_matches"] = (
        arsenal_matches + chelsea_matches
    )

    response = client.post(
        "/api/v1/predictions",
        json=payload,
    )

    assert response.status_code == 422
    assert "Riwayat Arsenal hanya 2 pertandingan" in (
        response.json()["detail"]
    )


def test_prediction_rejects_empty_historical_matches(
    client,
    valid_prediction_payload,
):
    payload = deepcopy(valid_prediction_payload)
    payload["historical_matches"] = []

    response = client.post(
        "/api/v1/predictions",
        json=payload,
    )

    assert response.status_code == 422


def test_prediction_rejects_negative_goals(
    client,
    valid_prediction_payload,
):
    payload = deepcopy(valid_prediction_payload)
    payload["historical_matches"][0]["home_goals"] = -1

    response = client.post(
        "/api/v1/predictions",
        json=payload,
    )

    assert response.status_code == 422


def test_health_endpoint_returns_success(client):
    response = client.get("/health")

    assert response.status_code == 200

    data = response.json()

    assert data["status"] == "ok"
    assert data["service"] == "epl-ml-service"
    assert data["model_loaded"] is True

    assert data["model_version"] == (
        "epl-result-live-2026-2027-v1"
    )


def test_health_endpoint_returns_503_when_not_initialized():
    """
    Test health endpoint without initializing PredictionService.
    """
    uninitialized_app = FastAPI()
    uninitialized_app.include_router(health_router)

    with TestClient(uninitialized_app) as test_client:
        response = test_client.get("/health")

    assert response.status_code == 503

    assert response.json()["detail"] == (
        "Prediction service is not initialized"
    )
