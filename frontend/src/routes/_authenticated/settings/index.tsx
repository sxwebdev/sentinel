import { H4 } from "@/shared/components/typography";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/settings/")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div>
      <H4>General</H4>
    </div>
  );
}
