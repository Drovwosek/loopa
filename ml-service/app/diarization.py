import logging
from typing import Optional

import torch

from .config import settings

logger = logging.getLogger(__name__)

# Ленивая загрузка пайплайна — грузим только при первом вызове
_pipeline = None


def _get_pipeline():
    global _pipeline
    if _pipeline is not None:
        return _pipeline

    from pyannote.audio import Pipeline

    logger.info("Загрузка пайплайна диаризации %s...", settings.DIARIZATION_MODEL)
    device = torch.device(settings.DEVICE)

    _pipeline = Pipeline.from_pretrained(
        settings.DIARIZATION_MODEL,
        use_auth_token=settings.HF_TOKEN,
    )
    _pipeline.to(device)
    logger.info("Пайплайн диаризации загружен на %s", device)
    return _pipeline


def diarize(audio_path: str, num_speakers: Optional[int] = None) -> list[dict]:
    """
    Диаризация аудиофайла.

    Args:
        audio_path: путь к аудиофайлу
        num_speakers: ожидаемое количество спикеров (опционально)

    Returns:
        Список сегментов с информацией о спикерах
    """
    pipeline = _get_pipeline()

    kwargs = {}
    if num_speakers is not None:
        kwargs["num_speakers"] = num_speakers

    diarization = pipeline(audio_path, **kwargs)

    segments = []
    for turn, _, speaker in diarization.itertracks(yield_label=True):
        segments.append({
            "speaker": speaker,
            "start": round(turn.start, 3),
            "end": round(turn.end, 3),
            "duration": round(turn.end - turn.start, 3),
        })

    return segments
