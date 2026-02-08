import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Button, Space, Spin, Alert, Typography, Card } from "antd";
import {
  ArrowLeftOutlined,
  DownloadOutlined,
  CopyOutlined,
} from "@ant-design/icons";
import {
  downloadExport,
  fetchSegments,
  fetchTask,
  getAudioUrl,
  updateSegment,
  updateSpeaker,
} from "../api";
import type { TaskResponse } from "../api";
import type { Segment } from "../types";
import StatusTag from "../components/common/StatusTag";
import AudioPlayer from "../components/player/AudioPlayer";
import TranscriptionEditor from "../components/editor/TranscriptionEditor";

const { Title, Paragraph } = Typography;

export default function TaskPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [task, setTask] = useState<TaskResponse | null>(null);
  const [segments, setSegments] = useState<Segment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentTimeMs, setCurrentTimeMs] = useState(0);

  // Загрузка задачи
  useEffect(() => {
    if (!id) return;
    let cancelled = false;

    const load = async () => {
      try {
        const data = await fetchTask(id);
        if (!cancelled) setTask(data);

        if (data.status === "готово") {
          const segs = await fetchSegments(id);
          if (!cancelled) setSegments(segs);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Ошибка загрузки");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    load();
    return () => { cancelled = true; };
  }, [id]);

  // Polling пока задача не готова
  useEffect(() => {
    if (!id || !task) return;
    if (task.status === "готово" || task.status === "ошибка") return;

    const timer = setInterval(async () => {
      try {
        const data = await fetchTask(id);
        setTask(data);
        if (data.status === "готово") {
          const segs = await fetchSegments(id);
          setSegments(segs);
        }
      } catch {
        // Ignore polling errors
      }
    }, 2500);

    return () => clearInterval(timer);
  }, [id, task?.status]);

  const handleSegmentClick = useCallback((startTimeMs: number) => {
    const seekFn = (window as any).__loopaAudioSeek;
    if (seekFn) seekFn(startTimeMs);
  }, []);

  const handleSegmentSave = useCallback(
    async (segmentId: string, text: string) => {
      if (!id) return;
      try {
        await updateSegment(id, segmentId, text);
        setSegments((prev) =>
          prev.map((s) =>
            s.id === segmentId ? { ...s, text, isCorrected: true } : s
          )
        );
      } catch {
        setError("Не удалось сохранить сегмент");
      }
    },
    [id]
  );

  const handleSpeakerRename = useCallback(
    async (speakerId: string, name: string) => {
      if (!id) return;
      try {
        await updateSpeaker(id, speakerId, name);
        setSegments((prev) =>
          prev.map((s) =>
            s.speakerId === speakerId ? { ...s, speakerName: name } : s
          )
        );
      } catch {
        setError("Не удалось переименовать спикера");
      }
    },
    [id]
  );

  const handleCopy = async () => {
    if (!task?.transcriptText) return;
    await navigator.clipboard.writeText(task.transcriptText);
  };

  if (!id) {
    return (
      <div>
        <Alert title="ID задачи не указан" type="error" />
        <Button onClick={() => navigate("/")}>Назад</Button>
      </div>
    );
  }

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate("/")}
        style={{ marginBottom: 16 }}
      >
        Назад
      </Button>

      {error && (
        <Alert
          title={error}
          type="error"
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 16 }}
        />
      )}

      {task ? (
        <>
          <Card style={{ marginBottom: 16 }}>
            <Space orientation="vertical">
              <Title level={4} style={{ margin: 0 }}>
                {task.originalName}
              </Title>
              <Space>
                <StatusTag status={task.status} />
                {task.completedAt && (
                  <span style={{ color: "#999" }}>
                    Завершено: {new Date(task.completedAt).toLocaleString()}
                  </span>
                )}
              </Space>
            </Space>
          </Card>

          {task.errorMessage && (
            <Alert
              title="Ошибка обработки"
              description={task.errorMessage}
              type="error"
              style={{ marginBottom: 16 }}
            />
          )}

          {task.status === "готово" && (
            <>
              {/* Аудиоплеер */}
              <AudioPlayer
                audioUrl={getAudioUrl(id)}
                onTimeUpdate={setCurrentTimeMs}
              />

              {/* Редактор транскрипции */}
              {segments.length > 0 ? (
                <TranscriptionEditor
                  segments={segments}
                  currentTimeMs={currentTimeMs}
                  onSegmentClick={handleSegmentClick}
                  onSegmentSave={handleSegmentSave}
                  onSpeakerRename={handleSpeakerRename}
                />
              ) : task.transcriptText ? (
                <Card title="Транскрипция">
                  <Paragraph>
                    <pre style={{ whiteSpace: "pre-wrap", fontFamily: "inherit" }}>
                      {task.transcriptText}
                    </pre>
                  </Paragraph>
                </Card>
              ) : null}

              {/* Действия */}
              <Card size="small" style={{ marginTop: 16 }}>
                <Space>
                  <Button icon={<CopyOutlined />} onClick={handleCopy}>
                    Копировать текст
                  </Button>
                  <Button
                    icon={<DownloadOutlined />}
                    onClick={() => downloadExport(id, "txt")}
                  >
                    Скачать TXT
                  </Button>
                  <Button
                    icon={<DownloadOutlined />}
                    onClick={() => downloadExport(id, "docx")}
                  >
                    Скачать DOCX
                  </Button>
                </Space>
              </Card>
            </>
          )}

          {(task.status === "ожидает" || task.status === "в процессе") && (
            <Card>
              <div style={{ textAlign: "center", padding: 32 }}>
                <Spin size="large" />
                <Title level={5} style={{ marginTop: 16 }}>
                  Идёт обработка...
                </Title>
                <Paragraph type="secondary">
                  Транскрибация может занять несколько минут
                </Paragraph>
              </div>
            </Card>
          )}
        </>
      ) : (
        <Alert title="Задача не найдена" type="warning" />
      )}
    </div>
  );
}
