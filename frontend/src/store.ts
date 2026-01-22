import { configureStore } from "@reduxjs/toolkit";
import historyReducer from "./store/historySlice";
import taskReducer from "./store/taskSlice";

export const store = configureStore({
  reducer: {
    history: historyReducer,
    task: taskReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
