import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Row, Col, Table, Button, Alert, Space, Popconfirm, Select } from "antd";
import { DeleteOutlined, EyeOutlined } from "@ant-design/icons";
import { deleteTask, uploadFile, fetchProjects } from "../api";
import { useAppDispatch, useAppSelector } from "../hooks";
import { loadHistory } from "../store/historySlice";
import FileUpload from "../components/upload/FileUpload";
import StatusTag from "../components/common/StatusTag";
import type { HistoryItem } from "../api";
import type { Project } from "../types";

export default function HomePage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const history = useAppSelector((state) => state.history.items);
  const loading = useAppSelector((state) => state.history.loading);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string | undefined>();

  useEffect(() => {
    dispatch(loadHistory());
    fetchProjects().then(setProjects).catch(() => {});
  }, [dispatch]);

  const handleUpload = async (file: File) => {
    setUploading(true);
    setError(null);
    try {
      const taskId = await uploadFile(file, selectedProjectId);
      navigate(`/tasks/${taskId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка загрузки");
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async (taskId: string) => {
    try {
      await deleteTask(taskId);
      dispatch(loadHistory());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка удаления");
    }
  };

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
      render: (status: string) => <StatusTag status={status} />,
    },
    {
      title: "Загружен",
      dataIndex: "uploadedAt",
      key: "uploadedAt",
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: "Действия",
      key: "actions",
      render: (_: unknown, record: HistoryItem) => (
        <Space>
          <Button
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/tasks/${record.id}`)}
          >
            Открыть
          </Button>
          <Popconfirm
            title="Удалить задачу?"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button size="small" danger icon={<DeleteOutlined />}>
              Удалить
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Row gutter={[24, 24]}>
      <Col xs={24} lg={12}>
        {projects.length > 0 && (
          <div style={{ marginBottom: 12 }}>
            <Select
              placeholder="Проект (необязательно)"
              allowClear
              style={{ width: "100%" }}
              value={selectedProjectId}
              onChange={setSelectedProjectId}
              options={projects.map((p) => ({ label: p.name, value: p.id }))}
            />
          </div>
        )}
        <FileUpload onUploadStart={handleUpload} uploading={uploading} />
        {error && (
          <Alert
            title={error}
            type="error"
            closable
            onClose={() => setError(null)}
            style={{ marginTop: 16 }}
          />
        )}
      </Col>
      <Col xs={24} lg={12}>
        <Table
          dataSource={history}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={false}
          size="small"
          title={() => <strong>Последние загрузки</strong>}
          locale={{ emptyText: "Нет загрузок" }}
        />
      </Col>
    </Row>
  );
}
