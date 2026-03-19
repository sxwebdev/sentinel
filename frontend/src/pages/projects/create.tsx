import { Button } from "@/shared/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/components/ui/card";
import { Field, FieldGroup, FieldLabel } from "@/shared/components/ui/field";
import { Input } from "@/shared/components/ui/input";
import { toast } from "sonner";
import { useState } from "react";
import { Spinner } from "@/shared/components/ui/spinner";
import { useProjectStore } from "@/app/stores/projects";
import Logo from "@/shared/components/logo";

export function ProjectCreate({ ...props }: React.ComponentProps<"div">) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const projectsStore = useProjectStore();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!name) {
      toast.error("Please fill in all required fields");
      return;
    }

    try {
      setIsLoading(true);
      await projectsStore.createProject(name, description);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      className="flex min-h-svh w-full items-center justify-center p-6 md:p-10"
      {...props}
    >
      <div className="w-full max-w-sm">
        <div className="flex flex-col gap-6">
          <Logo className="h-7 w-auto" />
          <Card>
            <CardHeader>
              <CardTitle>Create your first project</CardTitle>
              <CardDescription>
                Enter your project details below to get started
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit}>
                <FieldGroup>
                  <Field>
                    <FieldLabel htmlFor="name" required>
                      Name
                    </FieldLabel>
                    <Input
                      id="name"
                      type="text"
                      placeholder="Name"
                      required
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                    />
                  </Field>
                  <Field>
                    <FieldLabel htmlFor="description">Description</FieldLabel>
                    <Input
                      id="description"
                      type="text"
                      placeholder="Description"
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                    />
                  </Field>
                  <Field>
                    <Button type="submit">
                      {isLoading ? <Spinner /> : "Create Project"}
                    </Button>
                  </Field>
                </FieldGroup>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

export default ProjectCreate;
