import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { fetchHistory, HistoryItem } from "../api";

type HistoryState = {
  items: HistoryItem[];
  loading: boolean;
  error?: string;
};

const initialState: HistoryState = {
  items: [],
  loading: false,
};

export const loadHistory = createAsyncThunk("history/load", async () => {
  return await fetchHistory();
});

const historySlice = createSlice({
  name: "history",
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(loadHistory.pending, (state) => {
      state.loading = true;
      state.error = undefined;
    });
    builder.addCase(loadHistory.fulfilled, (state, action) => {
      state.loading = false;
      state.items = action.payload;
    });
    builder.addCase(loadHistory.rejected, (state, action) => {
      state.loading = false;
      state.error = action.error.message ?? "Failed to load history";
    });
  },
});

export default historySlice.reducer;
