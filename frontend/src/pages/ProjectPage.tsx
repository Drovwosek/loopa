import { useEffect, useState } from "react";
import {
  Card,
  Table,
  Button,
  Space,
  Input,
  Modal,
  Empty,
  Popconfirm,
  Typography,
} from "antd";
import { PlusOutlined, DeleteOutlined, FolderOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../hooks";
import { loadProjects, addProject, removeProject } from "../store/projectSlice";
import type { Project } from "../types";

const { Title } = Typography;

export default function ProjectPage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const projects = useAppSelector((state) => state.projects.items);
  const loading = useAppSelector((state) => state.projects.loading);
  const [modalOpen, setModalOpen] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  useEffect(() => {
    dispatch(loadProjects());
  }, [dispatch]);

  const handleCreate = async () => {
    if (!name.trim()) return;
    await dispatch(addProject({ name: name.trim(), description: description.trim() || undefined }));
    setModalOpen(false);
    setName("");
    setDescription("");
  };

  const columns = [
    {
      title: "",
      key: "icon",
      width: 40,
      render: () => <FolderOutlined style={{ fontSize: 18, color: "#1677ff" }} />,
    },
    {
      title: "Название",
      dataIndex: "name",
      key: "name",
      render: (name: string, record: Project) => (
        <div>
          <strong>{name}</strong>
          {record.description && (
            <div style={{ color: "#999", fontSize: 12 }}>{record.description}</div>
          )}
        </div>
      ),
    },
    {
      title: "Файлов",
      dataIndex: "fileCount",
      key: "fileCount",
      width: 100,
    },
    {
      title: "Создан",
      dataIndex: "createdAt",
      key: "createdAt",
      render: (date: string) => new Date(date).toLocaleDateString(),
      width: 120,
    },
    {
      title: "",
      key: "actions",
      width: 100,
      render: (_: unknown, record: Project) => (
        <Popconfirm
          title="Удалить проект?"
          onConfirm={() => dispatch(removeProject(record.id))}
        >
          <Button size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, width: "100%", justifyContent: "space-between" }}>
        <Title level={4} style={{ margin: 0 }}>Проекты</Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setModalOpen(true)}
        >
          Новый проект
        </Button>
      </Space>

      <Table
        dataSource={projects}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={false}
        locale={{ emptyText: <Empty description="Нет проектов" /> }}
        onRow={(record) => ({
          onClick: () => navigate(`/projects/${record.id}`),
          style: { cursor: "pointer" },
        })}
      />

      <Modal
        title="Новый проект"
        open={modalOpen}
        onOk={handleCreate}
        onCancel={() => setModalOpen(false)}
        okText="Создать"
        cancelText="Отмена"
        okButtonProps={{ disabled: !name.trim() }}
      >
        <Space orientation="vertical" style={{ width: "100%" }} size={12}>
          <Input
            placeholder="Название проекта"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onPressEnter={handleCreate}
            autoFocus
          />
          <Input.TextArea
            placeholder="Описание (необязательно)"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
          />
        </Space>
      </Modal>
    </div>
  );
}
