"""Alignment слов (Whisper) к спикерам (PyAnnote)."""


def _find_best_speaker(
    word_start: float,
    word_end: float,
    diarization_segments: list[dict],
) -> str | None:
    """Находит спикера с максимальным перекрытием по времени для данного слова."""
    best_speaker = None
    best_overlap = 0.0

    for seg in diarization_segments:
        overlap_start = max(word_start, seg["start"])
        overlap_end = min(word_end, seg["end"])
        overlap = max(0.0, overlap_end - overlap_start)

        if overlap > best_overlap:
            best_overlap = overlap
            best_speaker = seg["speaker"]

    return best_speaker


def _find_nearest_speaker(
    word_start: float,
    diarization_segments: list[dict],
) -> str:
    """Fallback: назначает слово ближайшему спикеру по времени."""
    if not diarization_segments:
        return "UNKNOWN"

    best_speaker = diarization_segments[0]["speaker"]
    best_distance = float("inf")

    for seg in diarization_segments:
        dist_start = abs(word_start - seg["start"])
        dist_end = abs(word_start - seg["end"])
        dist = min(dist_start, dist_end)

        if dist < best_distance:
            best_distance = dist
            best_speaker = seg["speaker"]

    return best_speaker


def align_words_to_speakers(
    words: list[dict],
    diarization_segments: list[dict],
) -> list[dict]:
    """
    Назначает каждое слово спикеру и группирует последовательные слова
    одного спикера в сегменты.

    Args:
        words: [{word, start, end}, ...] из Whisper
        diarization_segments: [{speaker, start, end}, ...] из PyAnnote

    Returns:
        [{speaker, start, end, text, words: [{word, start, end}]}, ...]
    """
    if not words:
        return []

    if not diarization_segments:
        return [{
            "speaker": "UNKNOWN",
            "start": words[0]["start"],
            "end": words[-1]["end"],
            "text": " ".join(w["word"] for w in words),
            "words": words,
        }]

    # Назначаем спикера каждому слову
    for word in words:
        speaker = _find_best_speaker(word["start"], word["end"], diarization_segments)
        if speaker is None:
            speaker = _find_nearest_speaker(word["start"], diarization_segments)
        word["speaker"] = speaker

    # Группируем последовательные слова одного спикера
    output_segments = []
    current_speaker = words[0]["speaker"]
    current_words = [words[0]]

    for word in words[1:]:
        if word["speaker"] != current_speaker:
            output_segments.append(_build_segment(current_speaker, current_words))
            current_speaker = word["speaker"]
            current_words = [word]
        else:
            current_words.append(word)

    output_segments.append(_build_segment(current_speaker, current_words))

    return output_segments


def _build_segment(speaker: str, words: list[dict]) -> dict:
    """Создаёт сегмент из группы слов одного спикера."""
    return {
        "speaker": speaker,
        "start": round(words[0]["start"], 3),
        "end": round(words[-1]["end"], 3),
        "text": " ".join(w["word"] for w in words),
        "words": [{"word": w["word"], "start": w["start"], "end": w["end"]} for w in words],
    }
