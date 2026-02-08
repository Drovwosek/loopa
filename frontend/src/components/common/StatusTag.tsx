import { Tag } from "antd";

const statusConfig: Record<string, { color: string; label: string }> = {
  "ожидает": { color: "default", label: "Ожидает" },
  "в процессе": { color: "processing", label: "В процессе" },
  "готово": { color: "success", label: "Готово" },
  "ошибка": { color: "error", label: "Ошибка" },
};

type StatusTagProps = {
  status: string;
};

export default function StatusTag({ status }: StatusTagProps) {
  const config = statusConfig[status] ?? { color: "default", label: status };
  return <Tag color={config.color}>{config.label}</Tag>;
}
