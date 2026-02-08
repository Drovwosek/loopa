import { useMemo, useState } from "react";
import { Card, Switch, Space, Typography, Empty } from "antd";
import type { Segment } from "../../types";
import SegmentCard from "./SegmentCard";
import SpeakerLabel from "./SpeakerLabel";

const { Text } = Typography;

type TranscriptionEditorProps = {
  segments: Segment[];
  currentTimeMs: number;
  onSegmentClick: (startTimeMs: number) => void;
  onSegmentSave: (segmentId: string, text: string) => void;
  onSpeakerRename: (speakerId: string, name: string) => void;
};

export default function TranscriptionEditor({
  segments,
  currentTimeMs,
  onSegmentClick,
  onSegmentSave,
  onSpeakerRename,
}: TranscriptionEditorProps) {
  const [showFillers, setShowFillers] = useState(true);

  // Уникальные спикеры
  const speakers = useMemo(() => {
    const map = new Map<string, string | undefined>();
    for (const seg of segments) {
      if (seg.speakerId && !map.has(seg.speakerId)) {
        map.set(seg.speakerId, seg.speakerName);
      }
    }
    return Array.from(map.entries());
  }, [segments]);

  // Группируем последовательные сегменты по спикерам
  const groupedSegments = useMemo(() => {
    const groups: { speakerId: string | undefined; segments: Segment[] }[] = [];
    let currentSpeaker: string | undefined;

    for (const seg of segments) {
      if (seg.speakerId !== currentSpeaker) {
        groups.push({ speakerId: seg.speakerId, segments: [seg] });
        currentSpeaker = seg.speakerId;
      } else {
        groups[groups.length - 1].segments.push(seg);
      }
    }
    return groups;
  }, [segments]);

  // Активный сегмент (по текущему времени)
  const activeSegmentId = useMemo(() => {
    for (const seg of segments) {
      if (currentTimeMs >= seg.startTime && currentTimeMs <= seg.endTime) {
        return seg.id;
      }
    }
    return null;
  }, [segments, currentTimeMs]);

  if (segments.length === 0) {
    return (
      <Card>
        <Empty description="Сегменты транскрипции пока отсутствуют" />
      </Card>
    );
  }

  return (
    <Card
      title="Транскрипция"
      extra={
        <Space>
          <Text type="secondary">Слова-паразиты:</Text>
          <Switch
            checked={showFillers}
            onChange={setShowFillers}
            checkedChildren="Показать"
            unCheckedChildren="Скрыть"
          />
        </Space>
      }
    >
      {/* Список спикеров */}
      {speakers.length > 1 && (
        <Space wrap style={{ marginBottom: 16 }}>
          {speakers.map(([speakerId, speakerName], idx) => (
            <SpeakerLabel
              key={speakerId}
              speakerId={speakerId}
              speakerName={speakerName}
              speakerIndex={idx}
              onRename={(name) => onSpeakerRename(speakerId, name)}
            />
          ))}
        </Space>
      )}

      {/* Сегменты */}
      {groupedSegments.map((group, gi) => (
        <div key={gi} style={{ marginBottom: 16 }}>
          {group.speakerId && speakers.length > 1 && (
            <Text strong style={{ display: "block", marginBottom: 4 }}>
              {group.segments[0]?.speakerName ??
                `Спикер ${speakers.findIndex(([id]) => id === group.speakerId) + 1}`}
            </Text>
          )}
          {group.segments
            .filter((seg) => showFillers || !seg.hasFillers)
            .map((seg) => (
              <SegmentCard
                key={seg.id}
                id={seg.id}
                startTime={seg.startTime}
                endTime={seg.endTime}
                text={seg.text}
                hasFillers={seg.hasFillers}
                isCorrected={seg.isCorrected}
                isActive={seg.id === activeSegmentId}
                showFillers={showFillers}
                onClick={() => onSegmentClick(seg.startTime)}
                onSave={(text) => onSegmentSave(seg.id, text)}
              />
            ))}
        </div>
      ))}
    </Card>
  );
}
