export type ResourceStatus = "draft" | "active" | "archived";

export interface Resource {
  id: string;
  title: string;
  description: string;
  status: ResourceStatus;
  createdAt: string;
  updatedAt: string;
}

export interface CreateResourceInput {
  title: string;
  description?: string;
  status: ResourceStatus;
}

export type UpdateResourceInput = Partial<CreateResourceInput>;
