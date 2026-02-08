import { configureStore } from "@reduxjs/toolkit";
import historyReducer from "./store/historySlice";
import taskReducer from "./store/taskSlice";
import projectReducer from "./store/projectSlice";

export const store = configureStore({
  reducer: {
    history: historyReducer,
    task: taskReducer,
    projects: projectReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
