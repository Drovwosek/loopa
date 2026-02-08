import { Route, Routes } from "react-router-dom";
import { ConfigProvider } from "antd";
import ruRU from "antd/locale/ru_RU";
import AppLayout from "./components/layout/AppLayout";
import HomePage from "./pages/HomePage";
import TaskPage from "./pages/TaskPage";
import ProjectPage from "./pages/ProjectPage";
import ProjectDetailPage from "./pages/ProjectDetailPage";

export default function App() {
  return (
    <ConfigProvider locale={ruRU}>
      <AppLayout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/tasks/:id" element={<TaskPage />} />
          <Route path="/projects" element={<ProjectPage />} />
          <Route path="/projects/:id" element={<ProjectDetailPage />} />
        </Routes>
      </AppLayout>
    </ConfigProvider>
  );
}
