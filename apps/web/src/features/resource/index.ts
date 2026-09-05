export { ResourceList } from "./components/resource-list";
export { ResourceCard } from "./components/resource-card";
export { ResourceDetail } from "./components/resource-detail";
export { ResourceFilters } from "./components/resource-filters";
export { ResourceForm } from "./components/resource-form";

export { useResources } from "./hooks/use-resources";
export { useResource } from "./hooks/use-resource";
export { useCreateResource } from "./hooks/use-create-resource";
export { useUpdateResource } from "./hooks/use-update-resource";
export { useDeleteResource } from "./hooks/use-delete-resource";

export { useResourceStore } from "./store";
export { resourceSchema, type ResourceInput } from "./schemas/resource";

export type {
  Resource,
  ResourceStatus,
  CreateResourceInput,
  UpdateResourceInput,
} from "./types";
