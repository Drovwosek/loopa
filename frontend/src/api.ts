const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api";

export type TaskResponse = {
  id: string;
  status: string;
  originalName: string;
  transcriptText?: string;
  errorMessage?: string;
  createdAt: string;
  completedAt?: string;
};

export type HistoryItem = {
  id: string;
  originalName: string;
  status: string;
  uploadedAt: string;
};

export async function uploadFile(file: File): Promise<string> {
  const form = new FormData();
  form.append("file", file);

  const res = await fetch(`${API_BASE}/uploads`, {
    method: "POST",
    body: form,
    credentials: "include",
  });
  if (!res.ok) {
    const data = await safeJson(res);
    throw new Error(data?.error ?? "Upload failed");
  }
  const data = (await res.json()) as { taskId: string };
  return data.taskId;
}

export async function fetchTask(taskId: string): Promise<TaskResponse> {
  const res = await fetch(`${API_BASE}/tasks/${taskId}`, {
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to load task");
  }
  return (await res.json()) as TaskResponse;
}

export async function fetchHistory(): Promise<HistoryItem[]> {
  const res = await fetch(`${API_BASE}/history`, {
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to load history");
  }
  return (await res.json()) as HistoryItem[];
}

export async function deleteTask(taskId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/tasks/${taskId}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to delete task");
  }
}

export async function downloadExport(taskId: string, format: "txt" | "docx") {
  const res = await fetch(`${API_BASE}/tasks/${taskId}/export?format=${format}`, {
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Export failed");
  }
  const blob = await res.blob();
  const disposition = res.headers.get("content-disposition");
  let filename = `transcript.${format}`;
  if (disposition) {
    const match = disposition.match(/filename=\"(.+)\"/);
    if (match?.[1]) filename = match[1];
  }
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

async function safeJson(res: Response) {
  try {
    return await res.json();
  } catch {
    return null;
  }
}
