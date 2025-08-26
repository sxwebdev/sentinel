import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/incidents")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/incidents"!</div>;
}
