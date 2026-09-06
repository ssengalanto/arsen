export interface ProblemFieldError {
  field: string;
  detail: string;
}

export interface ProblemDetail {
  type?: string;
  title?: string;
  status?: number;
  detail?: string;
  instance?: string;
  errors?: ProblemFieldError[];
}

// Wraps an RFC 9457 `application/problem+json` error. Only the known problem
// fields are retained — any extra fields (e.g. leaked tokens) are dropped.
export class ApiError extends Error {
  readonly status: number;
  readonly type: string | undefined;
  readonly title: string | undefined;
  readonly detail: string | undefined;
  readonly instance: string | undefined;
  readonly errors: ProblemFieldError[];

  constructor(status: number, problem: ProblemDetail = {}) {
    super(problem.detail ?? problem.title ?? `Request failed with status ${status}`);
    this.name = "ApiError";
    this.status = status;
    this.type = problem.type;
    this.title = problem.title;
    this.detail = problem.detail;
    this.instance = problem.instance;
    this.errors = problem.errors ?? [];
  }

  // Field → message map for react-hook-form `setError` binding.
  fieldErrors(): Record<string, string> {
    const map: Record<string, string> = {};
    for (const { field, detail } of this.errors) {
      if (field && !(field in map)) map[field] = detail;
    }
    return map;
  }

  static async fromResponse(res: Response): Promise<ApiError> {
    let problem: ProblemDetail = {};
    try {
      const contentType = res.headers.get("content-type") ?? "";
      if (contentType.includes("json")) {
        problem = sanitize(await res.clone().json());
      }
    } catch {
      // Non-JSON or unparseable body → generic error by status only.
    }
    return new ApiError(res.status, problem);
  }
}

function sanitize(body: unknown): ProblemDetail {
  if (typeof body !== "object" || body === null) return {};
  const b = body as Record<string, unknown>;
  const errors = Array.isArray(b.errors)
    ? b.errors
        .filter((e): e is Record<string, unknown> => typeof e === "object" && e !== null)
        .map((e) => ({
          field: typeof e.field === "string" ? e.field : "",
          detail: typeof e.detail === "string" ? e.detail : "",
        }))
    : undefined;
  return {
    type: typeof b.type === "string" ? b.type : undefined,
    title: typeof b.title === "string" ? b.title : undefined,
    status: typeof b.status === "number" ? b.status : undefined,
    detail: typeof b.detail === "string" ? b.detail : undefined,
    instance: typeof b.instance === "string" ? b.instance : undefined,
    errors,
  };
}
