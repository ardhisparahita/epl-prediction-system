
from types import SimpleNamespace

import pytest
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler

from app.ml.model_loader import (
    EXPECTED_CLASSES,
    EXPECTED_FEATURE_COUNT,
    ModelArtifacts,
    load_artifacts,
    read_json,
    validate_model,
)


EXPECTED_FEATURE_COLUMNS = [
    "H_gf_5",
    "H_ga_5",
    "H_pts_5",
    "H_sot_5",
    "A_gf_5",
    "A_ga_5",
    "A_pts_5",
    "A_sot_5",
    "diff_gf_5",
    "diff_ga_5",
    "diff_pts_5",
    "diff_sot_5",
]


@pytest.fixture(scope="module")
def loaded_artifacts():
    """Load production model artifacts once for this test module."""
    return load_artifacts()


def test_load_artifacts_returns_model_artifacts(loaded_artifacts):
    assert isinstance(loaded_artifacts, ModelArtifacts)

    assert isinstance(loaded_artifacts.model_lr, Pipeline)
    assert isinstance(loaded_artifacts.model_rf, Pipeline)


def test_models_have_valid_classifier_steps(loaded_artifacts):
    models = {
        "Logistic Regression": loaded_artifacts.model_lr,
        "Random Forest": loaded_artifacts.model_rf,
    }

    for model_name, model in models.items():
        assert "classifier" in model.named_steps, (
            f"{model_name} must contain a classifier step"
        )

        classifier = model.named_steps["classifier"]

        assert classifier.n_features_in_ == EXPECTED_FEATURE_COUNT
        assert set(map(str, classifier.classes_)) == set(EXPECTED_CLASSES)


def test_feature_columns_have_correct_order(loaded_artifacts):
    assert loaded_artifacts.feature_columns == EXPECTED_FEATURE_COLUMNS

    assert (
        loaded_artifacts.metadata["feature_columns"]
        == EXPECTED_FEATURE_COLUMNS
    )


def test_model_metadata_is_valid(loaded_artifacts):
    metadata = loaded_artifacts.metadata

    assert metadata["feature_count"] == EXPECTED_FEATURE_COUNT
    assert metadata["target_classes"] == EXPECTED_CLASSES

    assert metadata["model_is_frozen_for_live_season"] is True

    assert metadata["model_version"] == (
        "epl-result-live-2026-2027-v1"
    )


def test_active_teams_are_valid(loaded_artifacts):
    teams = loaded_artifacts.teams_active

    assert isinstance(teams, list)
    assert len(teams) == 20
    assert all(isinstance(team, str) and team.strip() for team in teams)
    assert len(teams) == len(set(teams))


def test_read_json_raises_error_when_file_not_found(tmp_path):
    missing_file = tmp_path / "missing.json"

    with pytest.raises(FileNotFoundError, match="JSON file not found"):
        read_json(missing_file)


def test_load_artifacts_raises_error_when_files_are_missing(tmp_path):
    with pytest.raises(
        FileNotFoundError,
        match="Missing model artifacts",
    ):
        load_artifacts(tmp_path)


def test_validate_model_rejects_non_pipeline():
    with pytest.raises(TypeError, match="must be an sklearn Pipeline"):
        validate_model(
            object(),
            "Test Model",
            EXPECTED_FEATURE_COUNT,
        )


def test_validate_model_rejects_pipeline_without_classifier():
    model = Pipeline([
        ("scaler", StandardScaler()),
    ])

    with pytest.raises(
        ValueError,
        match="does not contain a classifier step",
    ):
        validate_model(
            model,
            "Test Model",
            EXPECTED_FEATURE_COUNT,
        )


def test_validate_model_rejects_wrong_feature_count():
    classifier = SimpleNamespace(
        n_features_in_=11,
        classes_=["A", "D", "H"],
    )

    model = Pipeline([
        ("classifier", classifier),
    ])

    with pytest.raises(
        ValueError,
        match="expected 12 features",
    ):
        validate_model(
            model,
            "Test Model",
            EXPECTED_FEATURE_COUNT,
        )


def test_validate_model_rejects_unexpected_classes():
    classifier = SimpleNamespace(
        n_features_in_=EXPECTED_FEATURE_COUNT,
        classes_=["H", "D"],
    )

    model = Pipeline([
        ("classifier", classifier),
    ])

    with pytest.raises(
        ValueError,
        match="unexpected classes",
    ):
        validate_model(
            model,
            "Test Model",
            EXPECTED_FEATURE_COUNT,
        )
