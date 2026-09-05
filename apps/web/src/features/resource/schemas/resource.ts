import { z } from "zod";

export const resourceSchema = z.object({
  title: z
    .string()
    .min(1, "Title is required")
    .max(120, "Title must be 120 characters or fewer"),
  description: z
    .string()
    .max(2000, "Description must be 2000 characters or fewer")
    .optional(),
  status: z.enum(["draft", "active", "archived"]),
});

export type ResourceInput = z.infer<typeof resourceSchema>;
