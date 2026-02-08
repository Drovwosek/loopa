from pydantic import BaseModel
from typing import Optional


class DiarizationSegment(BaseModel):
    speaker: str
    start: float
    end: float
    duration: float


class DiarizationResponse(BaseModel):
    segments: list[DiarizationSegment]
    num_speakers: int


class TextProcessRequest(BaseModel):
    text: str
    detect_fillers: bool = True
    remove_fillers: bool = False


class TextSegment(BaseModel):
    text: str
    has_fillers: bool
    cleaned_text: str
    fillers_found: list[str]


class TextProcessResponse(BaseModel):
    segments: list[TextSegment]
    total_fillers: int


class WordTimestamp(BaseModel):
    word: str
    start: float
    end: float


class TranscribeSegment(BaseModel):
    speaker: str
    start: float
    end: float
    text: str
    words: list[WordTimestamp]
    has_fillers: bool
    fillers_found: list[str]


class TranscribeFullResponse(BaseModel):
    language: str
    full_text: str
    segments: list[TranscribeSegment]
    num_speakers: int
    processing_time_seconds: float
