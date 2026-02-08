import { useState } from "react";
import { Card, Tag, Input, Button, Space, Typography } from "antd";
import { EditOutlined, SaveOutlined, ClockCircleOutlined } from "@ant-design/icons";

const { Text, Paragraph } = Typography;
const { TextArea } = Input;

type SegmentCardProps = {
  id: string;
  startTime: number;
  endTime: number;
  text: string;
  hasFillers: boolean;
  isCorrected: boolean;
  isActive: boolean;
  showFillers: boolean;
  onClick: () => void;
  onSave: (text: string) => void;
};

function formatMs(ms: number): string {
  const totalSec = ms / 1000;
  const m = Math.floor(totalSec / 60);
  const s = Math.floor(totalSec % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export default function SegmentCard({
  startTime,
  endTime,
  text,
  hasFillers,
  isCorrected,
  isActive,
  onClick,
  onSave,
}: SegmentCardProps) {
  const [editing, setEditing] = useState(false);
  const [editedText, setEditedText] = useState(text);

  const handleSave = () => {
    setEditing(false);
    if (editedText.trim() !== text) {
      onSave(editedText.trim());
    }
  };

  return (
    <Card
      size="small"
      style={{
        marginBottom: 8,
        cursor: "pointer",
        borderLeft: isActive ? "3px solid #1677ff" : "3px solid transparent",
        backgroundColor: isActive ? "#f0f5ff" : undefined,
      }}
      onClick={onClick}
      onDoubleClick={() => {
        setEditing(true);
        setEditedText(text);
      }}
    >
      <Space orientation="vertical" style={{ width: "100%" }} size={4}>
        <Space size={8}>
          <Tag icon={<ClockCircleOutlined />} color="blue">
            {formatMs(startTime)} — {formatMs(endTime)}
          </Tag>
          {hasFillers && <Tag color="orange">Паразиты</Tag>}
          {isCorrected && <Tag color="green">Исправлено</Tag>}
        </Space>

        {editing ? (
          <Space orientation="vertical" style={{ width: "100%" }}>
            <TextArea
              value={editedText}
              onChange={(e) => setEditedText(e.target.value)}
              autoSize={{ minRows: 2 }}
              autoFocus
              onClick={(e) => e.stopPropagation()}
            />
            <Space>
              <Button
                type="primary"
                size="small"
                icon={<SaveOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  handleSave();
                }}
              >
                Сохранить
              </Button>
              <Button
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  setEditing(false);
                }}
              >
                Отмена
              </Button>
            </Space>
          </Space>
        ) : (
          <Space>
            <Paragraph style={{ margin: 0 }}>{text}</Paragraph>
            <Button
              type="text"
              size="small"
              icon={<EditOutlined />}
              onClick={(e) => {
                e.stopPropagation();
                setEditing(true);
                setEditedText(text);
              }}
            />
          </Space>
        )}
      </Space>
    </Card>
  );
}
