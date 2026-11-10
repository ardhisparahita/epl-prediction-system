# EPL Prediction ML Service

Machine Learning service for English Premier League (EPL) match result prediction using Logistic Regression and Random Forest. Built with Python, FastAPI, Pandas, NumPy, and Scikit-learn.

This service provides REST API endpoints for predicting football match outcomes based on historical team performance.

## 1. Overview

The service predicts the probability of three possible match outcomes:

| Class | Description |
| ----- | ----------- |
| H     | Home Win    |
| D     | Draw        |
| A     | Away Win    |

Two classification models are used:

- Logistic Regression
- Random Forest

The models were trained using EPL match data from the 2018/2019 through 2025/2026 seasons.

The production models are frozen for the 2026/2027 season. Model weights are not retrained during the live season.

**Model version:** `epl-result-live-2026-2027-v1`

## 2. Prediction Features

Each prediction uses 12 numerical features derived from the recent performance of both teams.

| Feature    | Description                           |
| ---------- | ------------------------------------- |
| H_gf_5     | Average home team's goals scored      |
| H_ga_5     | Average home team's goals conceded    |
| H_pts_5    | Average home team's points            |
| H_sot_5    | Average home team's shots on target   |
| A_gf_5     | Average away team's goals scored      |
| A_ga_5     | Average away team's goals conceded    |
| A_pts_5    | Average away team's points            |
| A_sot_5    | Average away team's shots on target   |
| diff_gf_5  | Difference in average goals scored    |
| diff_ga_5  | Difference in average goals conceded  |
| diff_pts_5 | Difference in average points          |
| diff_sot_5 | Difference in average shots on target |

### Historical Match Policy

- Uses the last five completed matches for each team.
- Requires at least three previous matches per team.
- Includes both home and away matches.
- Excludes matches played on the same calendar date as the predicted fixture.
- Missing shots-on-target values are excluded when calculating the average.
- If all SOT values are missing, the resulting feature is handled by the fitted imputation step inside the ML pipeline.

The feature order must remain identical to the order used during model training.

## 3. Technology Stack

| Technology   | Purpose                         |
| ------------ | ------------------------------- |
| Python 3.11  | Main programming language       |
| FastAPI      | REST API framework              |
| Uvicorn      | ASGI server                     |
| Pydantic     | Request and response validation |
| Pandas       | Historical match processing     |
| NumPy        | Numerical calculations          |
| Scikit-learn | Machine Learning models         |
| Joblib       | Loading serialized models       |
| Pytest       | Automated testing               |

## 4. Project Structure

```text
ml-service/
├── app/
│   ├── api/
│   │   └── routes/
│   │       ├── health.py
│   │       └── prediction.py
│   ├── core/
│   ├── ml/
│   │   └── model_loader.py
│   ├── schemas/
│   │   └── prediction.py
│   ├── services/
│   │   ├── feature_service.py
│   │   └── prediction_service.py
│   └── main.py
├── artifacts/
│   ├── model_lr.joblib
│   ├── model_rf.joblib
│   ├── feature_columns.json
│   ├── model_metadata.json
│   ├── teams_active.json
│   └── reports/
├── data/
│   ├── raw/
│   └── processed/
├── notebooks/
│   └── football_prediction_final.ipynb
├── tests/
│   ├── test_model_loader.py
│   ├── test_feature_service.py
│   ├── test_prediction_service.py
│   ├── test_prediction_api.py
│   └── request_arsenal_chelsea.json
├── requirements.txt
├── requirements-dev.txt
└── README.md
```

## 5. Installation

Make sure Python 3.11 is installed.

Clone the repository and navigate to the ML service directory.

```bash
cd ml-service
```

Create a virtual environment:

```bash
python -m venv .venv
```

Activate the virtual environment using Git Bash:

```bash
source .venv/Scripts/activate
```

Install production dependencies:

```bash
python -m pip install -r requirements.txt
```

For development and automated testing, install:

```bash
python -m pip install -r requirements-dev.txt
```

## 6. Model Artifacts

The production models and their metadata are stored in the `artifacts/` directory.

| File                 | Description                              |
| -------------------- | ---------------------------------------- |
| model_lr.joblib      | Trained Logistic Regression pipeline     |
| model_rf.joblib      | Trained Random Forest pipeline           |
| feature_columns.json | Ordered list of 12 prediction features   |
| model_metadata.json  | Model version and training configuration |
| teams_active.json    | List of supported EPL teams              |
| reports/             | Model evaluation and experiment reports  |

All required production artifacts must be available before starting the service.

The raw and processed datasets are excluded from Git. They are needed for experimentation and retraining, but not for running the existing production inference service.

## 7. Running the Service

From the `ml-service/` directory, activate the virtual environment and run:

```bash
python -m uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

The service will be available at:

```text
http://localhost:8000
```

### API Documentation

FastAPI provides interactive API documentation:

- Swagger UI: `http://localhost:8000/docs`
- ReDoc: `http://localhost:8000/redoc`

## 8. API Endpoints

### Health Check

**GET** `/health`

Checks whether the ML service and production models have been initialized.

Example response:

```json
{
  "status": "ok",
  "service": "epl-ml-service",
  "model_loaded": true,
  "model_version": "epl-result-live-2026-2027-v1"
}
```

### Match Prediction

**POST** `/api/v1/predictions`

Predicts the probability of the match outcome using both production models.

Request structure:

```json
{
  "home_team": "Arsenal",
  "away_team": "Chelsea",
  "kickoff_date": "2026-11-15",
  "historical_matches": [
    {
      "date": "2026-10-01",
      "home_team": "Arsenal",
      "away_team": "Opponent",
      "home_goals": 2,
      "away_goals": 0,
      "home_shots_on_target": 5,
      "away_shots_on_target": 2
    }
  ]
}
```

The example above illustrates the request structure only. A valid prediction requires at least three eligible historical matches for each team before the kickoff date.

The endpoint returns the probabilities from both models.

Example response structure:

```json
{
  "home_team": "Arsenal",
  "away_team": "Chelsea",
  "kickoff_date": "2026-11-15",
  "feature_count": 12,
  "history_window": 5,
  "minimum_history_matches": 3,
  "probabilities": {
    "logistic_regression": {
      "H": 0.40889,
      "D": 0.334658,
      "A": 0.256452
    },
    "random_forest": {
      "H": 0.386105,
      "D": 0.313836,
      "A": 0.300059
    }
  },
  "model_version": "epl-result-live-2026-2027-v1"
}
```

The probabilities above are an example from a previously tested prediction, not fixed API output.

## 9. Testing

The service uses Pytest for automated testing.

Run all tests:

```bash
python -m pytest -v tests/
```

### Test Coverage

| Test File                  | Test Cases |
| -------------------------- | ---------: |
| test_model_loader.py       |         11 |
| test_feature_service.py    |          8 |
| test_prediction_service.py |         13 |
| test_prediction_api.py     |          9 |
| **Total**                  |     **41** |

The current automated test suite has passed all 41 tests.

## 10. Integration

This ML service is designed to operate separately from the Go Fiber backend.

The Go backend communicates with this service through HTTP REST API requests.

Default ML service URL:

```text
http://localhost:8000
```

The URL can be configured by the backend using the `ML_SERVICE_URL` environment variable.

## 11. Notes

- Production model artifacts must be included in the repository.
- Do not commit the Python virtual environment.
- Do not commit Python cache files.
- Raw and processed datasets are excluded from Git.
- Model weights remain frozen for the 2026/2027 live season.
- Model updates should be performed through a separate training and evaluation process.

---
