import logging

from app.config import settings

logger = logging.getLogger(__name__)

_model = None


def _get_model():
    global _model
    if _model is not None:
        return _model

    from faster_whisper import WhisperModel

    logger.info(
        "Loading Whisper model: %s (device=%s, compute_type=%s)",
        settings.WHISPER_MODEL,
        settings.WHISPER_DEVICE,
        settings.WHISPER_COMPUTE_TYPE,
    )

    _model = WhisperModel(
        settings.WHISPER_MODEL,
        device=settings.WHISPER_DEVICE,
        compute_type=settings.WHISPER_COMPUTE_TYPE,
    )

    logger.info("Whisper model loaded successfully")
    return _model


def transcribe(audio_path: str, language: str | None = None) -> dict:
    """Транскрибирует аудиофайл через Faster-Whisper с word-level timestamps."""
    model = _get_model()

    kwargs = {"word_timestamps": True, "beam_size": 5}
    if language:
        kwargs["language"] = language

    segments, info = model.transcribe(audio_path, **kwargs)

    all_words = []
    full_text_parts = []

    for segment in segments:
        text = segment.text.strip()
        if text:
            full_text_parts.append(text)
        if segment.words:
            for word in segment.words:
                all_words.append({
                    "word": word.word.strip(),
                    "start": round(word.start, 3),
                    "end": round(word.end, 3),
                })

    return {
        "language": info.language,
        "language_probability": round(info.language_probability, 3),
        "full_text": " ".join(full_text_parts),
        "words": all_words,
    }
