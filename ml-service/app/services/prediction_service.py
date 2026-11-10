
from __future__ import annotations

from collections.abc import Mapping, Sequence
from typing import Any

import numpy as np
import pandas as pd
from sklearn.pipeline import Pipeline

from app.ml.model_loader import ModelArtifacts
from app.services.feature_service import FeatureService


class PredictionService:
    """
    Menghasilkan probabilitas hasil pertandingan menggunakan
    Logistic Regression dan Random Forest.
    """

    def __init__(
        self,
        artifacts: ModelArtifacts,
        feature_service: FeatureService,
    ) -> None:
        self.artifacts = artifacts
        self.feature_service = feature_service

        self.class_order = artifacts.metadata["target_classes"]
        self.active_teams = set(artifacts.teams_active)

        self.model_lr = artifacts.model_lr
        self.model_rf = artifacts.model_rf

        if self.class_order != ["H", "D", "A"]:
            raise ValueError(
                f"Unexpected class order: {self.class_order}"
            )

    def _predict_proba_in_class_order(
        self,
        model: Pipeline,
        features: pd.DataFrame,
        model_name: str,
    ) -> dict[str, float]:
        """
        Menghasilkan probabilitas dengan urutan kelas H, D, A.

        Urutan kelas diambil langsung dari classifier.classes_,
        karena scikit-learn dapat menyimpan kelas dalam urutan
        alfabetis A, D, H.
        """

        classifier = model.named_steps["classifier"]

        model_classes = [
            str(value)
            for value in classifier.classes_
        ]

        if set(model_classes) != set(self.class_order):
            raise ValueError(
                f"{model_name} has unexpected classes: {model_classes}"
            )

        raw_probabilities = np.asarray(
            model.predict_proba(features),
            dtype=float,
        )

        if raw_probabilities.shape != (1, len(model_classes)):
            raise ValueError(
                f"Unexpected probability shape from {model_name}: "
                f"{raw_probabilities.shape}"
            )

        if not np.isfinite(raw_probabilities).all():
            raise ValueError(
                f"{model_name} returned non-finite probabilities"
            )

        probabilities_by_class = dict(
            zip(model_classes, raw_probabilities[0])
        )

        total_probability = sum(probabilities_by_class.values())

        if not np.isclose(
            total_probability,
            1.0,
            atol=1e-6,
        ):
            raise ValueError(
                f"{model_name} probability sum is invalid: "
                f"{total_probability}"
            )

        return {
            class_name: round(
                float(probabilities_by_class[class_name]),
                6,
            )
            for class_name in self.class_order
        }

    def predict_fixture(
        self,
        home_team: str,
        away_team: str,
        kickoff_date: str,
        historical_matches: (
            pd.DataFrame
            | Sequence[Mapping[str, Any]]
        ),
    ) -> dict[str, Any]:
        """
        Memprediksi satu pertandingan berdasarkan riwayat historis.
        """

        if home_team not in self.active_teams:
            raise ValueError(
                f"Home team is not registered: {home_team}"
            )

        if away_team not in self.active_teams:
            raise ValueError(
                f"Away team is not registered: {away_team}"
            )

        if home_team == away_team:
            raise ValueError(
                "Home team and away team must be different"
            )

        # Membentuk 12 fitur pertandingan.
        features = self.feature_service.build_fixture_feature_row(
            home_team=home_team,
            away_team=away_team,
            kickoff_date=kickoff_date,
            historical_matches=historical_matches,
        )

        # Memastikan urutan fitur sesuai dengan model.
        if list(features.columns) != self.artifacts.feature_columns:
            raise ValueError(
                "Feature columns do not match production model"
            )

        # Prediksi Logistic Regression.
        lr_probabilities = self._predict_proba_in_class_order(
            model=self.model_lr,
            features=features,
            model_name="Logistic Regression",
        )

        # Prediksi Random Forest.
        rf_probabilities = self._predict_proba_in_class_order(
            model=self.model_rf,
            features=features,
            model_name="Random Forest",
        )

        return {
            "home_team": home_team,
            "away_team": away_team,
            "kickoff_date": str(
                pd.Timestamp(kickoff_date).date()
            ),
            "feature_count": len(self.artifacts.feature_columns),
            "history_window": self.feature_service.history_window,
            "minimum_history_matches": (
                self.feature_service.minimum_history_matches
            ),
            "probabilities": {
                "logistic_regression": lr_probabilities,
                "random_forest": rf_probabilities,
            },
            "model_version": self.artifacts.metadata["model_version"],
        }
