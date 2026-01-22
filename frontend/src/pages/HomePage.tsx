import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { deleteTask, uploadFile } from "../api";
import { useAppDispatch, useAppSelector } from "../hooks";
import { loadHistory } from "../store/historySlice";

export default function HomePage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const history = useAppSelector((state) => state.history.items);
  const loading = useAppSelector((state) => state.history.loading);
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dragOver, setDragOver] = useState(false);

  useEffect(() => {
    dispatch(loadHistory());
  }, [dispatch]);

  const fileName = useMemo(() => file?.name ?? "No file selected", [file]);

  const handleUpload = async () => {
    if (!file) {
      setError("Select a file first.");
      return;
    }
    setUploading(true);
    setError(null);
    try {
      const taskId = await uploadFile(file);
      navigate(`/tasks/${taskId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async (taskId: string) => {
    try {
      await deleteTask(taskId);
      dispatch(loadHistory());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Delete failed");
    }
  };

  const onDrop = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragOver(false);
    const dropped = event.dataTransfer.files[0];
    if (dropped) {
      setFile(dropped);
    }
  };

  return (
    <div>
      <section
        onDragOver={(event) => {
          event.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={onDrop}
        style={{
          border: `2px dashed ${dragOver ? "#333" : "#999"}`,
          padding: 24,
          borderRadius: 8,
          marginBottom: 24,
        }}
      >
        <h2>Upload media</h2>
        <p>MP3, WAV, MP4, MOV up to 1GB</p>
        <input
          type="file"
          accept=".mp3,.wav,.mp4,.mov,audio/*,video/*"
          onChange={(event) => setFile(event.target.files?.[0] ?? null)}
        />
        <p style={{ marginTop: 8 }}>{fileName}</p>
        <button onClick={handleUpload} disabled={uploading}>
          {uploading ? "Uploading..." : "Upload"}
        </button>
        {error && <p style={{ color: "crimson" }}>{error}</p>}
      </section>

      <section>
        <h2>Recent uploads</h2>
        {loading ? (
          <p>Loading...</p>
        ) : history.length === 0 ? (
          <p>No uploads yet.</p>
        ) : (
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th style={{ textAlign: "left" }}>File</th>
                <th style={{ textAlign: "left" }}>Status</th>
                <th style={{ textAlign: "left" }}>Uploaded</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {history.map((item) => (
                <tr key={item.id}>
                  <td>{item.originalName}</td>
                  <td>{item.status}</td>
                  <td>{new Date(item.uploadedAt).toLocaleString()}</td>
                  <td>
                    <button
                      onClick={() => navigate(`/tasks/${item.id}`)}
                      style={{ marginRight: 8 }}
                    >
                      Open
                    </button>
                    <button onClick={() => handleDelete(item.id)}>Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  );
}
