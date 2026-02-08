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
