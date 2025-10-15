import { agentsClient, VITE_SERVER_API_BASE_URL } from "@/api/api";
import {
  agentsCreate,
  agentsDelete,
  agentsList,
  agentsUpdate,
} from "@/api/gen/sentinel/agents/v1/agents-AgentsService_connectquery";
import {
  AgentsSubscribeRequestSchema,
  AgentStatus,
  type Agent,
} from "@/api/gen/sentinel/agents/v1/agents_pb";
import { ConfirmDialog } from "@/entities/confirmDialog/confirmDialog";
import { P } from "@/shared/components/typography";
import {
  Badge,
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
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
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardAction,
} from "@/shared/components/ui";
import { Separator } from "@/shared/components/ui/separator";
import { timestampDate } from "@bufbuild/protobuf/wkt";
import { useQuery, useMutation } from "@connectrpc/connect-query";
import { EllipsisIcon, PlusIcon } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { useForm } from "@tanstack/react-form";
import * as z from "zod";
import InputTag from "@/shared/components/ui/inputTag";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/shared/components/ui/alert";
import { CodeBlock } from "@/shared/components/ui/code";
import { useSubscriptionRefetch } from "@/shared/hooks/useSubscriptionRefetch";
import { create } from "@bufbuild/protobuf";

const statusTypeName = (type: AgentStatus): string => {
  switch (type) {
    case AgentStatus.ACTIVE:
      return "active";
    case AgentStatus.INACTIVE:
      return "inactive";
    case AgentStatus.UNSPECIFIED:
      return "unspecified";
    default:
      return "unknown";
  }
};

const environmentsCodeBlock = (token = "<agent_token>") => {
  return `HUB_SERVER_ADDR=${VITE_SERVER_API_BASE_URL}
TOKEN=${token}`;
};

const yamlCodeBlock = (token = "<agent_token>") => {
  return `hub_server:
  addr: ${VITE_SERVER_API_BASE_URL}
token: ${token}`;
};

const formSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters long"),
  description: z.union([z.string(), z.undefined()]),
  isEnabled: z.boolean(),
  tags: z.array(z.string()),
});

const formInitialValues = (agent?: Agent) => {
  return {
    name: agent?.name || "",
    description: agent?.description,
    isEnabled: agent?.isEnabled ?? true,
    tags: agent?.tags || [],
  };
};

type UpsertAgentProps = {
  open: boolean;
  agent?: Agent;
  onClose: () => void;
  onSubmit: (data: z.infer<typeof formSchema>) => Promise<void>;
  createdToken?: string;
};

const UpsertAgent = ({
  open,
  agent,
  onClose,
  onSubmit,
  createdToken,
}: UpsertAgentProps) => {
  const form = useForm({
    defaultValues: formInitialValues(agent),
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: ({ formApi, value }) => {
      onSubmit(value).then(() => formApi.reset());
    },
  });

  useEffect(() => {
    if (open) {
      form.reset(formInitialValues(agent));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, agent?.id]);

  // If we are in create mode (no agent) and have a created token, show instructions instead of form
  const isCreateMode = !agent;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="w-full">
        <DialogHeader>
          <DialogTitle>
            {agent
              ? "Edit Agent"
              : createdToken
                ? "Agent Created"
                : "Add Agent"}
          </DialogTitle>
          <DialogDescription>
            {agent
              ? "Update the agent details and click on the Update button."
              : createdToken
                ? "The agent has been created successfully. Here are the environment variables to connect your agent to the Sentinel hub:"
                : "Fill out the form below to create a new agent."}
          </DialogDescription>
        </DialogHeader>
        {isCreateMode && createdToken ? (
          <>
            <Alert>
              <AlertTitle>Important!</AlertTitle>
              <AlertDescription>
                Please copy and store this token securely. You won't be able to
                see it again!
              </AlertDescription>
            </Alert>

            <P>Example for environment variables:</P>
            <CodeBlock
              code={environmentsCodeBlock(createdToken)}
              language="markup"
              className="text-sm"
            />

            <P>Example for YAML file:</P>
            <CodeBlock
              code={yamlCodeBlock(createdToken)}
              language="yaml"
              className="text-sm"
            />

            <DialogFooter>
              <DialogClose asChild>
                <Button variant="outline">Close</Button>
              </DialogClose>
            </DialogFooter>
          </>
        ) : (
          <>
            <form
              id="upsert-agent-form"
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
                          placeholder="Agent name"
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
                        <FieldLabel htmlFor={field.name}>
                          Description
                        </FieldLabel>
                        <Input
                          id={field.name}
                          name={field.name}
                          value={field.state.value}
                          onBlur={field.handleBlur}
                          onChange={(e) => field.handleChange(e.target.value)}
                          aria-invalid={isInvalid}
                          placeholder="Agent description"
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

                <form.Field
                  name="tags"
                  children={(field) => (
                    <Field>
                      <FieldLabel htmlFor={field.name}>Tags</FieldLabel>
                      <InputTag
                        tags={field.state.value.map(
                          (tag: string, index: number) => ({
                            id: index.toString(),
                            text: tag,
                          }),
                        )}
                        setTags={(tags) => {
                          const items: string[] = [];
                          for (const tag of tags as { text: string }[]) {
                            items.push(tag.text);
                          }

                          field.handleChange(items);
                        }}
                      />
                    </Field>
                  )}
                />
              </FieldGroup>
            </form>
            <DialogFooter className="mt-2">
              <DialogClose asChild>
                <Button variant="outline">Cancel</Button>
              </DialogClose>
              <Button disabled={!form.state.isValid} form="upsert-agent-form">
                {agent ? "Update" : "Create"}
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
};

const AgentsPage = () => {
  const [deleteAgentId, setDeleteAgentId] = useState<string | undefined>();
  const [upsertAgentOpen, setUpsertAgentOpen] = useState(false);
  const [agent, setAgent] = useState<Agent | undefined>();
  const [createdToken, setCreatedToken] = useState<string | undefined>();

  const q = useQuery(agentsList);

  const createMutation = useMutation(agentsCreate, {
    onSuccess: (res) => {
      toast.success("Agent created successfully");
      setCreatedToken(res?.token);
      void q.refetch();
    },
    onError: (error) => {
      toast.error((error as Error).message || "Failed to create agent");
    },
  });

  const updateMutation = useMutation(agentsUpdate, {
    onSuccess: () => {
      toast.success("Agent updated successfully");
      void q.refetch();
    },
    onError: (error) => {
      toast.error((error as Error).message || "Failed to update agent");
    },
  });

  const deleteMutation = useMutation(agentsDelete, {
    onSuccess: () => {
      toast.success("Agent deleted successfully");
      void q.refetch();
      setDeleteAgentId(undefined);
    },
    onError: (error) => {
      toast.error((error as Error).message || "Failed to delete agent");
    },
  });

  const onCreateAgent = useCallback(
    async (data: z.infer<typeof formSchema>) => {
      await createMutation.mutateAsync(data);
    },
    [createMutation],
  );

  const onUpdateAgent = useCallback(
    async (data: z.infer<typeof formSchema>) => {
      if (!agent) return;
      await updateMutation.mutateAsync({ id: agent.id, ...data });
    },
    [agent, updateMutation],
  );

  const onDeleteAgent = useCallback(async () => {
    if (!deleteAgentId) return;
    await deleteMutation.mutateAsync({ id: deleteAgentId });
  }, [deleteAgentId, deleteMutation]);

  // setup subscription to refetch on new events
  const subRef = useRef<{ abort: () => void } | null>(null);
  useSubscriptionRefetch({
    refetch: q.refetch,
    subscribe: (signal) =>
      agentsClient.agentsSubscribe(create(AgentsSubscribeRequestSchema), {
        signal,
      }),
    debounceTimeout: 300,
    onErrorFn: (error) => {
      toast.error(`subscription error: ${error}`);
    },
    ref: subRef,
  });

  return (
    <>
      <ConfirmDialog
        open={!!deleteAgentId}
        setOpen={() => setDeleteAgentId(undefined)}
        onSubmit={onDeleteAgent}
        title="Delete an agent"
        description="Are you sure you want to delete this agent?"
        type="delete"
      />
      <UpsertAgent
        open={upsertAgentOpen}
        agent={agent}
        onClose={() => {
          setUpsertAgentOpen(false);
          setAgent(undefined);
          setCreatedToken(undefined);
        }}
        onSubmit={async (data) => {
          if (agent) {
            onUpdateAgent(data);
          } else {
            await onCreateAgent(data);
            // Keep dialog open to show token content
          }
          if (agent) {
            setUpsertAgentOpen(false);
            setAgent(undefined);
          }
        }}
        createdToken={createdToken}
      />
      <Card className="gap-3">
        <CardHeader>
          <CardTitle className="text-2xl">Agents</CardTitle>
          <CardDescription>Manage application agents</CardDescription>

          <CardAction>
            <Button
              size="sm"
              variant="secondary"
              onClick={() => {
                setAgent(undefined);
                setUpsertAgentOpen(true);
              }}
            >
              <PlusIcon size={16} /> Add Agent
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <Separator className="my-2 hidden md:block" />
          <P className="text-muted-foreground text-sm leading-relaxed">
            Agents embed into your internal infrastructure, collect data for
            specified services, and send it to the Sentinel hub.
            <br />
            On this page, you can create, update, and delete agents, as well as
            check their status.
          </P>

          <P className="text-muted-foreground text-sm leading-relaxed">
            To connect an agent to this Sentinel hub, you need to set the
            following environment variables in your agent's environment:
          </P>

          <CodeBlock
            code={environmentsCodeBlock()}
            language="markup"
            className="mt-4 text-sm"
          />

          <P className="text-muted-foreground text-sm leading-relaxed">
            Alternatively, you can use a YAML configuration file with the
            following content:
          </P>

          <CodeBlock
            code={yamlCodeBlock()}
            language="yaml"
            className="mt-4 text-sm"
          />

          <Separator className="my-5" />

          <div className="overflow-hidden rounded-md border">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted/50">
                  <TableHead>Name</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Token</TableHead>
                  <TableHead>Is enabled</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Last seen at</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {q.data?.items?.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="h-24 text-center">
                      No agents found.
                    </TableCell>
                  </TableRow>
                ) : (
                  q.data?.items?.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>{item.name}</TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            item.status === AgentStatus.ACTIVE
                              ? "success"
                              : item.status === AgentStatus.INACTIVE
                                ? "error"
                                : item.status === AgentStatus.UNSPECIFIED
                                  ? "warning"
                                  : "info"
                          }
                        >
                          {statusTypeName(item.status)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {item.tokenHint ? (
                          <span className="font-mono">{item.tokenHint}</span>
                        ) : (
                          <span className="text-muted-foreground">N/A</span>
                        )}
                      </TableCell>
                      <TableCell>{item.isEnabled ? "Yes" : "No"}</TableCell>
                      <TableCell>{item.systemInfo?.version || "N/A"}</TableCell>
                      <TableCell>
                        {item.lastSeenAt
                          ? timestampDate(item.lastSeenAt).toLocaleString()
                          : "N/A"}
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
                            <DropdownMenuGroup>
                              <DropdownMenuItem
                                onClick={() => {
                                  setAgent(item);
                                  setUpsertAgentOpen(true);
                                }}
                              >
                                Edit
                              </DropdownMenuItem>
                              {/* <DropdownMenuItem>Refresh token</DropdownMenuItem> */}
                            </DropdownMenuGroup>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive focus:text-destructive"
                              onClick={() => {
                                setDeleteAgentId(item?.id);
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
        </CardContent>
      </Card>
    </>
  );
};

export default AgentsPage;
