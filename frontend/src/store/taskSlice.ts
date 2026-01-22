import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { fetchTask, TaskResponse } from "../api";

type TaskState = {
  current?: TaskResponse;
  loading: boolean;
  error?: string;
};

const initialState: TaskState = {
  loading: false,
};

export const loadTask = createAsyncThunk("task/load", async (taskId: string) => {
  return await fetchTask(taskId);
});

const taskSlice = createSlice({
  name: "task",
  initialState,
  reducers: {
    clearTask: (state) => {
      state.current = undefined;
      state.error = undefined;
    },
  },
  extraReducers: (builder) => {
    builder.addCase(loadTask.pending, (state) => {
      state.loading = true;
      state.error = undefined;
    });
    builder.addCase(loadTask.fulfilled, (state, action) => {
      state.loading = false;
      state.current = action.payload;
    });
    builder.addCase(loadTask.rejected, (state, action) => {
      state.loading = false;
      state.error = action.error.message ?? "Failed to load task";
    });
  },
});

export const { clearTask } = taskSlice.actions;
export default taskSlice.reducer;
