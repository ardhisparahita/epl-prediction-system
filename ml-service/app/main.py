from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.api.routes.health import router as health_router
from app.api.routes.prediction import router as prediction_router

from app.ml.model_loader import load_artifacts

from app.services.feature_service import FeatureService
from app.services.prediction_service import PredictionService


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Load production model.
    artifacts = load_artifacts()

    # Initialize feature service.
    feature_service = FeatureService(
        feature_columns=artifacts.feature_columns,
        history_window=artifacts.metadata[
            "history_window_matches"
        ],
        minimum_history_matches=artifacts.metadata[
            "minimum_history_matches"
        ],
    )

    # Initialize prediction service.
    prediction_service = PredictionService(
        artifacts=artifacts,
        feature_service=feature_service,
    )

    app.state.artifacts = artifacts
    app.state.feature_service = feature_service
    app.state.prediction_service = prediction_service

    print("EPL ML Service initialized")
    print("Model version:", artifacts.metadata["model_version"])

    yield

    print("EPL ML Service shutdown")


app = FastAPI(
    title="EPL Prediction ML Service",
    description=(
        "Machine Learning API for English Premier League "
        "match result prediction using Random Forest "
        "and Logistic Regression."
    ),
    version="1.0.0",
    lifespan=lifespan,
)

# Health endpoint.
app.include_router(health_router)

# Prediction endpoint.
app.include_router(
    prediction_router,
    prefix="/api/v1/predictions",
)