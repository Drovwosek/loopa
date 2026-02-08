import re
from typing import Optional

# Слова-паразиты для русского языка
FILLER_WORDS = {
    "ну", "это", "типа", "короче", "значит", "вот",
    "как бы", "так сказать", "угу", "ага", "ээ", "мм",
    "в общем", "собственно", "допустим", "слушай", "блин",
    "прикинь", "реально", "конкретно",
}

# Многословные паразиты (проверяются отдельно)
MULTI_WORD_FILLERS = {
    "как бы", "так сказать", "в общем",
}

# Однословные паразиты
SINGLE_WORD_FILLERS = FILLER_WORDS - MULTI_WORD_FILLERS


def detect_fillers(text: str) -> list[str]:
    """Находит слова-паразиты в тексте."""
    found: list[str] = []
    lower = text.lower()

    # Сначала проверяем многословные паразиты
    for filler in MULTI_WORD_FILLERS:
        if filler in lower:
            count = lower.count(filler)
            found.extend([filler] * count)

    # Затем однословные
    words = re.findall(r"\b\w+\b", lower)
    for word in words:
        if word in SINGLE_WORD_FILLERS:
            found.append(word)

    return found


def remove_fillers(text: str) -> str:
    """Удаляет слова-паразиты из текста, сохраняя пунктуацию."""
    result = text

    # Удаляем многословные паразиты (с учётом регистра)
    for filler in MULTI_WORD_FILLERS:
        pattern = re.compile(re.escape(filler), re.IGNORECASE)
        result = pattern.sub("", result)

    # Удаляем однословные паразиты (только как отдельные слова)
    for filler in SINGLE_WORD_FILLERS:
        pattern = re.compile(r"\b" + re.escape(filler) + r"\b", re.IGNORECASE)
        result = pattern.sub("", result)

    # Очистка лишних пробелов
    result = re.sub(r"\s+", " ", result).strip()
    # Убираем пробелы перед запятыми/точками
    result = re.sub(r"\s+([,.\?!;:])", r"\1", result)

    return result


def process_text(text: str, should_detect: bool = True, should_remove: bool = False) -> dict:
    """Обрабатывает текст: определяет и опционально удаляет паразиты."""
    fillers_found = detect_fillers(text) if should_detect else []
    cleaned = remove_fillers(text) if should_remove else text

    return {
        "text": text,
        "has_fillers": len(fillers_found) > 0,
        "cleaned_text": cleaned,
        "fillers_found": fillers_found,
    }
