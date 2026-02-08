import logging
import os
import tempfile
from typing import Optional

from fastapi import FastAPI, File, UploadFile, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware

from .config import settings
from .diarization import diarize
from .models import DiarizationResponse, TextProcessRequest, TextProcessResponse
from .text_processor import process_text

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Loopa ML Service",
    version="1.0.0",
    description="Микросервис для диаризации и обработки текста",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.post("/diarize", response_model=DiarizationResponse)
async def diarize_audio(
    audio: UploadFile = File(...),
    num_speakers: Optional[int] = Query(None, ge=1, le=20),
):
    """Диаризация аудиофайла — определение сегментов по спикерам."""
    suffix = os.path.splitext(audio.filename or ".ogg")[1]
    try:
        with tempfile.NamedTemporaryFile(suffix=suffix, delete=False) as tmp:
            content = await audio.read()
            tmp.write(content)
            tmp_path = tmp.name

        segments = diarize(tmp_path, num_speakers=num_speakers)
        speakers = set(s["speaker"] for s in segments)

        return DiarizationResponse(
            segments=segments,
            num_speakers=len(speakers),
        )
    except Exception as e:
        logger.exception("Ошибка диаризации")
        raise HTTPException(status_code=500, detail=f"Ошибка диаризации: {str(e)}")
    finally:
        if "tmp_path" in locals():
            os.unlink(tmp_path)


@app.post("/process-text", response_model=TextProcessResponse)
async def process_text_endpoint(request: TextProcessRequest):
    """Обработка текста: определение и удаление слов-паразитов."""
    # Разбиваем текст на предложения для посегментной обработки
    import re
    sentences = re.split(r"(?<=[.!?])\s+", request.text)
    if not sentences or (len(sentences) == 1 and not sentences[0].strip()):
        sentences = [request.text]

    results = []
    total_fillers = 0

    for sentence in sentences:
        if not sentence.strip():
            continue
        result = process_text(
            sentence,
            should_detect=request.detect_fillers,
            should_remove=request.remove_fillers,
        )
        total_fillers += len(result["fillers_found"])
        results.append(result)

    return TextProcessResponse(
        segments=results,
        total_fillers=total_fillers,
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host=settings.HOST, port=settings.PORT)
