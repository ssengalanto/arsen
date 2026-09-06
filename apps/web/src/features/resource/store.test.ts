import { describe, it, expect, beforeEach } from "vitest";
import { useResourceStore } from "./store";

const STORAGE_KEY = "arsen-resource";

function persisted() {
  const raw = localStorage.getItem(STORAGE_KEY);
  return raw ? (JSON.parse(raw) as { state: Record<string, unknown> }).state : {};
}

beforeEach(() => {
  localStorage.clear();
  useResourceStore.setState({
    selectedIds: [],
    filters: { status: "all" },
    draft: null,
    wizardStep: 0,
  });
});

describe("resourceStore", () => {
  it("persists the draft across a close/reopen", () => {
    useResourceStore.getState().setDraft({ title: "Work in progress" });
    expect(persisted().draft).toEqual({ title: "Work in progress" });
  });

  it("clears the draft on successful submit", () => {
    useResourceStore.getState().setDraft({ title: "Work in progress" });
    useResourceStore.getState().clearDraft();
    expect(useResourceStore.getState().draft).toBeNull();
    expect(persisted().draft ?? null).toBeNull();
  });

  it("persists filters", () => {
    useResourceStore.getState().setStatusFilter("active");
    expect(persisted().filters).toEqual({ status: "active" });
  });

  it("does NOT persist selectedIds or wizardStep", () => {
    useResourceStore.getState().toggleSelected("abc");
    useResourceStore.getState().setWizardStep(3);
    expect(persisted().selectedIds).toBeUndefined();
    expect(persisted().wizardStep).toBeUndefined();
  });

  it("toggles a selected id on and off, and clears the whole selection", () => {
    const store = useResourceStore.getState();
    store.toggleSelected("a");
    store.toggleSelected("b");
    expect(useResourceStore.getState().selectedIds).toEqual(["a", "b"]);

    store.toggleSelected("a"); // already selected → removed
    expect(useResourceStore.getState().selectedIds).toEqual(["b"]);

    store.clearSelection();
    expect(useResourceStore.getState().selectedIds).toEqual([]);
  });
});
