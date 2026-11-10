from fastapi import APIRouter, HTTPException, Request

router = APIRouter(tags=["Health"])


@router.get("/health")
def health_check(request: Request):
    prediction_service = getattr(
        request.app.state,
        "prediction_service",
        None,
    )

    if prediction_service is None:
        raise HTTPException(
            status_code=503,
            detail="Prediction service is not initialized",
        )

    return {
        "status": "ok",
        "service": "epl-ml-service",
        "model_loaded": True,
        "model_version": (
            prediction_service.artifacts.metadata["model_version"]
        ),
    }