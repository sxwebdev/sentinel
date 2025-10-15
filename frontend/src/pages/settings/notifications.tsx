// import { connectQueryClient } from "@/api/api";
import {
  providerCreate,
  providerDelete,
  providersList,
  providerTest,
  providerUpdate,
} from "@/api/gen/sentinel/notifications/v1/notifications-NotificationService_connectquery";
import {
  ProviderType,
  type Provider,
  type ProviderConfig,
  ProviderConfigSchema,
  ShoutrrrConfigSchema,
  ProviderCreateRequestSchema,
  ProviderUpdateRequestSchema,
} from "@/api/gen/sentinel/notifications/v1/notifications_pb";
import { create } from "@bufbuild/protobuf";
import { ConfirmDialog } from "@/entities/confirmDialog/confirmDialog";
import { H4, P } from "@/shared/components/typography";
import {
  Badge,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogClose,
  DialogFooter,
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
  Input,
  Switch,
} from "@/shared/components/ui";
import { Button } from "@/shared/components/ui/button";
import { Separator } from "@/shared/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/components/ui/table";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import { EllipsisIcon, PlusIcon } from "lucide-react";
import { useForm } from "@tanstack/react-form";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import * as z from "zod";

const configByProviderType = (type: ProviderType, config?: ProviderConfig) => {
  switch (type) {
    case ProviderType.SHOUTRRR: {
      // cut middle symbols for long urls
      if (!config?.config?.value?.url) return null;
      const url = config.config.value.url;
      if (url.length > 30) {
        return url.slice(0, 15) + "..." + url.slice(-15);
      }
      return url;
    }
    default:
      return null;
  }
};

const providerTypeName = (type: ProviderType): string => {
  switch (type) {
    case ProviderType.SHOUTRRR:
      return "shoutrrr";
    default:
      return "unknown";
  }
};

const formSchema = z.object({
  url: z.string().min(1, "URL is required"),
  isEnabled: z.boolean(),
});

const formInitialValues = (provider?: Provider) => {
  return {
    url: provider?.config?.config?.value?.url || "",
    isEnabled: provider?.isEnabled ?? true,
  };
};

type UpsertProviderProps = {
  open: boolean;
  provider?: Provider;
  onClose: () => void;
  onSubmit: (data: z.infer<typeof formSchema>) => Promise<void>;
};

const UpsertProvider = ({
  open,
  provider,
  onClose,
  onSubmit,
}: UpsertProviderProps) => {
  const form = useForm({
    defaultValues: formInitialValues(provider),
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: ({ formApi, value }) => {
      onSubmit(value).then(() => formApi.reset());
    },
  });

  useEffect(() => {
    if (open) {
      form.reset(formInitialValues(provider));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, provider?.id]);

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {provider ? "Edit Provider" : "Add Provider"}
          </DialogTitle>
          <DialogDescription>
            {provider
              ? "Update the provider details and click on the Update button."
              : "Fill out the form below to create a new provider."}
          </DialogDescription>
        </DialogHeader>
        <form
          id="upsert-provider-form"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
          className="space-y-8"
        >
          <FieldGroup>
            <form.Field
              name="url"
              children={(field) => {
                const isInvalid =
                  field.state.meta.isTouched && !field.state.meta.isValid;
                return (
                  <Field data-invalid={isInvalid}>
                    <FieldLabel htmlFor={field.name}>URL</FieldLabel>
                    <Input
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      aria-invalid={isInvalid}
                      placeholder="Provider URL"
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
              name="isEnabled"
              children={(field) => (
                <Field>
                  <div className="flex flex-row items-center justify-between">
                    <FieldLabel htmlFor={field.name}>Is enabled</FieldLabel>
                    <Switch
                      id={field.name}
                      name={field.name}
                      checked={field.state.value}
                      onCheckedChange={(checked) =>
                        field.handleChange(!!checked)
                      }
                    />
                  </div>
                </Field>
              )}
            />
          </FieldGroup>
        </form>
        <DialogFooter className="mt-2">
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button disabled={!form.state.isValid} form="upsert-provider-form">
            {provider ? "Update" : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

const NotificationsPage = () => {
  const [deleteProviderId, setDeleteProviderId] = useState<
    string | undefined
  >();
  const [upsertProviderOpen, setUpsertProviderOpen] = useState(false);
  const [provider, setProvider] = useState<Provider | undefined>();

  const q = useQuery(providersList);

  const createMutation = useMutation(providerCreate, {
    onSuccess: () => {
      toast.success("Provider created successfully");
      void q.refetch();
    },
    onError: (error) => {
      toast.error((error as Error).message || "Failed to create provider");
    },
  });

  const updateMutation = useMutation(providerUpdate, {
    onSuccess: () => {
      toast.success("Provider updated successfully");
      void q.refetch();
    },
    onError: (error) => {
      toast.error((error as Error).message || "Failed to update provider");
    },
  });

  const deleteMutation = useMutation(providerDelete, {
    onSuccess: () => {
      toast.success("Provider deleted successfully");
      void q.refetch();
      setDeleteProviderId(undefined);
    },
    onError: (error) => {
      toast.error((error as Error).message || "Failed to delete provider");
    },
  });

  const testMutation = useMutation(providerTest, {
    onSuccess: () => {
      toast.success("Test notification sent successfully");
    },
    onError: (error) => {
      toast.error(
        (error as Error).message || "Failed to send test notification",
      );
    },
  });

  const onCreateProvider = useCallback(
    async (data: z.infer<typeof formSchema>) => {
      const createRequest = create(ProviderCreateRequestSchema, {
        type: ProviderType.SHOUTRRR,
        config: create(ProviderConfigSchema, {
          config: {
            case: "shoutrrr",
            value: create(ShoutrrrConfigSchema, { url: data.url }),
          },
        }),
        isEnabled: data.isEnabled,
      });
      await createMutation.mutateAsync(createRequest);
    },
    [createMutation],
  );

  const onUpdateProvider = useCallback(
    async (data: z.infer<typeof formSchema>) => {
      if (!provider) return;

      // Build full update request, preserving type and updating config oneof
      await updateMutation.mutateAsync(
        create(ProviderUpdateRequestSchema, {
          id: provider.id,
          type: provider.type,
          config: create(ProviderConfigSchema, {
            config: {
              case: "shoutrrr",
              value: create(ShoutrrrConfigSchema, { url: data.url }),
            },
          }),
          isEnabled: data.isEnabled,
        }),
      );
    },
    [provider, updateMutation],
  );

  const onDeleteProvider = useCallback(async () => {
    if (!deleteProviderId) return;
    await deleteMutation.mutateAsync({ id: deleteProviderId });
  }, [deleteMutation, deleteProviderId]);

  const onToggleIsEnabledProvider = useCallback(
    async (p: Provider) => {
      await updateMutation.mutateAsync(
        create(ProviderUpdateRequestSchema, {
          id: p.id,
          type: p.type,
          config: p.config,
          isEnabled: !p.isEnabled,
        }),
      );
    },
    [updateMutation],
  );

  const onTestProvider = useCallback(
    async (provider: Provider) => {
      await testMutation.mutateAsync({ id: provider.id });
    },
    [testMutation],
  );

  return (
    <>
      <ConfirmDialog
        open={!!deleteProviderId}
        setOpen={() => setDeleteProviderId(undefined)}
        onSubmit={onDeleteProvider}
        title="Delete a provider"
        description="Are you sure you want to delete this provider?"
        type="delete"
      />
      <UpsertProvider
        open={upsertProviderOpen}
        provider={provider}
        onClose={() => {
          setUpsertProviderOpen(false);
          setProvider(undefined);
        }}
        onSubmit={async (data) => {
          if (provider) {
            onUpdateProvider(data);
          } else {
            onCreateProvider(data);
          }
          setUpsertProviderOpen(false);
          setProvider(undefined);
        }}
      />
      <H4 className="flex justify-between">
        <span>Notifications</span>
        <Button
          size="sm"
          variant="secondary"
          onClick={() => {
            setProvider(undefined);
            setUpsertProviderOpen(true);
          }}
        >
          <PlusIcon size={16} /> Add Provider
        </Button>
      </H4>
      <P className="text-muted-foreground text-sm leading-relaxed">
        Manage your notification providers here.
        <br />
        You can add, edit, or remove providers to receive alerts and updates.
      </P>
      <Separator className="my-5" />
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50">
              <TableHead>Provider type</TableHead>
              <TableHead>Is enabled</TableHead>
              <TableHead>Config</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {q.data?.items?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="h-24 text-center">
                  No providers found.
                </TableCell>
              </TableRow>
            ) : (
              q.data?.items?.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>{providerTypeName(item.type)}</TableCell>
                  <TableCell>
                    <Badge variant={item.isEnabled ? "success" : "warning"}>
                      {item.isEnabled ? "Enabled" : "Disabled"}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {configByProviderType(item.type, item.config)}
                  </TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <div className="flex justify-end">
                          <Button
                            size="icon"
                            variant="ghost"
                            className="shadow-none"
                            aria-label="Actions"
                          >
                            <EllipsisIcon size={16} aria-hidden="true" />
                          </Button>
                        </div>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => onTestProvider(item)}>
                          Test
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuGroup>
                          <DropdownMenuItem
                            onClick={() => {
                              setProvider(item);
                              setUpsertProviderOpen(true);
                            }}
                          >
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => onToggleIsEnabledProvider(item)}
                          >
                            {item.isEnabled ? "Disable" : "Enable"}
                          </DropdownMenuItem>
                        </DropdownMenuGroup>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => {
                            setDeleteProviderId(item?.id);
                          }}
                        >
                          <span>Delete</span>
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </>
  );
};

export default NotificationsPage;
