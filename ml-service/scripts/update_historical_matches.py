
import argparse
import shutil
from datetime import datetime
from pathlib import Path

import pandas as pd


ROOT = Path(__file__).resolve().parents[2]

RAW_FILE = ROOT / "ml-service/data/raw/E0_2627.csv"
HISTORY_FILE = ROOT / "ml-service/data/processed/historical_matches.csv"

COLUMNS = [
    "Date",
    "HomeTeam",
    "AwayTeam",
    "FTHG",
    "FTAG",
    "HST",
    "AST",
]

KEYS = ["Date", "HomeTeam", "AwayTeam"]
SCORE_COLUMNS = ["FTHG", "FTAG"]
OPTIONAL_INTEGER_COLUMNS = ["HST", "AST"]
ALL_INTEGER_COLUMNS = SCORE_COLUMNS + OPTIONAL_INTEGER_COLUMNS


def normalize_matches(
    df: pd.DataFrame,
    source: str,
    dayfirst: bool,
) -> pd.DataFrame:
    """Validate and normalize match data."""

    missing = set(COLUMNS) - set(df.columns)
    if missing:
        raise ValueError(
            f"{source}: missing columns {sorted(missing)}"
        )

    result = df[COLUMNS].copy()

    date_text = result["Date"].astype("string").str.strip()

    dates = pd.to_datetime(
        date_text,
        dayfirst=dayfirst,
        errors="coerce",
    )

    invalid_dates = (
        date_text.notna()
        & date_text.ne("")
        & dates.isna()
    )

    if invalid_dates.any():
        raise ValueError(
            f"{source}: invalid dates found"
        )

    result["Date"] = dates.dt.strftime("%Y-%m-%d")

    for column in ["HomeTeam", "AwayTeam"]:
        result[column] = (
            result[column].astype("string").str.strip()
        )

    for column in ALL_INTEGER_COLUMNS:
        values = result[column].astype("string").str.strip()

        missing_values = (
            values.isna()
            | values.str.upper().isin(
                ["", "NA", "NAN", "NULL", "<NA>"]
            )
        )

        numeric = pd.to_numeric(
            values.mask(missing_values),
            errors="coerce",
        )

        invalid = ~missing_values & numeric.isna()

        if invalid.any():
            raise ValueError(
                f"{source}: invalid numeric values in {column}"
            )

        fractional = (
            numeric.notna()
            & (numeric % 1 != 0)
        )

        if fractional.any():
            raise ValueError(
                f"{source}: non-integer values in {column}"
            )

        negative = numeric.notna() & (numeric < 0)

        if negative.any():
            raise ValueError(
                f"{source}: negative values in {column}"
            )

        result[column] = numeric.astype("Int64")

    # Only completed matches are retained.
    completed = result.dropna(
        subset=[
            "Date",
            "HomeTeam",
            "AwayTeam",
            "FTHG",
            "FTAG",
        ]
    ).copy()

    if completed.empty:
        return completed

    # Completed matches must have valid team names.
    invalid_teams = (
        completed["HomeTeam"].eq("")
        | completed["AwayTeam"].eq("")
    )

    if invalid_teams.any():
        raise ValueError(
            f"{source}: completed match has an empty team name"
        )

    return completed


def make_csv_compatible(df: pd.DataFrame) -> pd.DataFrame:
    """Write integers without decimal points and null SOT as blanks."""

    result = df.copy()

    for column in SCORE_COLUMNS:
        result[column] = result[column].astype("int64")

    for column in OPTIONAL_INTEGER_COLUMNS:
        values = result[column].astype("Int64")

        result[column] = (
            values.astype(object).where(values.notna(), "")
        )

    return result


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Update historical EPL match dataset."
    )

    parser.add_argument(
        "--raw",
        type=Path,
        default=RAW_FILE,
        help="Path to the latest EPL raw CSV",
    )

    parser.add_argument(
        "--history",
        type=Path,
        default=HISTORY_FILE,
        help="Path to historical matches CSV",
    )

    parser.add_argument(
        "--apply",
        action="store_true",
        help="Apply changes and create a backup first",
    )

    args = parser.parse_args()

    if not args.raw.exists():
        raise FileNotFoundError(args.raw)

    if not args.history.exists():
        raise FileNotFoundError(args.history)

    print("Reading raw dataset...")
    raw_df = pd.read_csv(args.raw)

    # Keep only Premier League rows when division is available.
    if "Div" in raw_df.columns:
        raw_df = raw_df[
            raw_df["Div"].astype("string").str.strip().eq("E0")
        ].copy()

    incoming = normalize_matches(
        raw_df,
        source="raw dataset",
        dayfirst=True,
    )

    print("Reading historical dataset...")
    history_df = pd.read_csv(args.history)

    history = normalize_matches(
        history_df,
        source="historical dataset",
        dayfirst=False,
    )

    # Existing dataset must not contain duplicate fixtures.
    if history.duplicated(KEYS).any():
        raise ValueError(
            "Historical dataset already contains duplicate fixtures"
        )

    # Remove duplicate fixtures inside the new source.
    if incoming.duplicated(KEYS).any():
        duplicate_groups = incoming[
            incoming.duplicated(KEYS, keep=False)
        ]

        for _, group in duplicate_groups.groupby(KEYS):
            if len(group[SCORE_COLUMNS].drop_duplicates()) > 1:
                raise ValueError(
                    "Conflicting scores for the same raw fixture"
                )

        incoming = incoming.drop_duplicates(
            subset=KEYS,
            keep="last",
        )

    # Existing scores must not silently change.
    existing_scores = {
        tuple(row[key] for key in KEYS): (
            int(row["FTHG"]),
            int(row["FTAG"]),
        )
        for _, row in history.iterrows()
    }

    for _, row in incoming.iterrows():
        key = tuple(row[column] for column in KEYS)

        if key in existing_scores:
            source_scores = (
                int(row["FTHG"]),
                int(row["FTAG"]),
            )

            if existing_scores[key] != source_scores:
                raise ValueError(
                    f"Conflicting existing result for fixture {key}"
                )

    existing_keys = set(existing_scores)

    new_matches = incoming[
        [
            tuple(row[column] for column in KEYS)
            not in existing_keys
            for _, row in incoming.iterrows()
        ]
    ].copy()

    combined = pd.concat(
        [history, new_matches],
        ignore_index=True,
    )

    combined = combined.sort_values(
        KEYS,
        kind="stable",
    ).reset_index(drop=True)

    print("\nUpdate summary")
    print("------------------------------")
    print(f"Raw records       : {len(raw_df)}")
    print(f"Completed records : {len(incoming)}")
    print(f"Existing records  : {len(history)}")
    print(f"New matches       : {len(new_matches)}")
    print(f"Total after update: {len(combined)}")

    if not new_matches.empty:
        print("\nNew match date range:")
        print(
            new_matches["Date"].min(),
            "to",
            new_matches["Date"].max(),
        )

    if not args.apply:
        print("\nDRY RUN: historical CSV was not modified.")
        print("Run again with --apply to save changes.")
        return

    if new_matches.empty:
        print("\nNo new matches. Nothing to update.")
        return

    # Create a timestamped backup before replacing the CSV.
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

    backup = args.history.with_name(
        f"{args.history.stem}.backup-{timestamp}"
        f"{args.history.suffix}"
    )

    shutil.copy2(args.history, backup)

    temporary = args.history.with_name(
        f"{args.history.stem}.tmp{args.history.suffix}"
    )

    try:
        make_csv_compatible(combined).to_csv(
            temporary,
            index=False,
            lineterminator="\n",
        )

        temporary.replace(args.history)

    except Exception:
        temporary.unlink(missing_ok=True)
        raise

    print("\nDataset successfully updated.")
    print(f"Backup: {backup}")
    print(f"Updated: {args.history}")


if __name__ == "__main__":
    main()
