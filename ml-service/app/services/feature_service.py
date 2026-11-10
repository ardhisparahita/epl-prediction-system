
from __future__ import annotations

from collections.abc import Mapping, Sequence
from typing import Any

import numpy as np
import pandas as pd


class FeatureService:
    """
    Membentuk 12 fitur pertandingan berdasarkan maksimal lima
    pertandingan terakhir setiap tim.

    Riwayat mencakup pertandingan kandang dan tandang.
    Pertandingan pada tanggal yang sama dengan tanggal prediksi
    tidak boleh digunakan.
    """

    def __init__(
        self,
        feature_columns: list[str],
        history_window: int = 5,
        minimum_history_matches: int = 3,
    ) -> None:
        if history_window < 1:
            raise ValueError("history_window must be at least 1")

        if minimum_history_matches < 1:
            raise ValueError(
                "minimum_history_matches must be at least 1"
            )

        if minimum_history_matches > history_window:
            raise ValueError(
                "minimum_history_matches cannot exceed history_window"
            )

        if len(feature_columns) != 12:
            raise ValueError(
                f"Expected 12 feature columns, got {len(feature_columns)}"
            )

        self.feature_columns = feature_columns
        self.history_window = history_window
        self.minimum_history_matches = minimum_history_matches

    @staticmethod
    def _normalize_historical_matches(
        historical_matches: (
            pd.DataFrame
            | Sequence[Mapping[str, Any]]
        ),
    ) -> pd.DataFrame:
        """
        Validasi dan normalisasi data pertandingan historis.

        Pertandingan yang belum mempunyai skor akhir tidak digunakan.
        HST dan AST boleh kosong karena SOT menggunakan rata-rata
        dari nilai yang tersedia.
        """

        if isinstance(historical_matches, pd.DataFrame):
            history = historical_matches.copy()
        else:
            history = pd.DataFrame(historical_matches)

        # Periksa data kosong sebelum memeriksa kolom wajib.
        if history.empty:
            raise ValueError("Historical matches are empty")

        required_columns = [
            "Date",
            "HomeTeam",
            "AwayTeam",
            "FTHG",
            "FTAG",
        ]

        missing_columns = [
            column
            for column in required_columns
            if column not in history.columns
        ]

        if missing_columns:
            raise ValueError(
                f"Missing historical match columns: {missing_columns}"
            )

        # SOT merupakan data opsional.
        for column in ["HST", "AST"]:
            if column not in history.columns:
                history[column] = np.nan

        # Pertahankan urutan sumber jika ada beberapa pertandingan
        # pada tanggal yang sama.
        history["SourceOrder"] = np.arange(len(history))

        history["Date"] = pd.to_datetime(
            history["Date"],
            errors="coerce",
        ).dt.normalize()

        numeric_columns = ["FTHG", "FTAG", "HST", "AST"]

        for column in numeric_columns:
            history[column] = pd.to_numeric(
                history[column],
                errors="coerce",
            )

        # Hanya pertandingan dengan skor akhir yang valid.
        # Pertandingan yang belum selesai tidak boleh masuk riwayat.
        history = history.dropna(
            subset=[
                "Date",
                "HomeTeam",
                "AwayTeam",
                "FTHG",
                "FTAG",
            ]
        ).copy()

        # Data tersedia, tetapi tidak ada pertandingan selesai yang valid.
        if history.empty:
            raise ValueError(
                "No completed historical matches are available"
            )

        history = history.sort_values(
            ["Date", "SourceOrder"],
            kind="stable",
        ).reset_index(drop=True)

        return history

    def _get_team_history_before(
        self,
        team_name: str,
        cutoff_date: pd.Timestamp,
        historical_matches: pd.DataFrame,
    ) -> dict[str, float]:
        """
        Menghitung statistik tim berdasarkan seluruh pertandingan
        kandang dan tandang sebelum cutoff_date.
        """

        # Aturan anti-data-leakage:
        # Pertandingan pada tanggal yang sama tidak digunakan.
        prior_matches = historical_matches[
            historical_matches["Date"] < cutoff_date
        ]

        team_games: list[dict[str, Any]] = []

        # Riwayat ketika tim bermain sebagai tuan rumah.
        home_rows = prior_matches[
            prior_matches["HomeTeam"] == team_name
        ]

        for row in home_rows.itertuples(index=False):
            points = (
                3 if row.FTHG > row.FTAG
                else 1 if row.FTHG == row.FTAG
                else 0
            )

            team_games.append({
                "Date": row.Date,
                "SourceOrder": row.SourceOrder,
                "GF": float(row.FTHG),
                "GA": float(row.FTAG),
                "Points": float(points),
                "SOT": row.HST,
            })

        # Riwayat ketika tim bermain sebagai tim tamu.
        away_rows = prior_matches[
            prior_matches["AwayTeam"] == team_name
        ]

        for row in away_rows.itertuples(index=False):
            points = (
                3 if row.FTAG > row.FTHG
                else 1 if row.FTAG == row.FTHG
                else 0
            )

            team_games.append({
                "Date": row.Date,
                "SourceOrder": row.SourceOrder,
                "GF": float(row.FTAG),
                "GA": float(row.FTHG),
                "Points": float(points),
                "SOT": row.AST,
            })

        # Urutkan pertandingan secara kronologis.
        team_games.sort(
            key=lambda item: (
                item["Date"],
                item["SourceOrder"],
            )
        )

        # Ambil maksimal lima pertandingan terakhir.
        recent = team_games[-self.history_window:]

        if len(recent) < self.minimum_history_matches:
            raise ValueError(
                f"Riwayat {team_name} hanya {len(recent)} pertandingan "
                f"sebelum {cutoff_date.date()}; "
                f"minimal {self.minimum_history_matches} diperlukan."
            )

        gf_values = [item["GF"] for item in recent]
        ga_values = [item["GA"] for item in recent]
        points_values = [item["Points"] for item in recent]

        # Nilai SOT yang kosong diabaikan.
        valid_sot = [
            item["SOT"]
            for item in recent
            if pd.notna(item["SOT"])
        ]

        return {
            "gf_5": float(np.mean(gf_values)),
            "ga_5": float(np.mean(ga_values)),
            "pts_5": float(np.mean(points_values)),
            "sot_5": (
                float(np.mean(valid_sot))
                if valid_sot
                else np.nan
            ),
        }

    def build_fixture_feature_row(
        self,
        home_team: str,
        away_team: str,
        kickoff_date: str,
        historical_matches: (
            pd.DataFrame
            | Sequence[Mapping[str, Any]]
        ),
    ) -> pd.DataFrame:
        """
        Membentuk satu baris berisi 12 fitur untuk prediksi fixture.
        """

        if home_team == away_team:
            raise ValueError(
                "Home team and away team must be different"
            )

        try:
            cutoff = pd.Timestamp(kickoff_date)
        except (TypeError, ValueError) as exc:
            raise ValueError(
                f"Invalid kickoff date: {kickoff_date}"
            ) from exc

        if pd.isna(cutoff):
            raise ValueError(
                f"Invalid kickoff date: {kickoff_date}"
            )

        # Fitur menggunakan kebijakan kalender, bukan jam kickoff.
        if cutoff.tz is not None:
            cutoff = cutoff.tz_localize(None)

        cutoff = cutoff.normalize()

        history = self._normalize_historical_matches(
            historical_matches
        )

        home_stats = self._get_team_history_before(
            home_team,
            cutoff,
            history,
        )

        away_stats = self._get_team_history_before(
            away_team,
            cutoff,
            history,
        )

        row: dict[str, float] = {}

        feature_suffixes = [
            "gf_5",
            "ga_5",
            "pts_5",
            "sot_5",
        ]

        for suffix in feature_suffixes:
            h_value = home_stats[suffix]
            a_value = away_stats[suffix]

            row[f"H_{suffix}"] = h_value
            row[f"A_{suffix}"] = a_value

            row[f"diff_{suffix}"] = (
                h_value - a_value
                if pd.notna(h_value) and pd.notna(a_value)
                else np.nan
            )

        # Pastikan urutan kolom identik dengan model training.
        return pd.DataFrame(
            [row],
            columns=self.feature_columns,
            dtype=float,
        )
