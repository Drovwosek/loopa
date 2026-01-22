import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { downloadExport } from "../api";
import { useAppDispatch, useAppSelector } from "../hooks";
import { clearTask, loadTask } from "../store/taskSlice";

export default function TaskPage() {
  const { id } = useParams();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const task = useAppSelector((state) => state.task.current);
  const loading = useAppSelector((state) => state.task.loading);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    dispatch(loadTask(id));
    return () => {
      dispatch(clearTask());
    };
  }, [dispatch, id]);

  useEffect(() => {
    if (!id) return;
    if (!task || (task.status !== "готово" && task.status !== "ошибка")) {
      const timer = setInterval(() => dispatch(loadTask(id)), 2500);
      return () => clearInterval(timer);
    }
  }, [dispatch, id, task]);

  const handleCopy = async () => {
    if (!task?.transcriptText) return;
    await navigator.clipboard.writeText(task.transcriptText);
  };

  const handleDownload = async (format: "txt" | "docx") => {
    if (!id) return;
    setError(null);
    try {
      await downloadExport(id, format);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Export failed");
    }
  };

  if (!id) {
    return (
      <div>
        <p>Task id is missing.</p>
        <button onClick={() => navigate("/")}>Back</button>
      </div>
    );
  }

  if (loading && !task) {
    return <p>Loading...</p>;
  }

  return (
    <div>
      <button onClick={() => navigate("/")}>Back</button>
      <h2 style={{ marginTop: 16 }}>Task details</h2>
      {task ? (
        <>
          <p>
            <strong>File:</strong> {task.originalName}
          </p>
          <p>
            <strong>Status:</strong> {task.status}
          </p>
          {task.errorMessage && (
            <p style={{ color: "crimson" }}>{task.errorMessage}</p>
          )}
          {task.transcriptText && (
            <div>
              <h3>Transcript</h3>
              <textarea
                readOnly
                value={task.transcriptText}
                style={{ width: "100%", height: 240 }}
              />
              <div style={{ marginTop: 12 }}>
                <button onClick={handleCopy} style={{ marginRight: 8 }}>
                  Copy all
                </button>
                <button onClick={() => handleDownload("txt")} style={{ marginRight: 8 }}>
                  Download TXT
                </button>
                <button onClick={() => handleDownload("docx")}>
                  Download DOCX
                </button>
              </div>
            </div>
          )}
        </>
      ) : (
        <p>Task not found.</p>
      )}
      {error && <p style={{ color: "crimson" }}>{error}</p>}
    </div>
  );
}
