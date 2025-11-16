from fastapi import FastAPI
from pydantic import BaseModel


app = FastAPI(title="Chunk Size Optimizer Service")


class ChunkSizeRequest(BaseModel):
    size_bytes: int
    mime_type: str | None = None
    estimated_bandwidth_mbps: float | None = None
    latency_ms: float | None = None


class ChunkSizeResponse(BaseModel):
    chunk_size_mb: float


def simple_lightgbm_like_model(
    size_bytes: int,
    estimated_bandwidth_mbps: float | None,
    latency_ms: float | None,
) -> float:
    """
    Placeholder for a real LightGBM/XGBoost model.
    Implements a simple heuristic but can be swapped with a trained model.
    Returns chunk size in megabytes.
    """
    mb = 1024 * 1024
    size_mb = size_bytes / mb

    # Base rule on file size
    if size_mb <= 100:
        base = 8.0
    elif size_mb <= 1024:
        base = 32.0
    elif size_mb <= 10_240:
        base = 64.0
    else:
        base = 128.0

    # Adjust for bandwidth (higher bandwidth -> bigger chunks)
    if estimated_bandwidth_mbps is not None:
        if estimated_bandwidth_mbps > 200:
            base *= 1.5
        elif estimated_bandwidth_mbps < 20:
            base *= 0.75

    # Adjust for latency (higher latency -> bigger chunks)
    if latency_ms is not None:
        if latency_ms > 100:
            base *= 1.25
        elif latency_ms < 20:
            base *= 0.9

    # Clamp to a reasonable range
    return max(4.0, min(base, 256.0))


@app.post("/predict-chunk-size", response_model=ChunkSizeResponse)
def predict_chunk_size(req: ChunkSizeRequest) -> ChunkSizeResponse:
    chunk_size_mb = simple_lightgbm_like_model(
        size_bytes=req.size_bytes,
        estimated_bandwidth_mbps=req.estimated_bandwidth_mbps,
        latency_ms=req.latency_ms,
    )
    return ChunkSizeResponse(chunk_size_mb=chunk_size_mb)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)


