import AgentsPage from "@/pages/agents/agents";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/agents")({
  component: RouteComponent,
});

function RouteComponent() {
  return <AgentsPage />;
}
