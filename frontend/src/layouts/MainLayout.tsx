import {
  Activity,
  BrainCircuit,
  ChartNoAxesCombined,
  Database,
} from "lucide-react";
import { NavLink, Outlet } from "react-router";

const navigation = [
  { label: "Dashboard", path: "/", icon: Activity, end: true },
  { label: "Predictions", path: "/predictions", icon: BrainCircuit },
  { label: "Match History", path: "/matches", icon: Database },
  {
    label: "Model Evaluation",
    path: "/evaluation",
    icon: ChartNoAxesCombined,
  },
];

export default function MainLayout() {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 md:flex">
      <aside className="border-b border-slate-800 bg-slate-900 p-5 md:min-h-screen md:w-64 md:border-r md:border-b-0">
        <div className="mb-8 flex items-center gap-3">
          <div className="rounded-xl bg-emerald-500 p-2 text-slate-950">
            <Activity size={22} />
          </div>
          <div>
            <h1 className="font-bold">EPL Prediction</h1>
            <p className="text-xs text-slate-400">Football Analytics</p>
          </div>
        </div>

        <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-slate-500">
          Main Menu
        </p>

        <nav className="flex gap-2 overflow-x-auto md:flex-col">
          {navigation.map(({ label, path, icon: Icon, end }) => (
            <NavLink
              key={path}
              to={path}
              end={end}
              className={({ isActive }) =>
                `flex shrink-0 items-center gap-3 rounded-lg px-4 py-3 text-sm transition ${
                  isActive
                    ? "bg-emerald-500 font-semibold text-slate-950"
                    : "text-slate-400 hover:bg-slate-800 hover:text-white"
                }`
              }
            >
              <Icon size={19} />
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="min-w-0 flex-1">
        <header className="flex items-center justify-between border-b border-slate-800 px-6 py-5">
          <div>
            <p className="text-sm font-medium">English Premier League</p>
            <p className="text-xs text-slate-500">Season 2026/2027</p>
          </div>

          <span className="rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3 py-1.5 text-xs text-emerald-400">
            Development
          </span>
        </header>

        <main className="mx-auto max-w-7xl p-6 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
