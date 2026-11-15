import { BrowserRouter, Route, Routes } from "react-router";
import MainLayout from "./layouts/MainLayout";
import DashboardPage from "./pages/DashboardPage";
import PredictionsPage from "./pages/PredictionsPage";

function PlaceholderPage({ title }: { title: string }) {
  return (
    <div>
      <h2 className="text-3xl font-bold">{title}</h2>
      <p className="mt-2 text-slate-400">
        This module will be developed in the next stage.
      </p>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<MainLayout />}>
          <Route index element={<DashboardPage />} />

          <Route path="predictions" element={<PredictionsPage />} />

          <Route
            path="matches"
            element={<PlaceholderPage title="Match History" />}
          />

          <Route
            path="evaluation"
            element={<PlaceholderPage title="Model Evaluation" />}
          />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
