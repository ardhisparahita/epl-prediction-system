export interface Match {
  id: number;
  date: string;
  home_team: string;
  away_team: string;
  home_goals: number;
  away_goals: number;
  home_shots_on_target: number | null;
  away_shots_on_target: number | null;
}

export interface MatchPagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface MatchListResponse {
  data: Match[];
  pagination: MatchPagination;
}

export interface MatchFilters {
  team?: string;
  from?: string;
  to?: string;
  page?: number;
  limit?: number;
}

export async function getMatches(
  filters: MatchFilters,
  signal?: AbortSignal,
): Promise<MatchListResponse> {
  const params = new URLSearchParams();

  if (filters.team?.trim()) {
    params.set("team", filters.team.trim());
  }

  if (filters.from) params.set("from", filters.from);
  if (filters.to) params.set("to", filters.to);

  params.set("page", String(filters.page ?? 1));
  params.set("limit", String(filters.limit ?? 20));

  let response: Response;

  try {
    response = await fetch(`/api/v1/matches?${params.toString()}`, {
      method: "GET",
      signal,
    });
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw error;
    }

    throw new Error(
      "Tidak dapat terhubung ke backend. Pastikan Go API berjalan di port 8080.",
    );
  }

  let payload: unknown;

  try {
    payload = await response.json();
  } catch {
    throw new Error("Backend mengirim respons yang tidak dapat dibaca.");
  }

  if (!response.ok) {
    const errorPayload = payload as { message?: string };

    throw new Error(
      errorPayload.message ??
        `Gagal mengambil riwayat pertandingan (HTTP ${response.status}).`,
    );
  }

  return payload as MatchListResponse;
}
