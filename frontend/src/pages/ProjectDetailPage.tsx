import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Table, Button, Space, Alert, Typography, Spin } from "antd";
import { ArrowLeftOutlined, EyeOutlined } from "@ant-design/icons";
import { fetchProjectFiles, type ProjectFileItem } from "../api";
import StatusTag from "../components/common/StatusTag";

const { Title } = Typography;

export default function ProjectDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [files, setFiles] = useState<ProjectFileItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    fetchProjectFiles(id)
      .then(setFiles)
      .catch((err) => setError(err instanceof Error ? err.message : "Ошибка"))
      .finally(() => setLoading(false));
  }, [id]);

  const columns = [
    {
      title: "Файл",
      dataIndex: "originalName",
      key: "originalName",
    },
    {
      title: "Статус",
      dataIndex: "status",
      key: "status",
      render: (status?: string) => status ? <StatusTag status={status} /> : "—",
    },
    {
      title: "Загружен",
      dataIndex: "uploadedAt",
      key: "uploadedAt",
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: "",
      key: "actions",
      render: (_: unknown, record: ProjectFileItem) =>
        record.taskId ? (
          <Button
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/tasks/${record.taskId}`)}
          >
            Открыть
          </Button>
        ) : null,
    },
  ];

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/projects")}>
          К проектам
        </Button>
      </Space>

      {error && (
        <Alert title={error} type="error" style={{ marginBottom: 16 }} />
      )}

      <Title level={4}>Файлы проекта</Title>

      <Table
        dataSource={files}
        columns={columns}
        rowKey="fileId"
        pagination={false}
        locale={{ emptyText: "Нет файлов в проекте" }}
      />
    </div>
  );
}
