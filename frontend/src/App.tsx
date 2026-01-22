import { Link, Route, Routes } from "react-router-dom";
import HomePage from "./pages/HomePage";
import TaskPage from "./pages/TaskPage";

export default function App() {
  return (
    <div style={{ maxWidth: 960, margin: "0 auto", padding: 24 }}>
      <header style={{ marginBottom: 24 }}>
        <Link to="/" style={{ textDecoration: "none", color: "#111" }}>
          <h1>Loopa</h1>
        </Link>
      </header>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/tasks/:id" element={<TaskPage />} />
      </Routes>
    </div>
  );
}
