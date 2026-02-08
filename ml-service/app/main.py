import asyncio
import logging
import os
import tempfile
import time
from typing import Optional

from fastapi import FastAPI, File, UploadFile, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware

from .config import settings
from .diarization import diarize
from .alignment import align_words_to_speakers
from .transcription import transcribe
from .models import (
    DiarizationResponse,
    TextProcessRequest,
    TextProcessResponse,
    TranscribeFullResponse,
)
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


_transcribe_lock = asyncio.Semaphore(1)


def _do_transcribe_full(
    audio_path: str,
    language: Optional[str],
    num_speakers: Optional[int],
    detect_fillers: bool,
) -> dict:
    """Синхронный pipeline: Whisper → PyAnnote → alignment → fillers."""
    start_time = time.time()

    # Шаг 1: Транскрибация через Faster-Whisper
    whisper_result = transcribe(audio_path, language=language)

    # Шаг 2: Диаризация через PyAnnote
    try:
        diarization_segments = diarize(audio_path, num_speakers=num_speakers)
    except Exception as e:
        logger.warning("Диаризация не удалась (non-fatal): %s", e)
        diarization_segments = []

    # Шаг 3: Alignment слов к спикерам
    aligned = align_words_to_speakers(whisper_result["words"], diarization_segments)

    # Шаг 4: Детектор паразитов по сегментам
    for seg in aligned:
        if detect_fillers:
            result = process_text(seg["text"], should_detect=True, should_remove=False)
            seg["has_fillers"] = result["has_fillers"]
            seg["fillers_found"] = result["fillers_found"]
        else:
            seg["has_fillers"] = False
            seg["fillers_found"] = []

    speakers = set(s["speaker"] for s in aligned)

    return {
        "language": whisper_result["language"],
        "full_text": whisper_result["full_text"],
        "segments": aligned,
        "num_speakers": len(speakers),
        "processing_time_seconds": round(time.time() - start_time, 2),
    }


@app.post("/transcribe-full", response_model=TranscribeFullResponse)
async def transcribe_full_endpoint(
    audio: UploadFile = File(...),
    language: Optional[str] = Query(None, description="Код языка (ru, en, ...) или пусто для автодетекта"),
    num_speakers: Optional[int] = Query(None, ge=1, le=20, description="Ожидаемое количество спикеров"),
    detect_fillers: bool = Query(True, description="Определять слова-паразиты"),
):
    """Полный pipeline: транскрибация + диаризация + alignment + детектор паразитов."""
    suffix = os.path.splitext(audio.filename or ".wav")[1]
    try:
        with tempfile.NamedTemporaryFile(suffix=suffix, delete=False) as tmp:
            content = await audio.read()
            tmp.write(content)
            tmp_path = tmp.name

        async with _transcribe_lock:
            loop = asyncio.get_event_loop()
            result = await loop.run_in_executor(
                None,
                _do_transcribe_full,
                tmp_path,
                language,
                num_speakers,
                detect_fillers,
            )

        return TranscribeFullResponse(**result)
    except Exception as e:
        logger.exception("Ошибка транскрибации")
        raise HTTPException(status_code=500, detail=f"Ошибка транскрибации: {str(e)}")
    finally:
        if "tmp_path" in locals():
            os.unlink(tmp_path)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host=settings.HOST, port=settings.PORT)
