
from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import joblib
from sklearn.pipeline import Pipeline

BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_ARTIFACTS_DIR = BASE_DIR / "artifacts"

EXPECTED_CLASSES = ["H", "D", "A"]
EXPECTED_FEATURE_COUNT = 12

@dataclass(frozen=True)
class ModelArtifacts:
    model_lr: Pipeline
    model_rf: Pipeline
    feature_columns: list[str]
    metadata: dict[str, Any]
    teams_active: list[str]

def read_json(path: Path) -> Any:
    if not path.is_file():
        raise FileNotFoundError(f"JSON file not found: {path}")

    with path.open("r", encoding="utf-8") as file:
        return json.load(file)

def validate_model(
    model: Pipeline,
    model_name: str,
    feature_count: int,
) -> None:
    if not isinstance(model, Pipeline):
        raise TypeError(
            f"{model_name} must be an sklearn Pipeline"
        )

    if "classifier" not in model.named_steps:
        raise ValueError(
            f"{model_name} does not contain a classifier step"
        )

    classifier = model.named_steps["classifier"]

    actual_feature_count = getattr(
        classifier, "n_features_in_", None
    )

    if actual_feature_count != feature_count:
        raise ValueError(
            f"{model_name} expected {feature_count} features, "
            f"got {actual_feature_count}"
        )

    model_classes = getattr(classifier, "classes_", None)

    if model_classes is None:
        raise ValueError(
            f"{model_name} does not expose classes_"
        )

    actual_classes = {str(value) for value in model_classes}
    expected_classes = set(EXPECTED_CLASSES)

    if actual_classes != expected_classes:
        raise ValueError(
            f"{model_name} has unexpected classes: "
            f"{actual_classes}"
        )

def load_artifacts(
    artifacts_dir: Path | None = None,
) -> ModelArtifacts:
    directory = (
        Path(artifacts_dir)
        if artifacts_dir is not None
        else DEFAULT_ARTIFACTS_DIR
    )

    required_files = [
        "model_lr.joblib",
        "model_rf.joblib",
        "feature_columns.json",
        "model_metadata.json",
        "teams_active.json",
    ]

    missing_files = [
        filename
        for filename in required_files
        if not (directory / filename).is_file()
    ]

    if missing_files:
        raise FileNotFoundError(
            f"Missing model artifacts in {directory}: "
            f"{', '.join(missing_files)}"
        )

    feature_columns = read_json(
        directory / "feature_columns.json"
    )

    metadata = read_json(
        directory / "model_metadata.json"
    )

    teams_active = read_json(
        directory / "teams_active.json"
    )

    if not isinstance(feature_columns, list):
        raise ValueError("feature_columns.json must contain a list")

    if len(feature_columns) != EXPECTED_FEATURE_COUNT:
        raise ValueError(
            f"Expected {EXPECTED_FEATURE_COUNT} features, "
            f"got {len(feature_columns)}"
        )

    if len(set(feature_columns)) != len(feature_columns):
        raise ValueError("Duplicate feature names detected")

    if metadata.get("feature_columns") != feature_columns:
        raise ValueError(
            "Feature columns do not match model metadata"
        )

    if metadata.get("feature_count") != len(feature_columns):
        raise ValueError(
            "Feature count does not match model metadata"
        )

    if metadata.get("target_classes") != EXPECTED_CLASSES:
        raise ValueError(
            f"Unexpected target classes: "
            f"{metadata.get('target_classes')}"
        )

    if metadata.get("model_is_frozen_for_live_season") is not True:
        raise ValueError(
            "Production model is not frozen for the live season"
        )

    if not isinstance(teams_active, list) or not teams_active:
        raise ValueError("teams_active.json must contain a non-empty list")

    model_lr = joblib.load(directory / "model_lr.joblib")
    model_rf = joblib.load(directory / "model_rf.joblib")

    validate_model(
        model_lr,
        "Logistic Regression",
        len(feature_columns),
    )

    validate_model(
        model_rf,
        "Random Forest",
        len(feature_columns),
    )

    return ModelArtifacts(
        model_lr=model_lr,
        model_rf=model_rf,
        feature_columns=feature_columns,
        metadata=metadata,
        teams_active=teams_active,
    )