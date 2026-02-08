import type { Segment, Project } from "./types";

const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api";

export type TaskResponse = {
  id: string;
  status: string;
  originalName: string;
  transcriptText?: string;
  errorMessage?: string;
  createdAt: string;
  completedAt?: string;
  segments?: Segment[];
  numSpeakers?: number;
};

export type HistoryItem = {
  id: string;
  originalName: string;
  status: string;
  uploadedAt: string;
};

export type ProjectFileItem = {
  fileId: string;
  originalName: string;
  uploadedAt: string;
  taskId?: string;
  status?: string;
};

export async function uploadFile(file: File, projectId?: string): Promise<string> {
  const form = new FormData();
  form.append("file", file);
  if (projectId) {
    form.append("projectId", projectId);
  }

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

export async function fetchSegments(taskId: string): Promise<Segment[]> {
  const res = await fetch(`${API_BASE}/tasks/${taskId}/segments`, {
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to load segments");
  }
  return (await res.json()) as Segment[];
}

export async function updateSegment(
  taskId: string,
  segmentId: string,
  text: string
): Promise<void> {
  const res = await fetch(`${API_BASE}/tasks/${taskId}/segments/${segmentId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ text }),
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to update segment");
  }
}

export async function updateSpeaker(
  taskId: string,
  speakerId: string,
  name: string
): Promise<void> {
  const res = await fetch(
    `${API_BASE}/tasks/${taskId}/speakers/${speakerId}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
      credentials: "include",
    }
  );
  if (!res.ok) {
    throw new Error("Failed to update speaker");
  }
}

export function getAudioUrl(taskId: string): string {
  return `${API_BASE}/tasks/${taskId}/audio`;
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
    const match = disposition.match(/filename="(.+)"/);
    if (match?.[1]) filename = match[1];
  }
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

// --- Projects API ---

export async function fetchProjects(): Promise<Project[]> {
  const res = await fetch(`${API_BASE}/projects`, {
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to load projects");
  }
  return (await res.json()) as Project[];
}

export async function createProject(
  name: string,
  description?: string
): Promise<Project> {
  const res = await fetch(`${API_BASE}/projects`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, description }),
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to create project");
  }
  return (await res.json()) as Project;
}

export async function fetchProjectFiles(projectId: string): Promise<ProjectFileItem[]> {
  const res = await fetch(`${API_BASE}/projects/${projectId}/files`, {
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to load project files");
  }
  return (await res.json()) as ProjectFileItem[];
}

export async function deleteProject(projectId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/projects/${projectId}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Failed to delete project");
  }
}

async function safeJson(res: Response) {
  try {
    return await res.json();
  } catch {
    return null;
  }
}
