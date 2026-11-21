import { useEffect, useRef, useState, type FormEvent } from "react";
import {
  AlertCircle,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Filter,
  LoaderCircle,
  RefreshCw,
  Search,
  X,
} from "lucide-react";

import {
  getMatches,
  type Match,
  type MatchFilters,
  type MatchPagination,
} from "../services/matchService";

const PAGE_SIZE = 20;

function formatDate(date: string) {
  const parsedDate = new Date(`${date}T00:00:00`);

  if (Number.isNaN(parsedDate.getTime())) {
    return date;
  }

  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(parsedDate);
}

function formatShots(shots: number | null) {
  return shots === null ? "—" : shots;
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
    // Use the fallback below if the native picker cannot be opened.
  }

  input.focus();
  input.click();
}

function MatchRow({ match }: { match: Match }) {
  return (
    <tr className="border-t border-slate-800 transition hover:bg-slate-800/40">
      <td className="whitespace-nowrap px-5 py-4 text-sm text-slate-400">
        {formatDate(match.date)}
      </td>

      <td className="px-5 py-4">
        <div className="flex items-center justify-end gap-2 font-medium text-white">
          {match.home_team}
        </div>
      </td>

      <td className="whitespace-nowrap px-5 py-4 text-center">
        <span className="rounded-lg border border-slate-700 bg-slate-950 px-3 py-1.5 font-bold tabular-nums text-white">
          {match.home_goals} - {match.away_goals}
        </span>
      </td>

      <td className="px-5 py-4 font-medium text-white">{match.away_team}</td>

      <td className="px-5 py-4 text-center text-sm tabular-nums text-slate-400">
        {formatShots(match.home_shots_on_target)}
      </td>

      <td className="px-5 py-4 text-center text-sm tabular-nums text-slate-400">
        {formatShots(match.away_shots_on_target)}
      </td>
    </tr>
  );
}

export default function MatchHistoryPage() {
  const [team, setTeam] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const fromDateRef = useRef<HTMLInputElement>(null);
  const toDateRef = useRef<HTMLInputElement>(null);

  const [appliedFilters, setAppliedFilters] = useState<MatchFilters>({});
  const [page, setPage] = useState(1);

  const [matches, setMatches] = useState<Match[]>([]);
  const [pagination, setPagination] = useState<MatchPagination | null>(null);

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [reloadKey, setReloadKey] = useState(0);

  // Fetch match history whenever filters or page changes.
  useEffect(() => {
    const controller = new AbortController();

    async function fetchMatches() {
      setIsLoading(true);
      setError("");
      setMatches([]);
      setPagination(null);

      try {
        const result = await getMatches(
          {
            ...appliedFilters,
            page,
            limit: PAGE_SIZE,
          },
          controller.signal,
        );

        setMatches(result.data);
        setPagination(result.pagination);
      } catch (err) {
        if (err instanceof Error && err.name === "AbortError") {
          return;
        }

        setError(
          err instanceof Error
            ? err.message
            : "Terjadi kesalahan saat mengambil riwayat pertandingan.",
        );
      } finally {
        if (!controller.signal.aborted) {
          setIsLoading(false);
        }
      }
    }

    fetchMatches();

    return () => controller.abort();
  }, [appliedFilters, page, reloadKey]);

  // Apply selected filters.
  function handleFilterSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (from && to && from > to) {
      setError("Tanggal awal tidak boleh melebihi tanggal akhir.");
      return;
    }

    setError("");
    setPage(1);

    setAppliedFilters({
      team: team.trim(),
      from,
      to,
    });
  }

  // Reset all filters and return to the first page.
  function handleReset() {
    setTeam("");
    setFrom("");
    setTo("");
    setPage(1);
    setAppliedFilters({});
    setError("");
  }

  const total = pagination?.total ?? 0;
  const totalPages = pagination?.total_pages ?? 0;

  const startItem = total === 0 ? 0 : (page - 1) * PAGE_SIZE + 1;
  const endItem = Math.min(page * PAGE_SIZE, total);

  return (
    <div>
      {/* Page Header */}
      <div className="mb-8 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <p className="mb-2 text-sm font-medium tracking-wider text-emerald-400">
            MATCH DATABASE
          </p>

          <h2 className="text-3xl font-bold">Match History</h2>

          <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">
            Browse historical Premier League matches, review match results, and
            filter matches by team or date.
          </p>
        </div>

        <button
          type="button"
          onClick={() => setReloadKey((current) => current + 1)}
          disabled={isLoading}
          className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-700 bg-slate-900 px-4 py-2.5 text-sm font-medium text-slate-300 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-50"
        >
          <RefreshCw size={16} className={isLoading ? "animate-spin" : ""} />
          Refresh
        </button>
      </div>

      {/* Summary */}
      <div className="mb-6 grid gap-4 sm:grid-cols-3">
        <div className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
          <p className="text-sm text-slate-400">Total Matches</p>

          <p className="mt-2 text-2xl font-bold tabular-nums text-white">
            {isLoading ? "—" : total.toLocaleString("id-ID")}
          </p>
        </div>

        <div className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
          <p className="text-sm text-slate-400">Current Page</p>

          <p className="mt-2 text-2xl font-bold tabular-nums text-white">
            {isLoading ? "—" : `${page} / ${totalPages}`}
          </p>
        </div>

        <div className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
          <p className="text-sm text-slate-400">Displayed Matches</p>

          <p className="mt-2 text-2xl font-bold tabular-nums text-white">
            {isLoading ? "—" : matches.length}
          </p>
        </div>
      </div>

      {/* Filters */}
      <section className="mb-6 rounded-2xl border border-slate-800 bg-slate-900 p-6">
        <div className="mb-5 flex items-center gap-3">
          <div className="rounded-xl bg-emerald-500/10 p-3 text-emerald-400">
            <Filter size={21} />
          </div>

          <div>
            <h3 className="font-semibold">Match Filters</h3>

            <p className="text-xs text-slate-400">
              Narrow down the historical match records
            </p>
          </div>
        </div>

        <form onSubmit={handleFilterSubmit}>
          <div className="grid gap-4 md:grid-cols-3">
            {/* Team Filter */}
            <div>
              <label
                htmlFor="filter-team"
                className="mb-2 block text-sm font-medium text-slate-300"
              >
                Team
              </label>

              <input
                id="filter-team"
                type="text"
                value={team}
                onChange={(event) => setTeam(event.target.value)}
                placeholder="e.g. Arsenal"
                className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-sm text-white outline-none transition placeholder:text-slate-600 focus:border-emerald-500"
              />
            </div>

            {/* From Date */}
            <div>
              <label
                htmlFor="filter-from"
                className="mb-2 block text-sm font-medium text-slate-300"
              >
                From Date
              </label>

              <div className="relative">
                <input
                  ref={fromDateRef}
                  id="filter-from"
                  type="date"
                  value={from}
                  max={to || undefined}
                  onChange={(event) => setFrom(event.target.value)}
                  className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 pr-12 text-sm text-white outline-none transition focus:border-emerald-500 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:opacity-0"
                />

                <button
                  type="button"
                  aria-label="Open from date calendar"
                  title="Choose start date"
                  onClick={() => openDatePicker(fromDateRef.current)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 rounded-md p-1 text-emerald-400 transition hover:bg-slate-800 hover:text-emerald-300"
                >
                  <CalendarDays size={19} />
                </button>
              </div>
            </div>

            {/* To Date */}
            <div>
              <label
                htmlFor="filter-to"
                className="mb-2 block text-sm font-medium text-slate-300"
              >
                To Date
              </label>

              <div className="relative">
                <input
                  ref={toDateRef}
                  id="filter-to"
                  type="date"
                  value={to}
                  min={from || undefined}
                  onChange={(event) => setTo(event.target.value)}
                  className="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 pr-12 text-sm text-white outline-none transition focus:border-emerald-500 [color-scheme:dark] [&::-webkit-calendar-picker-indicator]:opacity-0"
                />

                <button
                  type="button"
                  aria-label="Open to date calendar"
                  title="Choose end date"
                  onClick={() => openDatePicker(toDateRef.current)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 rounded-md p-1 text-emerald-400 transition hover:bg-slate-800 hover:text-emerald-300"
                >
                  <CalendarDays size={19} />
                </button>
              </div>
            </div>
          </div>

          {/* Filter Actions */}
          <div className="mt-5 flex flex-wrap gap-3">
            <button
              type="submit"
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-500 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-emerald-400"
            >
              <Search size={17} />
              Apply Filters
            </button>

            <button
              type="button"
              onClick={handleReset}
              className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-700 px-5 py-3 text-sm font-medium text-slate-300 transition hover:bg-slate-800"
            >
              <X size={17} />
              Reset
            </button>
          </div>
        </form>
      </section>

      {/* Error State */}
      {error && (
        <div
          role="alert"
          className="mb-6 flex items-start gap-3 rounded-xl border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-300"
        >
          <AlertCircle size={19} className="mt-0.5 shrink-0" />
          <p>{error}</p>
        </div>
      )}

      {/* Match Table */}
      <section className="overflow-hidden rounded-2xl border border-slate-800 bg-slate-900">
        <div className="flex flex-col justify-between gap-2 border-b border-slate-800 px-6 py-5 sm:flex-row sm:items-center">
          <div>
            <h3 className="font-semibold">Historical Results</h3>

            <p className="mt-1 text-xs text-slate-400">
              Match results and shots on target
            </p>
          </div>

          {!isLoading && !error && (
            <p className="text-xs text-slate-400">
              Showing {startItem.toLocaleString("id-ID")}–
              {endItem.toLocaleString("id-ID")} of{" "}
              {total.toLocaleString("id-ID")} matches
            </p>
          )}
        </div>

        <div className="overflow-x-auto">
          <table className="w-full min-w-[850px] border-collapse text-left">
            <thead className="bg-slate-950/70">
              <tr className="text-xs uppercase tracking-wider text-slate-500">
                <th className="px-5 py-4 font-semibold">Date</th>

                <th className="px-5 py-4 text-right font-semibold">
                  Home Team
                </th>

                <th className="px-5 py-4 text-center font-semibold">Score</th>

                <th className="px-5 py-4 font-semibold">Away Team</th>

                <th className="px-5 py-4 text-center font-semibold">
                  Home SOT
                </th>

                <th className="px-5 py-4 text-center font-semibold">
                  Away SOT
                </th>
              </tr>
            </thead>

            <tbody>
              {/* Loading State */}
              {isLoading && (
                <tr>
                  <td colSpan={6} className="px-5 py-16 text-center">
                    <div className="flex flex-col items-center gap-3 text-slate-400">
                      <LoaderCircle
                        size={27}
                        className="animate-spin text-emerald-400"
                      />

                      <p className="text-sm">Loading match history...</p>
                    </div>
                  </td>
                </tr>
              )}

              {/* Match Data */}
              {!isLoading &&
                !error &&
                matches.length > 0 &&
                matches.map((match) => (
                  <MatchRow key={match.id} match={match} />
                ))}

              {/* Empty State */}
              {!isLoading && !error && matches.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-5 py-16 text-center">
                    <div className="text-slate-400">
                      <p className="font-medium text-white">No matches found</p>

                      <p className="mt-1 text-sm">Try changing your filters.</p>
                    </div>
                  </td>
                </tr>
              )}

              {/* Error Table State */}
              {!isLoading && error && (
                <tr>
                  <td
                    colSpan={6}
                    className="px-5 py-12 text-center text-sm text-slate-500"
                  >
                    Match data could not be displayed.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        <div className="flex flex-col justify-between gap-4 border-t border-slate-800 px-5 py-4 sm:flex-row sm:items-center">
          <p className="text-sm text-slate-400">
            Page <span className="font-semibold text-white">{page}</span> of{" "}
            <span className="font-semibold text-white">{totalPages}</span>
          </p>

          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setPage((current) => Math.max(1, current - 1))}
              disabled={isLoading || page <= 1 || !!error}
              className="inline-flex items-center gap-2 rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-300 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-40"
            >
              <ChevronLeft size={17} />
              Previous
            </button>

            <button
              type="button"
              onClick={() => setPage((current) => current + 1)}
              disabled={
                isLoading || page >= totalPages || totalPages === 0 || !!error
              }
              className="inline-flex items-center gap-2 rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-300 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-40"
            >
              Next
              <ChevronRight size={17} />
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}
