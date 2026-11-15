export interface PredictionRequest {
  home_team: string;
  away_team: string;
  kickoff_date: string;
}

export type Outcome = "H" | "D" | "A";

export interface ModelPrediction {
  prediction: Outcome;
  probabilities: Record<Outcome, number>;
}

export interface PredictionResponse {
  featureCount?: number;
  logisticRegression: ModelPrediction;
  randomForest: ModelPrediction;
}

type JsonObject = Record<string, unknown>;

function isObject(value: unknown): value is JsonObject {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function firstObject(
  source: JsonObject,
  keys: string[],
): JsonObject | undefined {
  for (const key of keys) {
    if (isObject(source[key])) return source[key];
  }
  return undefined;
}

function firstString(source: JsonObject, keys: string[]): string | undefined {
  for (const key of keys) {
    const value = source[key];
    if (typeof value === "string" && value.trim()) return value;
  }
  return undefined;
}

function readNumber(source: JsonObject, keys: string[]): number | undefined {
  for (const key of keys) {
    const value = source[key];

    if (typeof value === "number" && Number.isFinite(value)) return value;

    if (typeof value === "string" && value.trim()) {
      const parsed = Number(value);
      if (Number.isFinite(parsed)) return parsed;
    }
  }

  return undefined;
}

function mapOutcome(value: string | undefined): Outcome | undefined {
  const normalized = value?.trim().toLowerCase();

  if (["h", "home", "home_win", "home win"].includes(normalized ?? "")) {
    return "H";
  }

  if (["d", "draw"].includes(normalized ?? "")) {
    return "D";
  }

  if (["a", "away", "away_win", "away win"].includes(normalized ?? "")) {
    return "A";
  }

  return undefined;
}

function normalizeModel(
  rawModel: unknown,
  root: JsonObject,
  prefix: string,
): ModelPrediction {
  const model = isObject(rawModel) ? rawModel : {};

  const probabilitySource =
    firstObject(model, ["probabilities", "probability", "probs"]) ??
    firstObject(root, [`${prefix}_probabilities`, `${prefix}Probabilities`]) ??
    model;

  const home = readNumber(probabilitySource, [
    "H",
    "h",
    "home",
    "home_win",
    "homeWin",
  ]);

  const draw = readNumber(probabilitySource, ["D", "d", "draw"]);

  const away = readNumber(probabilitySource, [
    "A",
    "a",
    "away",
    "away_win",
    "awayWin",
  ]);

  if (home === undefined || draw === undefined || away === undefined) {
    throw new Error(
      `Format probabilitas ${prefix} dari backend belum sesuai dengan format yang diharapkan.`,
    );
  }

  const scale = Math.max(home, draw, away) <= 1.001 ? 100 : 1;

  const probabilities: Record<Outcome, number> = {
    H: home * scale,
    D: draw * scale,
    A: away * scale,
  };

  const rawPrediction =
    firstString(model, [
      "prediction",
      "predicted_result",
      "predictedResult",
      "predicted_class",
      "predictedClass",
    ]) ?? firstString(root, [`${prefix}_prediction`, `${prefix}Prediction`]);

  const prediction =
    mapOutcome(rawPrediction) ??
    (Object.entries(probabilities).sort(
      (a, b) => b[1] - a[1],
    )[0][0] as Outcome);

  return { prediction, probabilities };
}

function normalizeResponse(payload: unknown): PredictionResponse {
  if (!isObject(payload)) {
    throw new Error("Respons backend tidak memiliki format JSON yang valid.");
  }

  const root =
    (isObject(payload.data) ? payload.data : undefined) ??
    (isObject(payload.result) ? payload.result : undefined) ??
    payload;

  const models =
    firstObject(root, [
      "probabilities",
      "models",
      "predictions",
      "model_results",
    ]) ?? root;

  const lrRaw =
    models.logistic_regression ?? models.logisticRegression ?? models.lr;

  const rfRaw = models.random_forest ?? models.randomForest ?? models.rf;

  const featureCountValue = root.feature_count ?? root.featureCount;
  const featureCount = Number(featureCountValue);

  return {
    featureCount: Number.isFinite(featureCount) ? featureCount : undefined,
    logisticRegression: normalizeModel(lrRaw, root, "logistic_regression"),
    randomForest: normalizeModel(rfRaw, root, "random_forest"),
  };
}

export async function createPrediction(
  input: PredictionRequest,
): Promise<PredictionResponse> {
  let response: Response;

  try {
    response = await fetch("/api/v1/predictions", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(input),
    });
  } catch {
    throw new Error(
      "Tidak dapat terhubung ke backend. Pastikan API Go sedang berjalan di port 8080.",
    );
  }

  let payload: unknown;

  try {
    payload = await response.json();
  } catch {
    throw new Error("Backend mengirim respons yang tidak dapat dibaca.");
  }

  if (!response.ok) {
    const errorObject = isObject(payload) ? payload : {};

    const message =
      firstString(errorObject, ["error", "message", "detail"]) ??
      `Permintaan prediksi gagal dengan HTTP ${response.status}.`;

    throw new Error(message);
  }

  return normalizeResponse(payload);
}
