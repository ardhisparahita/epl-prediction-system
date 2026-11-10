
from dataclasses import replace
from types import SimpleNamespace

import numpy as np
import pandas as pd
import pytest

from app.ml.model_loader import load_artifacts
from app.services.feature_service import FeatureService
from app.services.prediction_service import PredictionService


@pytest.fixture(scope="module")
def artifacts():
    """Load production model artifacts once."""
    return load_artifacts()


@pytest.fixture(scope="module")
def prediction_service(artifacts):
    """Create PredictionService using production models."""

    feature_service = FeatureService(
        feature_columns=artifacts.feature_columns,
        history_window=5,
        minimum_history_matches=3,
    )

    return PredictionService(
        artifacts=artifacts,
        feature_service=feature_service,
    )


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
        "Date": date,
        "HomeTeam": home_team,
        "AwayTeam": away_team,
        "FTHG": home_goals,
        "FTAG": away_goals,
        "HST": home_sot,
        "AST": away_sot,
    }


@pytest.fixture
def historical_matches():
    """Synthetic historical matches for Arsenal and Chelsea."""

    return [
        # Arsenal: home and away matches.
        make_match("2026-01-01", "Arsenal", "Opponent 1", 2, 0, 5, 2),
        make_match("2026-01-05", "Opponent 2", "Arsenal", 1, 1, 4, 6),
        make_match("2026-01-10", "Arsenal", "Opponent 3", 0, 2, 3, 7),
        make_match("2026-01-15", "Opponent 4", "Arsenal", 0, 3, 2, 8),
        make_match("2026-01-20", "Arsenal", "Opponent 5", 1, 0, 4, 1),

        # Chelsea: home and away matches.
        make_match("2026-01-02", "Opponent 6", "Chelsea", 1, 2, 2, 3),
        make_match("2026-01-06", "Chelsea", "Opponent 7", 1, 1, 4, 4),
        make_match("2026-01-11", "Opponent 8", "Chelsea", 2, 0, 5, 2),
        make_match("2026-01-16", "Chelsea", "Opponent 9", 3, 1, 7, 5),
        make_match("2026-01-21", "Opponent 10", "Chelsea", 0, 1, 1, 6),
    ]


class FakeModel:
    """Lightweight model for testing probability validation."""

    def __init__(self, classes, probabilities):
        self.named_steps = {
            "classifier": SimpleNamespace(
                classes_=np.asarray(classes)
            )
        }
        self.probabilities = np.asarray(probabilities, dtype=float)

    def predict_proba(self, features):
        return self.probabilities


def test_predict_fixture_returns_expected_response(
    prediction_service,
    historical_matches,
):
    result = prediction_service.predict_fixture(
        home_team="Arsenal",
        away_team="Chelsea",
        kickoff_date="2026-02-01",
        historical_matches=historical_matches,
    )

    assert result["home_team"] == "Arsenal"
    assert result["away_team"] == "Chelsea"
    assert result["kickoff_date"] == "2026-02-01"

    assert result["feature_count"] == 12
    assert result["history_window"] == 5
    assert result["minimum_history_matches"] == 3

    assert result["model_version"] == (
        "epl-result-live-2026-2027-v1"
    )

    assert "logistic_regression" in result["probabilities"]
    assert "random_forest" in result["probabilities"]


def test_prediction_probabilities_have_correct_class_order(
    prediction_service,
    historical_matches,
):
    result = prediction_service.predict_fixture(
        home_team="Arsenal",
        away_team="Chelsea",
        kickoff_date="2026-02-01",
        historical_matches=historical_matches,
    )

    expected_classes = ["H", "D", "A"]

    for model_name in ["logistic_regression", "random_forest"]:
        probabilities = result["probabilities"][model_name]

        assert list(probabilities.keys()) == expected_classes


def test_prediction_probabilities_sum_to_one(
    prediction_service,
    historical_matches,
):
    result = prediction_service.predict_fixture(
        home_team="Arsenal",
        away_team="Chelsea",
        kickoff_date="2026-02-01",
        historical_matches=historical_matches,
    )

    for model_name in ["logistic_regression", "random_forest"]:
        probabilities = result["probabilities"][model_name]

        assert sum(probabilities.values()) == pytest.approx(
            1.0,
            abs=1e-5,
        )

        assert all(
            0.0 <= value <= 1.0
            for value in probabilities.values()
        )


def test_probabilities_are_mapped_from_model_class_order(
    prediction_service,
):
    model = FakeModel(
        classes=["A", "D", "H"],
        probabilities=[[0.2, 0.3, 0.5]],
    )

    result = prediction_service._predict_proba_in_class_order(
        model=model,
        features=pd.DataFrame([[0.0, 0.0, 0.0]]),
        model_name="Test Model",
    )

    assert list(result.keys()) == ["H", "D", "A"]

    assert result == {
        "H": 0.5,
        "D": 0.3,
        "A": 0.2,
    }


def test_rejects_unregistered_home_team(
    prediction_service,
    historical_matches,
):
    with pytest.raises(
        ValueError,
        match="Home team is not registered",
    ):
        prediction_service.predict_fixture(
            home_team="Unknown Team",
            away_team="Chelsea",
            kickoff_date="2026-02-01",
            historical_matches=historical_matches,
        )


def test_rejects_unregistered_away_team(
    prediction_service,
    historical_matches,
):
    with pytest.raises(
        ValueError,
        match="Away team is not registered",
    ):
        prediction_service.predict_fixture(
            home_team="Arsenal",
            away_team="Unknown Team",
            kickoff_date="2026-02-01",
            historical_matches=historical_matches,
        )


def test_rejects_same_home_and_away_team(
    prediction_service,
    historical_matches,
):
    with pytest.raises(
        ValueError,
        match="Home team and away team must be different",
    ):
        prediction_service.predict_fixture(
            home_team="Arsenal",
            away_team="Arsenal",
            kickoff_date="2026-02-01",
            historical_matches=historical_matches,
        )


def test_rejects_mismatched_feature_columns(artifacts):
    class WrongFeatureService:
        history_window = 5
        minimum_history_matches = 3

        def build_fixture_feature_row(self, **kwargs):
            return pd.DataFrame(
                [[0.0] * 12],
                columns=[f"wrong_feature_{i}" for i in range(12)],
            )

    service = PredictionService(
        artifacts=artifacts,
        feature_service=WrongFeatureService(),
    )

    with pytest.raises(
        ValueError,
        match="Feature columns do not match production model",
    ):
        service.predict_fixture(
            home_team="Arsenal",
            away_team="Chelsea",
            kickoff_date="2026-02-01",
            historical_matches=[],
        )


def test_rejects_unexpected_model_classes(
    prediction_service,
):
    model = FakeModel(
        classes=["H", "D", "X"],
        probabilities=[[0.5, 0.3, 0.2]],
    )

    with pytest.raises(
        ValueError,
        match="has unexpected classes",
    ):
        prediction_service._predict_proba_in_class_order(
            model=model,
            features=pd.DataFrame([[0.0]]),
            model_name="Test Model",
        )


def test_rejects_invalid_probability_shape(
    prediction_service,
):
    model = FakeModel(
        classes=["A", "D", "H"],
        probabilities=[[0.5, 0.5]],
    )

    with pytest.raises(
        ValueError,
        match="Unexpected probability shape",
    ):
        prediction_service._predict_proba_in_class_order(
            model=model,
            features=pd.DataFrame([[0.0]]),
            model_name="Test Model",
        )


def test_rejects_non_finite_probabilities(
    prediction_service,
):
    model = FakeModel(
        classes=["A", "D", "H"],
        probabilities=[[0.2, np.nan, 0.8]],
    )

    with pytest.raises(
        ValueError,
        match="non-finite probabilities",
    ):
        prediction_service._predict_proba_in_class_order(
            model=model,
            features=pd.DataFrame([[0.0]]),
            model_name="Test Model",
        )


def test_rejects_invalid_probability_sum(
    prediction_service,
):
    model = FakeModel(
        classes=["A", "D", "H"],
        probabilities=[[0.1, 0.2, 0.3]],
    )

    with pytest.raises(
        ValueError,
        match="probability sum is invalid",
    ):
        prediction_service._predict_proba_in_class_order(
            model=model,
            features=pd.DataFrame([[0.0]]),
            model_name="Test Model",
        )


def test_rejects_unexpected_metadata_class_order(artifacts):
    invalid_metadata = {
        **artifacts.metadata,
        "target_classes": ["A", "D", "H"],
    }

    invalid_artifacts = replace(
        artifacts,
        metadata=invalid_metadata,
    )

    feature_service = FeatureService(
        feature_columns=artifacts.feature_columns,
        history_window=5,
        minimum_history_matches=3,
    )

    with pytest.raises(
        ValueError,
        match="Unexpected class order",
    ):
        PredictionService(
            artifacts=invalid_artifacts,
            feature_service=feature_service,
        )
