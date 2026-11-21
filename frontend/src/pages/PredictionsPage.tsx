import { useRef, useState, type FormEvent } from "react";
import {
  AlertCircle,
  BrainCircuit,
  CalendarDays,
  LoaderCircle,
  ShieldCheck,
  Swords,
} from "lucide-react";

import {
  createPrediction,
  type ModelPrediction,
  type Outcome,
  type PredictionResponse,
} from "../services/predictionService";

const teams = [
  { label: "AFC Bournemouth", value: "Bournemouth" },
  { label: "Arsenal", value: "Arsenal" },
  { label: "Aston Villa", value: "Aston Villa" },
  { label: "Brentford", value: "Brentford" },
  { label: "Brighton & Hove Albion", value: "Brighton" },
  { label: "Chelsea", value: "Chelsea" },
  { label: "Coventry City", value: "Coventry" },
  { label: "Crystal Palace", value: "Crystal Palace" },
  { label: "Everton", value: "Everton" },
  { label: "Fulham", value: "Fulham" },
  { label: "Hull City", value: "Hull City" },
  { label: "Ipswich Town", value: "Ipswich" },
  { label: "Leeds United", value: "Leeds" },
  { label: "Liverpool", value: "Liverpool" },
  { label: "Manchester City", value: "Man City" },
  { label: "Manchester United", value: "Man United" },
  { label: "Newcastle United", value: "Newcastle" },
  { label: "Nottingham Forest", value: "Nott'm Forest" },
  { label: "Sunderland", value: "Sunderland" },
  { label: "Tottenham Hotspur", value: "Tottenham" },
];

const outcomes: { key: Outcome; label: string }[] = [
  { key: "H", label: "Home Win" },
  { key: "D", label: "Draw" },
  { key: "A", label: "Away Win" },
];

function getOutcomeLabel(outcome: Outcome) {
  return outcomes.find((item) => item.key === outcome)?.label ?? outcome;
}

// Open the browser's native date picker.
function openDatePicker(input: HTMLInputElement | null) {
  if (!input) return;

  try {
    if (typeof input.showPicker === "function") {
      input.showPicker();
      return;
    }
  } catch {
    // Use the fallback if the native picker cannot be opened.
  }

  input.focus();
  input.click();
}

function ModelCard({
  title,
  description,
  model,
}: {
  title: string;
  description: string;
  model: ModelPrediction;
}) {
  return (
    <article className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h3 className="text-lg font-bold">{title}</h3>
          <p className="mt-1 text-sm text-slate-400">{description}</p>
        </div>

        <div className="rounded-xl bg-emerald-500/10 p-3 text-emerald-400">
          <BrainCircuit size={22} />
        </div>
      </div>

      <div className="mt-6 rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-4">
        <p className="text-xs font-medium uppercase tracking-wider text-slate-400">
          Predicted Result
        </p>

        <p className="mt-2 text-2xl font-bold text-emerald-400">
          {getOutcomeLabel(model.prediction)}
        </p>
      </div>

      <div className="mt-6 space-y-5">
        <h4 className="text-sm font-semibold text-slate-300">
          Result Probabilities
        </h4>

        {outcomes.map(({ key, label }) => {
          const percentage = model.probabilities[key];

          return (
            <div key={key}>
              <div className="mb-2 flex justify-between text-sm">
                <span className="text-slate-400">{label}</span>

                <span className="font-semibold tabular-nums">
                  {percentage.toFixed(2)}%
                </span>
              </div>

              <div className="h-2 overflow-hidden rounded-full bg-slate-800">
                <div
                  className={`h-full rounded-full transition-all ${
                    key === "H"
                      ? "bg-emerald-400"
                      : key === "D"
                        ? "bg-amber-400"
                        : "bg-sky-400"
                  }`}
                  style={{
                    width: `${Math.max(0, Math.min(100, percentage))}%`,
                  }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </article>
  );
}

export default function PredictionsPage() {
  const [homeTeam, setHomeTeam] = useState("");
  const [awayTeam, setAwayTeam] = useState("");
  const [kickoffDate, setKickoffDate] = useState("");
  const [result, setResult] = useState<PredictionResponse | null>(null);
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  // Ref for opening the native date picker.
  const kickoffDateRef = useRef<HTMLInputElement>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setResult(null);

    if (homeTeam === awayTeam) {
      setError("Tim kandang dan tim tandang harus berbeda.");
      return;
    }

    setIsLoading(true);

    try {
      const prediction = await createPrediction({
        home_team: homeTeam,
        away_team: awayTeam,
        kickoff_date: kickoffDate,
      });

      setResult(prediction);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Terjadi kesalahan saat melakukan prediksi.",
      );
    } finally {
      setIsLoading(false);
    }
  }

  const selectedHome = teams.find((team) => team.value === homeTeam)?.label;
  const selectedAway = teams.find((team) => team.value === awayTeam)?.label;

  return (
    <div>
      {/* Page Header */}
      <div className="mb-8">
        <p className="mb-2 text-sm font-medium tracking-wider text-emerald-400">
          MATCH ANALYSIS
        </p>

        <h2 className="text-3xl font-bold">Match Predictions</h2>

        <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">
          Select two Premier League clubs and a match date to estimate the match
          result using Logistic Regression and Random Forest.
        </p>
      </div>

      {/* Match Configuration */}
      <section className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
        <div className="mb-6 flex items-center gap-3">
          <div className="rounded-xl bg-emerald-500/10 p-3 text-emerald-400">
            <Swords size={22} />
          </div>

          <div>
            <h3 className="font-semibold">Match Configuration</h3>
            <p className="text-xs text-slate-400">Enter the fixture details</p>
          </div>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="grid gap-5 md:grid-cols-2">
            {/* Home Team */}
            <div>
              <label
                htmlFor="home-team"
                className="mb-2 block text-sm font-medium text-slate-300"
              >
                Home Team
              </label>

              <select
                id="home-team"
                required
                value={homeTeam}
                onChange={(event) => {
                  setHomeTeam(event.target.value);
                  setResult(null);
                  setError("");
                }}
                className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-white outline-none transition focus:border-emerald-500"
              >
                <option value="">Select home team</option>

                {teams.map((team) => (
                  <option key={team.value} value={team.value}>
                    {team.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Away Team */}
            <div>
              <label
                htmlFor="away-team"
                className="mb-2 block text-sm font-medium text-slate-300"
              >
                Away Team
              </label>

              <select
                id="away-team"
                required
                value={awayTeam}
                onChange={(event) => {
                  setAwayTeam(event.target.value);
                  setResult(null);
                  setError("");
                }}
                className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-white outline-none transition focus:border-emerald-500"
              >
                <option value="">Select away team</option>

                {teams.map((team) => (
                  <option
                    key={team.value}
                    value={team.value}
                    disabled={team.value === homeTeam}
                  >
                    {team.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Match Date */}
            <div className="md:col-span-2">
              <label
                htmlFor="kickoff-date"
                className="mb-2 block text-sm font-medium text-slate-300"
              >
                Match Date
              </label>

              <div className="relative">
                <input
                  ref={kickoffDateRef}
                  id="kickoff-date"
                  type="date"
                  required
                  value={kickoffDate}
                  onChange={(event) => {
                    setKickoffDate(event.target.value);
                    setResult(null);
                    setError("");
                  }}
                  className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 pr-12 text-sm text-white outline-none transition focus:border-emerald-500 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:opacity-0"
                />

                <button
                  type="button"
                  aria-label="Open match date calendar"
                  title="Choose match date"
                  onClick={() => openDatePicker(kickoffDateRef.current)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 rounded-md p-1 text-emerald-400 transition hover:bg-slate-800 hover:text-emerald-300"
                >
                  <CalendarDays size={19} />
                </button>
              </div>
            </div>
          </div>

          {/* Information */}
          <div className="mt-5 flex items-start gap-2 rounded-xl bg-slate-800/60 p-4 text-xs leading-5 text-slate-400">
            <ShieldCheck
              size={17}
              className="mt-0.5 shrink-0 text-emerald-400"
            />

            <p>
              Prediction uses historical matches before the selected match date.
              At least three previous matches are required for each team.
            </p>
          </div>

          {/* Error Message */}
          {error && (
            <div
              role="alert"
              className="mt-5 flex items-start gap-3 rounded-xl border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-300"
            >
              <AlertCircle size={19} className="mt-0.5 shrink-0" />
              <p>{error}</p>
            </div>
          )}

          {/* Submit Button */}
          <button
            type="submit"
            disabled={isLoading}
            className="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 px-6 py-3 font-semibold text-slate-950 transition hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-60 md:w-auto"
          >
            {isLoading ? (
              <>
                <LoaderCircle size={18} className="animate-spin" />
                Processing Prediction...
              </>
            ) : (
              <>
                <BrainCircuit size={18} />
                Generate Prediction
              </>
            )}
          </button>
        </form>
      </section>

      {/* Prediction Results */}
      {result && (
        <section className="mt-8">
          <div className="mb-6 flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
            <div>
              <p className="mb-2 text-sm font-medium text-emerald-400">
                PREDICTION RESULT
              </p>

              <h3 className="text-2xl font-bold">
                {selectedHome} <span className="text-slate-500">vs</span>{" "}
                {selectedAway}
              </h3>

              <p className="mt-2 text-sm text-slate-400">
                Match date: {kickoffDate}
              </p>
            </div>

            <div className="rounded-lg border border-slate-800 bg-slate-900 px-4 py-2 text-sm text-slate-400">
              Features used:{" "}
              <span className="font-semibold text-white">
                {result.featureCount ?? "—"}
              </span>
            </div>
          </div>

          <div className="grid gap-5 xl:grid-cols-2">
            <ModelCard
              title="Logistic Regression"
              description="Probability-based classification"
              model={result.logisticRegression}
            />

            <ModelCard
              title="Random Forest"
              description="Ensemble-based classification"
              model={result.randomForest}
            />
          </div>
        </section>
      )}
    </div>
  );
}
