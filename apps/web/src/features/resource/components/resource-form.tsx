"use client";

import { useEffect, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { ApiError } from "@/lib/api/error";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { resourceSchema, type ResourceInput } from "../schemas/resource";
import { useResourceStore } from "../store";
import { useCreateResource } from "../hooks/use-create-resource";

const STATUS_OPTIONS: { value: ResourceInput["status"]; label: string }[] = [
  { value: "draft", label: "Draft" },
  { value: "active", label: "Active" },
  { value: "archived", label: "Archived" },
];

const selectClassName =
  "h-8 w-full rounded-lg border border-input bg-transparent px-2.5 py-1 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-50";

// Create form for the worked slice. The in-progress draft is mirrored into the
// persisted Zustand store on every keystroke so a reload restores it; on a
// successful create we reset the fields, clear that draft, then dismiss.
export function ResourceForm({ onCreated }: { onCreated?: () => void }) {
  const draft = useResourceStore((s) => s.draft);
  const setDraft = useResourceStore((s) => s.setDraft);
  const clearDraft = useResourceStore((s) => s.clearDraft);
  const { create } = useCreateResource();

  const form = useForm<ResourceInput>({
    resolver: zodResolver(resourceSchema),
    defaultValues: {
      title: draft?.title ?? "",
      description: draft?.description ?? "",
      status: draft?.status ?? "draft",
    },
  });
  const { register, handleSubmit, control, reset, setError, formState } = form;
  const { errors, isSubmitting } = formState;

  // Mirror the in-progress values into the persisted draft on every change so a
  // reload (or close/reopen) restores them. `useWatch` is React-Compiler-safe,
  // unlike the callback form of `watch`. After a successful submit we stop
  // mirroring so the empty reset can't repopulate the draft we just cleared.
  const [submitted, setSubmitted] = useState(false);
  const watched = useWatch({ control });
  useEffect(() => {
    if (submitted) return;
    setDraft(watched);
  }, [watched, setDraft, submitted]);

  const onSubmit = handleSubmit(async (values) => {
    try {
      await create({
        title: values.title,
        description: values.description,
        status: values.status,
      });
      setSubmitted(true);
      reset({ title: "", description: "", status: "draft" });
      clearDraft();
      onCreated?.();
      toast.success("Resource created");
    } catch (err) {
      if (err instanceof ApiError) {
        const fieldErrors = err.fieldErrors();
        for (const [field, message] of Object.entries(fieldErrors)) {
          setError(field as keyof ResourceInput, { message });
        }
        if (Object.keys(fieldErrors).length === 0) {
          setError("root", { message: err.detail ?? "Something went wrong." });
        }
      } else {
        setError("root", { message: "Something went wrong." });
      }
    }
  });

  return (
    <form onSubmit={(e) => void onSubmit(e)} className="flex flex-col gap-4" noValidate>
      {errors.root && (
        <p className="text-destructive text-sm" role="alert">
          {errors.root.message}
        </p>
      )}

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="title">Title</Label>
        <Input id="title" autoComplete="off" {...register("title")} />
        {errors.title && <p className="text-destructive text-xs">{errors.title.message}</p>}
      </div>

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="description">Description</Label>
        <textarea
          id="description"
          rows={4}
          className={selectClassName + " h-auto resize-y"}
          {...register("description")}
        />
        {errors.description && (
          <p className="text-destructive text-xs">{errors.description.message}</p>
        )}
      </div>

      <div className="flex flex-col gap-1.5">
        <Label htmlFor="status">Status</Label>
        <select id="status" className={selectClassName} {...register("status")}>
          {STATUS_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        {errors.status && <p className="text-destructive text-xs">{errors.status.message}</p>}
      </div>

      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? "Saving…" : "Create resource"}
      </Button>
    </form>
  );
}
