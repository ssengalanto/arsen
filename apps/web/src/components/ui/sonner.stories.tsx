import type { Meta, StoryObj } from "@storybook/nextjs";
import { toast } from "sonner";

import { Button } from "./button";
import { Toaster } from "./sonner";

const meta = {
  title: "UI/Toaster",
  component: Toaster,
  parameters: {
    docs: {
      description: {
        component:
          "App-wide toast host. Mount `<Toaster />` once at the root, then call `toast()` from anywhere.",
      },
    },
  },
} satisfies Meta<typeof Toaster>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => (
    <div className="flex flex-wrap gap-2">
      <Toaster {...args} />
      <Button onClick={() => toast("Event has been created.")}>Default</Button>
      <Button
        variant="outline"
        onClick={() => toast.success("Changes saved.")}
      >
        Success
      </Button>
      <Button
        variant="outline"
        onClick={() => toast.error("Something went wrong.")}
      >
        Error
      </Button>
      <Button
        variant="outline"
        onClick={() => toast.warning("Your session expires soon.")}
      >
        Warning
      </Button>
      <Button
        variant="outline"
        onClick={() => toast.info("A new update is available.")}
      >
        Info
      </Button>
    </div>
  ),
};
