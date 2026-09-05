import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/test/msw/server";
import { problem } from "@/test/msw/fixtures";
import { ApiError } from "@/lib/api/error";
import {
  listResources,
  getResource,
  createResource,
  updateResource,
  deleteResource,
} from "./index";
import type { Resource } from "../types";

const row: Resource = {
  id: "res-1",
  title: "Row",
  description: "",
  status: "draft",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
};

beforeEach(() => {
  server.resetHandlers();
});

describe("resource fetchers", () => {
  it("listResources GETs the collection", async () => {
    server.use(http.get("/api/resources", () => HttpResponse.json([row])));
    await expect(listResources()).resolves.toEqual([row]);
  });

  it("getResource GETs a single row by encoded id", async () => {
    server.use(
      http.get("/api/resources/res%201", () => HttpResponse.json({ ...row, id: "res 1" })),
    );
    await expect(getResource("res 1")).resolves.toMatchObject({ id: "res 1" });
  });

  it("createResource POSTs the input and returns the created row", async () => {
    server.use(
      http.post("/api/resources", async ({ request }) => {
        const body = (await request.json()) as { title: string };
        return HttpResponse.json({ ...row, title: body.title }, { status: 201 });
      }),
    );
    await expect(createResource({ title: "New", status: "draft" })).resolves.toMatchObject({
      title: "New",
    });
  });

  it("updateResource PATCHes the input", async () => {
    server.use(
      http.patch("/api/resources/res-1", async ({ request }) => {
        const body = (await request.json()) as { status: string };
        return HttpResponse.json({ ...row, status: body.status });
      }),
    );
    await expect(updateResource("res-1", { status: "archived" })).resolves.toMatchObject({
      status: "archived",
    });
  });

  it("deleteResource DELETEs and resolves void on 204", async () => {
    server.use(
      http.delete("/api/resources/res-1", () => new HttpResponse(null, { status: 204 })),
    );
    await expect(deleteResource("res-1")).resolves.toBeUndefined();
  });

  it("throws an ApiError on a problem response", async () => {
    server.use(
      http.get("/api/resources/missing", () =>
        HttpResponse.json(problem(404, "not found"), { status: 404 }),
      ),
    );
    await expect(getResource("missing")).rejects.toBeInstanceOf(ApiError);
  });
});
