import type { Meta, StoryObj } from "@storybook/nextjs";

import { Input } from "./input";

const meta = {
  title: "UI/Input",
  component: Input,
  argTypes: {
    type: {
      control: "select",
      options: ["text", "email", "password", "number", "search"],
    },
    disabled: { control: "boolean" },
  },
  args: {
    type: "text",
    placeholder: "Type here...",
  },
} satisfies Meta<typeof Input>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Email: Story = {
  args: { type: "email", placeholder: "you@example.com" },
};

export const Password: Story = {
  args: { type: "password", placeholder: "••••••••" },
};

export const Disabled: Story = {
  args: { disabled: true, placeholder: "Disabled" },
};

export const Invalid: Story = {
  args: { "aria-invalid": true, defaultValue: "not-an-email" },
};
