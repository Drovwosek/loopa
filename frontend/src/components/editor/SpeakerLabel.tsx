import { useState } from "react";
import { Tag, Input, Space } from "antd";
import { UserOutlined, EditOutlined, CheckOutlined } from "@ant-design/icons";

// Цвета для спикеров
const SPEAKER_COLORS = [
  "#1677ff", "#52c41a", "#faad14", "#eb2f96",
  "#722ed1", "#13c2c2", "#fa541c", "#2f54eb",
];

type SpeakerLabelProps = {
  speakerId: string;
  speakerName?: string;
  speakerIndex: number;
  onRename: (name: string) => void;
};

export default function SpeakerLabel({
  speakerId,
  speakerName,
  speakerIndex,
  onRename,
}: SpeakerLabelProps) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(speakerName ?? `Спикер ${speakerIndex + 1}`);
  const color = SPEAKER_COLORS[speakerIndex % SPEAKER_COLORS.length];

  const handleSave = () => {
    setEditing(false);
    if (name.trim()) {
      onRename(name.trim());
    }
  };

  if (editing) {
    return (
      <Space size={4}>
        <Input
          size="small"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onPressEnter={handleSave}
          onBlur={handleSave}
          autoFocus
          style={{ width: 150 }}
        />
        <CheckOutlined
          style={{ cursor: "pointer", color }}
          onClick={handleSave}
        />
      </Space>
    );
  }

  return (
    <Tag
      color={color}
      icon={<UserOutlined />}
      style={{ cursor: "pointer" }}
      onClick={() => setEditing(true)}
    >
      {name} <EditOutlined style={{ fontSize: 10 }} />
    </Tag>
  );
}
