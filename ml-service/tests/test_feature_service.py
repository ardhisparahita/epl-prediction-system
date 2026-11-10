
import numpy as np
import pandas as pd
import pytest

from app.services.feature_service import FeatureService


FEATURE_COLUMNS = [
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


@pytest.fixture
def feature_service():
    return FeatureService(
        feature_columns=FEATURE_COLUMNS,
        history_window=5,
        minimum_history_matches=3,
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
    return [
        make_match("2025-12-01", "Arsenal", "Old Team A", 9, 0, 9, 0),
        make_match("2025-12-02", "Old Team B", "Chelsea", 0, 9, 0, 9),

        make_match("2026-01-01", "Arsenal", "Opponent 1", 2, 0, 5, 2),
        make_match("2026-01-05", "Opponent 2", "Arsenal", 1, 1, 4, 6),
        make_match("2026-01-10", "Arsenal", "Opponent 3", 0, 2, 3, 7),
        make_match("2026-01-15", "Opponent 4", "Arsenal", 0, 3, 2, 8),
        make_match("2026-01-20", "Arsenal", "Opponent 5", 1, 0, 4, 1),

        make_match("2026-01-02", "Opponent 6", "Chelsea", 1, 2, 2, 3),
        make_match("2026-01-06", "Chelsea", "Opponent 7", 1, 1, 4, 4),
        make_match("2026-01-11", "Opponent 8", "Chelsea", 2, 0, 5, 2),
        make_match("2026-01-16", "Chelsea", "Opponent 9", 3, 1, 7, 5),
        make_match("2026-01-21", "Opponent 10", "Chelsea", 0, 1, 1, 6),
    ]


def test_build_fixture_feature_row_uses_last_five_matches(
    feature_service,
    historical_matches,
):
    result = feature_service.build_fixture_feature_row(
        home_team="Arsenal",
        away_team="Chelsea",
        kickoff_date="2026-02-01",
        historical_matches=historical_matches,
    )

    assert result.shape == (1, 12)
    assert list(result.columns) == FEATURE_COLUMNS

    expected = [
        1.4, 0.6, 2.0, 5.2,
        1.4, 1.0, 2.0, 4.4,
        0.0, -0.4, 0.0, 0.8,
    ]

    np.testing.assert_allclose(
        result.iloc[0].to_numpy(dtype=float),
        expected,
        rtol=1e-6,
    )


def test_excludes_matches_on_same_calendar_date(
    feature_service,
    historical_matches,
):
    baseline = feature_service.build_fixture_feature_row(
        "Arsenal",
        "Chelsea",
        "2026-02-01",
        historical_matches,
    )

    history_with_same_date = historical_matches + [
        make_match(
            "2026-02-01",
            "Arsenal",
            "Opponent 11",
            10,
            0,
            10,
            0,
        ),
        make_match(
            "2026-02-01",
            "Opponent 12",
            "Chelsea",
            0,
            10,
            0,
            10,
        ),
    ]

    result = feature_service.build_fixture_feature_row(
        "Arsenal",
        "Chelsea",
        "2026-02-01",
        history_with_same_date,
    )

    np.testing.assert_allclose(
        result.to_numpy(dtype=float),
        baseline.to_numpy(dtype=float),
    )


def test_raises_error_when_team_has_insufficient_history(
    feature_service,
):
    history = [
        make_match("2026-01-01", "Arsenal", "Opponent 1", 1, 0),
        make_match("2026-01-02", "Opponent 2", "Arsenal", 0, 1),

        make_match("2026-01-01", "Chelsea", "Opponent 3", 1, 0),
        make_match("2026-01-02", "Opponent 4", "Chelsea", 0, 1),
        make_match("2026-01-03", "Chelsea", "Opponent 5", 2, 0),
    ]

    with pytest.raises(
        ValueError,
        match="Riwayat Arsenal hanya 2 pertandingan",
    ):
        feature_service.build_fixture_feature_row(
            "Arsenal",
            "Chelsea",
            "2026-02-01",
            history,
        )


def test_missing_sot_values_are_ignored(feature_service):
    history = [
        make_match("2026-01-01", "Arsenal", "Opponent 1", 1, 0, 2, 1),
        make_match("2026-01-02", "Opponent 2", "Arsenal", 2, 2, 4, None),
        make_match("2026-01-03", "Arsenal", "Opponent 3", 3, 1, 6, 2),

        make_match("2026-01-01", "Chelsea", "Opponent 4", 0, 0, None, 3),
        make_match("2026-01-02", "Opponent 5", "Chelsea", 1, 1, 2, None),
        make_match("2026-01-03", "Chelsea", "Opponent 6", 2, 0, None, 1),
    ]

    result = feature_service.build_fixture_feature_row(
        "Arsenal",
        "Chelsea",
        "2026-02-01",
        history,
    )

    assert result.loc[0, "H_sot_5"] == pytest.approx(4.0)
    assert np.isnan(result.loc[0, "A_sot_5"])
    assert np.isnan(result.loc[0, "diff_sot_5"])


def test_rejects_same_home_and_away_team(feature_service):
    with pytest.raises(
        ValueError,
        match="Home team and away team must be different",
    ):
        feature_service.build_fixture_feature_row(
            "Arsenal",
            "Arsenal",
            "2026-02-01",
            [],
        )


def test_rejects_invalid_kickoff_date(feature_service):
    with pytest.raises(
        ValueError,
        match="Invalid kickoff date",
    ):
        feature_service.build_fixture_feature_row(
            "Arsenal",
            "Chelsea",
            "invalid-date",
            [],
        )


def test_rejects_empty_historical_matches(feature_service):
    with pytest.raises(
        ValueError,
        match="Historical matches are empty",
    ):
        feature_service.build_fixture_feature_row(
            "Arsenal",
            "Chelsea",
            "2026-02-01",
            [],
        )


def test_rejects_missing_required_columns(feature_service):
    history = [
        {
            "Date": "2026-01-01",
            "HomeTeam": "Arsenal",
            "FTHG": 1,
        }
    ]

    with pytest.raises(
        ValueError,
        match="Missing historical match columns",
    ):
        feature_service.build_fixture_feature_row(
            "Arsenal",
            "Chelsea",
            "2026-02-01",
            history,
        )
