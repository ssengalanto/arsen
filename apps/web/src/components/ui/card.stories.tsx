import type { Meta, StoryObj } from "@storybook/nextjs";

import { Button } from "./button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "./card";

const meta = {
  title: "UI/Card",
  component: Card,
  argTypes: {
    size: { control: "select", options: ["default", "sm"] },
  },
  args: { size: "default" },
} satisfies Meta<typeof Card>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => (
    <Card {...args} className="w-80">
      <CardHeader>
        <CardTitle>Project settings</CardTitle>
        <CardDescription>Manage how this project behaves.</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground">
          Cards group related content and actions in a single surface.
        </p>
      </CardContent>
      <CardFooter>
        <Button size="sm">Save</Button>
      </CardFooter>
    </Card>
  ),
};

export const WithAction: Story = {
  render: (args) => (
    <Card {...args} className="w-80">
      <CardHeader>
        <CardTitle>Notifications</CardTitle>
        <CardDescription>You have 3 unread messages.</CardDescription>
        <CardAction>
          <Button variant="ghost" size="sm">
            Clear
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground">Content sits below the header.</p>
      </CardContent>
    </Card>
  ),
};

export const Small: Story = {
  args: { size: "sm" },
  render: (args) => (
    <Card {...args} className="w-72">
      <CardHeader>
        <CardTitle>Compact card</CardTitle>
        <CardDescription>Uses reduced spacing.</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground">A denser layout variant.</p>
      </CardContent>
    </Card>
  ),
};
