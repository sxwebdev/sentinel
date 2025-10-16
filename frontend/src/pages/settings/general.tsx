import {
  projectDelete,
  projectGet,
  projectUpdate,
} from "@/api/gen/sentinel/projects/v1/projects-ProjectsService_connectquery";
import type { Project } from "@/api/gen/sentinel/projects/v1/projects_pb";
import { useProjectStore } from "@/app/stores/projects";
import { H4, P } from "@/shared/components/typography";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Separator,
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
  Input,
  Button,
  Textarea,
  AlertDialog,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogTrigger,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogCancel,
} from "@/shared/components/ui";
import {
  Item,
  ItemActions,
  ItemContent,
  ItemDescription,
  ItemTitle,
} from "@/shared/components/ui/item";
import { Spinner } from "@/shared/components/ui/spinner";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import { useForm } from "@tanstack/react-form";
import { SaveIcon, TrashIcon } from "lucide-react";
import { toast } from "sonner";
import * as z from "zod";
import { useShallow } from "zustand/react/shallow";

const formSchema = z.object({
  name: z.string().min(1, "Project name is required"),
  description: z.union([z.string(), z.undefined()]),
  settings: z.object({
    monitorDefaults: z.object({
      interval: z.bigint().min(BigInt(1), "Interval must be positive"),
      timeout: z.bigint().min(BigInt(1), "Timeout must be positive"),
      retries: z.bigint().min(BigInt(0), "Retries must be zero or positive"),
    }),
  }),
});

const formInitialValues = (project?: Project): z.infer<typeof formSchema> => ({
  name: project?.name ?? "",
  description: project?.description ?? "",
  settings: {
    monitorDefaults: {
      interval: project?.settings?.monitorDefaults?.interval ?? BigInt(60000),
      timeout: project?.settings?.monitorDefaults?.timeout ?? BigInt(10000),
      retries: project?.settings?.monitorDefaults?.retries ?? BigInt(10),
    },
  },
});

const GeneralSettingsPage = () => {
  const { selectedProjectId, selectProject, loadProjects } = useProjectStore(
    useShallow((s) => ({
      selectedProjectId: s.selectedProjectId,
      loadProjects: s.loadProjects,
      selectProject: s.selectProject,
    })),
  );

  const q = useQuery(projectGet, { id: selectedProjectId! });

  const updateMutation = useMutation(projectUpdate, {
    onSuccess: () => {
      toast.success("Project updated successfully");
      void q.refetch();
    },
    onError: (error) => {
      toast.error((error as Error).message || "Failed to update project");
    },
  });

  const deleteMutation = useMutation(projectDelete, {
    onSuccess: () => {
      selectProject(undefined);
    },
    onError: (error) => {
      toast.error((error as Error).message || "Failed to delete project");
    },
  });

  const form = useForm({
    defaultValues: formInitialValues(q.data?.item),
    validators: {
      onSubmit: formSchema,
      onChange: formSchema,
    },
    onSubmit: async ({ formApi, value }) => {
      // call the mutation to update the project
      const data = await updateMutation.mutateAsync({
        id: selectedProjectId!,
        ...value,
      });

      // update form values to the latest returned from the server
      formApi.reset(formInitialValues(data.item), { keepDefaultValues: true });

      // reload projects in case name changed
      loadProjects();
    },
  });

  return (
    <>
      <H4>General</H4>
      <P className="text-muted-foreground text-sm leading-relaxed">
        On this page you can edit general settings for your account.
      </P>
      <Separator className="my-5" />
      <Card>
        <CardHeader>
          <CardTitle>Project settings</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            id="general-form"
            onSubmit={(e) => {
              e.preventDefault();
              e.stopPropagation();
              form.handleSubmit();
            }}
            className="space-y-8"
          >
            <FieldGroup>
              <form.Field
                name="name"
                children={(field) => {
                  const isInvalid =
                    field.state.meta.isTouched && !field.state.meta.isValid;
                  return (
                    <Field data-invalid={isInvalid}>
                      <FieldLabel htmlFor={field.name}>Name</FieldLabel>
                      <Input
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                        aria-invalid={isInvalid}
                        placeholder="Name"
                        autoComplete="off"
                      />
                      {isInvalid && (
                        <FieldError errors={field.state.meta.errors} />
                      )}
                    </Field>
                  );
                }}
              />

              <form.Field
                name="description"
                children={(field) => {
                  const isInvalid =
                    field.state.meta.isTouched && !field.state.meta.isValid;
                  return (
                    <Field data-invalid={isInvalid}>
                      <FieldLabel htmlFor={field.name}>Description</FieldLabel>
                      <Textarea
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onBlur={field.handleBlur}
                        onChange={(e) => field.handleChange(e.target.value)}
                        aria-invalid={isInvalid}
                        placeholder="Description"
                        autoComplete="off"
                        rows={4}
                      />
                      {isInvalid && (
                        <FieldError errors={field.state.meta.errors} />
                      )}
                    </Field>
                  );
                }}
              />
              <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                <form.Field
                  name="settings.monitorDefaults.interval"
                  children={(field) => {
                    const isInvalid =
                      field.state.meta.isTouched && !field.state.meta.isValid;
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>
                          Default Interval
                        </FieldLabel>
                        <Input
                          id={field.name}
                          name={field.name}
                          type="number"
                          value={field.state.value.toString()}
                          onBlur={field.handleBlur}
                          onChange={(e) =>
                            field.handleChange(BigInt(e.target.value))
                          }
                          aria-invalid={isInvalid}
                          placeholder="Default Interval"
                          autoComplete="off"
                        />
                        {isInvalid && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    );
                  }}
                />
                <form.Field
                  name="settings.monitorDefaults.timeout"
                  children={(field) => {
                    const isInvalid =
                      field.state.meta.isTouched && !field.state.meta.isValid;
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>
                          Default Timeout
                        </FieldLabel>
                        <Input
                          id={field.name}
                          name={field.name}
                          type="number"
                          value={field.state.value.toString()}
                          onBlur={field.handleBlur}
                          onChange={(e) =>
                            field.handleChange(BigInt(e.target.value))
                          }
                          aria-invalid={isInvalid}
                          placeholder="Default Timeout"
                          autoComplete="off"
                        />
                        {isInvalid && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    );
                  }}
                />
                <form.Field
                  name="settings.monitorDefaults.retries"
                  children={(field) => {
                    const isInvalid =
                      field.state.meta.isTouched && !field.state.meta.isValid;
                    return (
                      <Field data-invalid={isInvalid}>
                        <FieldLabel htmlFor={field.name}>Retries</FieldLabel>
                        <Input
                          id={field.name}
                          name={field.name}
                          type="number"
                          value={field.state.value.toString()}
                          onBlur={field.handleBlur}
                          onChange={(e) =>
                            field.handleChange(BigInt(e.target.value))
                          }
                          aria-invalid={isInvalid}
                          placeholder="Retries"
                          autoComplete="off"
                        />
                        {isInvalid && (
                          <FieldError errors={field.state.meta.errors} />
                        )}
                      </Field>
                    );
                  }}
                />
              </div>
            </FieldGroup>
            <form.Subscribe
              selector={(state) => [
                state.isValid,
                state.isSubmitting,
                state.isDefaultValue,
              ]}
              children={([isValid, isSubmitting, isDefaultValue]) => {
                return (
                  <Button
                    disabled={!isValid || isDefaultValue}
                    form="general-form"
                  >
                    {isSubmitting ? <Spinner /> : <SaveIcon />}
                    Save settings
                  </Button>
                );
              }}
            />
          </form>
        </CardContent>
      </Card>
      <H4 className="mt-10">Danger zone</H4>
      <P className="text-muted-foreground text-sm leading-relaxed">
        Dangerous actions that can result in data loss or other serious
        consequences.
      </P>
      <Item
        variant="outline"
        className="border-destructive bg-destructive/5 dark:bg-destructive/20 mt-4"
      >
        <ItemContent>
          <ItemTitle>Delete this project</ItemTitle>
          <ItemDescription>
            Once you delete a project, there is no going back. Please be
            certain.
          </ItemDescription>
        </ItemContent>
        <ItemActions>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive">Delete this project</Button>
            </AlertDialogTrigger>

            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                <AlertDialogDescription>
                  Are you sure you want to delete{" "}
                  <span className="font-extrabold text-zinc-950">
                    {q.data?.item?.name}
                  </span>{" "}
                  project?
                  <br />
                  This action cannot be undone.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <Button
                  variant="destructive"
                  onClick={() => {
                    deleteMutation.mutate({ id: selectedProjectId });
                  }}
                >
                  <TrashIcon />
                  Yes, delete this project
                </Button>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </ItemActions>
      </Item>
    </>
  );
};

export default GeneralSettingsPage;
