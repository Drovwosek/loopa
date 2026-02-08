import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { fetchProjects, createProject, deleteProject } from "../api";
import type { Project } from "../types";

type ProjectState = {
  items: Project[];
  loading: boolean;
  error?: string;
};

const initialState: ProjectState = {
  items: [],
  loading: false,
};

export const loadProjects = createAsyncThunk("projects/load", async () => {
  return await fetchProjects();
});

export const addProject = createAsyncThunk(
  "projects/add",
  async ({ name, description }: { name: string; description?: string }) => {
    return await createProject(name, description);
  }
);

export const removeProject = createAsyncThunk(
  "projects/remove",
  async (projectId: string) => {
    await deleteProject(projectId);
    return projectId;
  }
);

const projectSlice = createSlice({
  name: "projects",
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(loadProjects.pending, (state) => {
      state.loading = true;
      state.error = undefined;
    });
    builder.addCase(loadProjects.fulfilled, (state, action) => {
      state.loading = false;
      state.items = action.payload;
    });
    builder.addCase(loadProjects.rejected, (state, action) => {
      state.loading = false;
      state.error = action.error.message ?? "Failed to load projects";
    });
    builder.addCase(addProject.fulfilled, (state, action) => {
      state.items.unshift(action.payload);
    });
    builder.addCase(removeProject.fulfilled, (state, action) => {
      state.items = state.items.filter((p) => p.id !== action.payload);
    });
  },
});

export default projectSlice.reducer;
