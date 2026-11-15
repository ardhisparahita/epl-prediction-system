import {
  BrainCircuit,
  ChartNoAxesCombined,
  Database,
  Trophy,
} from "lucide-react";

const statistics = [
  {
    label: "Prediction Models",
    value: "2",
    description: "Logistic Regression & Random Forest",
    icon: BrainCircuit,
  },
  {
    label: "Result Classes",
    value: "3",
    description: "Home Win, Draw, Away Win",
    icon: Trophy,
  },
  {
    label: "Historical Data",
    value: "2018–2026",
    description: "EPL match history",
    icon: Database,
  },
  {
    label: "Evaluation",
    value: "Macro F1",
    description: "Model performance measurement",
    icon: ChartNoAxesCombined,
  },
];

export default function DashboardPage() {
  return (
    <div>
      <div className="mb-8">
        <p className="mb-2 text-sm font-medium text-emerald-400">OVERVIEW</p>
        <h2 className="text-3xl font-bold">Dashboard</h2>
        <p className="mt-2 text-slate-400">
          Welcome to the English Premier League Match Prediction System.
        </p>
      </div>

      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {statistics.map(({ label, value, description, icon: Icon }) => (
          <article
            key={label}
            className="rounded-xl border border-slate-800 bg-slate-900 p-5"
          >
            <div className="mb-5 flex items-center justify-between">
              <p className="text-sm text-slate-400">{label}</p>
              <Icon className="text-emerald-400" size={21} />
            </div>

            <h3 className="text-2xl font-bold">{value}</h3>
            <p className="mt-2 text-xs text-slate-500">{description}</p>
          </article>
        ))}
      </section>

      <section className="mt-6 rounded-xl border border-slate-800 bg-slate-900 p-6">
        <h3 className="text-lg font-semibold">About This System</h3>

        <p className="mt-3 max-w-3xl text-sm leading-7 text-slate-400">
          This application predicts English Premier League match results using
          historical team performance and two classification algorithms:
          Logistic Regression and Random Forest.
        </p>

        <div className="mt-6 grid gap-4 md:grid-cols-2">
          <div className="rounded-lg bg-slate-800/70 p-4">
            <h4 className="font-medium text-emerald-400">
              Logistic Regression
            </h4>
            <p className="mt-2 text-sm text-slate-400">
              A classification model used to estimate the probability of each
              match result.
            </p>
          </div>

          <div className="rounded-lg bg-slate-800/70 p-4">
            <h4 className="font-medium text-emerald-400">Random Forest</h4>
            <p className="mt-2 text-sm text-slate-400">
              An ensemble learning model used to classify match outcomes based
              on historical features.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
