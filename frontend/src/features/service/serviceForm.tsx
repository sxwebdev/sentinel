import {Form, Formik} from "formik";
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Input,
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Switch,
  Textarea,
} from "@/shared/components/ui";
import {PlusIcon, TrashIcon} from "lucide-react";

interface HTTP {
  condition?: string;
  endpoints?: {
    body?: string;
    expected_status?: number;
    headers?: {key: string; value: string};
    json_path?: string;
    method?: string;
    name?: string;
    password?: string;
    url?: string;
    username?: string;
  }[];
  timeout?: number;
}

interface TCP {
  endpoint?: string;
  expect_data?: string;
  send_data?: string;
}

interface GRPC {
  check_type?: string;
  endpoint?: string;
  service_name?: string;
  tls?: boolean;
  insecure_tls?: boolean;
}

interface CreateService {
  name: string;
  protocol: string;
  interval?: number;
  timeout?: number;
  retries?: number;
  tags?: string[];
  is_enabled?: boolean;
  config?: {
    http?: HTTP;
    tcp?: TCP;
    grpc?: GRPC;
  };
}

interface ServiceFormProps {
  initialValues: CreateService;
}

const GRPCForm = ({
  values,
  setFieldValue,
}: {
  values: CreateService;
  setFieldValue: (field: string, value: any) => void;
}) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>gRPC Configuration</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          <Label required>Endpoint</Label>
          <Input
            name="config.grpc.endpoint"
            placeholder="localhost:50051"
            value={values.config?.grpc?.endpoint}
            onChange={(e) =>
              setFieldValue("config.grpc.endpoint", e.target.value)
            }
          />
        </div>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div className="flex flex-col gap-2">
            <Label>Check Type</Label>
            <Select
              name="config.grpc.check_type"
              value={values.config?.grpc?.check_type}
              onValueChange={(value) =>
                setFieldValue("config.grpc.check_type", value)
              }
            >
              <SelectTrigger className="w-full">
                <SelectValue
                  className="w-full"
                  placeholder="Select Check Type"
                />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="health">Health Check</SelectItem>
                <SelectItem value="connectivity">Connectivity</SelectItem>
                <SelectItem value="reflection">Reflection</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-2">
            <Label>Service Name</Label>
            <Input
              name="config.grpc.service_name"
              placeholder="Service Name"
              value={values.config?.grpc?.service_name}
              onChange={(e) =>
                setFieldValue("config.grpc.service_name", e.target.value)
              }
            />
          </div>
        </div>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div className="flex flex-col gap-2">
            <Label>Use TLS</Label>
            <Switch
              checked={values.config?.grpc?.tls}
              onCheckedChange={(checked) =>
                setFieldValue("config.grpc.tls", checked)
              }
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label>Insecure TLS</Label>
            <Switch
              checked={values.config?.grpc?.insecure_tls}
              onCheckedChange={(checked) =>
                setFieldValue("config.grpc.insecure_tls", checked)
              }
            />
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

const TCPForm = ({
  values,
  setFieldValue,
}: {
  values: CreateService;
  setFieldValue: (field: string, value: any) => void;
}) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>TCP Configuration</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          <Label required>Endpoint</Label>
          <Input
            name="endpoint"
            placeholder="localhost:8080"
            value={values.config?.tcp?.endpoint}
            onChange={(e) => setFieldValue("endpoint", e.target.value)}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label>Send Data</Label>
          <Textarea
            name="config.tcp.send_data"
            placeholder="Send Data"
            value={values.config?.tcp?.send_data}
            onChange={(e) =>
              setFieldValue("config.tcp.send_data", e.target.value)
            }
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label>Expected Response</Label>
          <Input
            name="config.tcp.expect_data"
            placeholder="Expected Response"
            value={values.config?.tcp?.expect_data}
            onChange={(e) =>
              setFieldValue("config.tcp.expect_data", e.target.value)
            }
          />
        </div>
      </CardContent>
    </Card>
  );
};

const HTTPForm = ({
  values,
  setFieldValue,
}: {
  values: CreateService;
  setFieldValue: (field: string, value: any) => void;
}) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>HTTP Configuration</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          <Label>JavaScript Condition</Label>
          <div>
            <Textarea
              name="config.http.condition"
              placeholder="// Example: return Math.abs(results.main.value - results.backup.value) > 5;"
              value={values.config?.http?.condition}
              onChange={(e) =>
                setFieldValue("config.http.condition", e.target.value)
              }
            />
            <small className="text-muted-foreground text-xs">
              JavaScript condition that returns true to trigger an incident.
              Available variables:
              <code className="text-xs font-mono">
                results.endpoint_name.value
              </code>
              ,{" "}
              <code className="text-xs font-mono">
                results.endpoint_name.success
              </code>
              , etc.
            </small>
          </div>
        </div>
        {values.config?.http?.endpoints?.map((endpoint, index) => (
          <Card key={index}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Endpoint {index + 1}</CardTitle>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => {
                    setFieldValue(
                      "config.http.endpoints",
                      values.config?.http?.endpoints?.filter(
                        (_, i) => i !== index
                      )
                    );
                  }}
                >
                  <TrashIcon />
                  Remove
                </Button>
              </div>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="flex flex-col gap-2">
                  <Label>Name</Label>
                  <Input
                    name={`endpoints.${index}.name`}
                    placeholder="Name"
                    value={endpoint.name}
                    onChange={(e) =>
                      setFieldValue(`endpoints.${index}.name`, e.target.value)
                    }
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label>URL</Label>
                  <Input
                    name={`endpoints.${index}.url`}
                    placeholder="URL"
                    value={endpoint.url}
                    onChange={(e) =>
                      setFieldValue(`endpoints.${index}.url`, e.target.value)
                    }
                  />
                </div>

                <div className="flex flex-col gap-2">
                  <Label>Method</Label>
                  <Select
                    name={`endpoints.${index}.method`}
                    value={endpoint.method}
                    onValueChange={(value) =>
                      setFieldValue(`endpoints.${index}.method`, value)
                    }
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue
                        className="w-full"
                        placeholder="Select Method"
                      />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="GET">GET</SelectItem>
                      <SelectItem value="POST">POST</SelectItem>
                      <SelectItem value="PUT">PUT</SelectItem>
                      <SelectItem value="DELETE">DELETE</SelectItem>
                      <SelectItem value="HEAD">HEAD</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex flex-col gap-2">
                  <Label>Expected Status</Label>
                  <Input
                    name={`endpoints.${index}.expected_status`}
                    placeholder="Expected Status"
                    value={endpoint.expected_status}
                    onChange={(e) =>
                      setFieldValue(
                        `endpoints.${index}.expected_status`,
                        e.target.value
                      )
                    }
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label>JSON Path</Label>
                  <div>
                    <Input
                      name={`endpoints.${index}.json_path`}
                      placeholder="results.block_number"
                      value={endpoint.json_path}
                      onChange={(e) =>
                        setFieldValue(
                          `endpoints.${index}.json_path`,
                          e.target.value
                        )
                      }
                    />
                    <small className="text-muted-foreground text-xs">
                      Path to extract value from JSON response (e.g.,
                      "result.block_number")
                    </small>
                  </div>
                </div>
              </div>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div>
                  <Label>Username</Label>
                  <Input
                    name={`endpoints.${index}.username`}
                    placeholder="Username"
                    value={endpoint.username}
                    onChange={(e) =>
                      setFieldValue(
                        `endpoints.${index}.username`,
                        e.target.value
                      )
                    }
                  />
                </div>
                <div>
                  <Label>Password</Label>
                  <Input
                    name={`endpoints.${index}.password`}
                    placeholder="Password"
                    value={endpoint.password}
                    onChange={(e) =>
                      setFieldValue(
                        `endpoints.${index}.password`,
                        e.target.value
                      )
                    }
                  />
                </div>
              </div>
              <div className="flex flex-col gap-2">
                <Label>Headers</Label>
                <Textarea
                  name={`endpoints.${index}.headers`}
                  placeholder="Content-Type: application/json"
                  value={endpoint.headers}
                  onChange={(e) =>
                    setFieldValue(`endpoints.${index}.headers`, e.target.value)
                  }
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label>Body</Label>
                <Textarea
                  name={`endpoints.${index}.body`}
                  placeholder="Body"
                  value={endpoint.body}
                  onChange={(e) =>
                    setFieldValue(`endpoints.${index}.body`, e.target.value)
                  }
                />
              </div>
            </CardContent>
          </Card>
        ))}
        <Button
          className="w-fit"
          variant="outline"
          onClick={() => {
            setFieldValue("endpoints", [
              ...(values.config?.http?.endpoints || []),
              {
                name: "",
                url: "",
                method: "GET",
                expected_status: 200,
                json_path: "",
                headers: "",
                body: "",
              },
            ]);
          }}
        >
          <PlusIcon /> Add Endpoint
        </Button>
      </CardContent>
    </Card>
  );
};

export const ServiceForm = ({initialValues}: ServiceFormProps) => {
  return (
    <Formik initialValues={initialValues} onSubmit={() => {}}>
      {({values, setFieldValue}) => (
        <Form className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label required>Service Name</Label>
            <Input
              name="name"
              placeholder="Name"
              value={values.name}
              onChange={(e) => setFieldValue("name", e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label required>Protocol</Label>
            <Select
              value={values.protocol}
              onValueChange={(value) => setFieldValue("protocol", value)}
            >
              <SelectTrigger className="w-full">
                <SelectValue className="w-full" placeholder="Select Protocol" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="http">HTTP/HTTPS</SelectItem>
                <SelectItem value="tcp">TCP</SelectItem>
                <SelectItem value="grpc">gRPC</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="flex flex-col gap-2">
              <Label>Interval(seconds)</Label>
              <Input
                name="interval"
                placeholder="Interval"
                value={values.interval}
                onChange={(e) => setFieldValue("interval", e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label>Timeout(seconds)</Label>
              <Input
                name="timeout"
                placeholder="Timeout"
                value={values.timeout}
                onChange={(e) => setFieldValue("timeout", e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label>Retries</Label>
              <Input
                name="retries"
                placeholder="Retries"
                value={values.retries}
                onChange={(e) => setFieldValue("retries", e.target.value)}
              />
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <Label>Tags (comma-separated)</Label>
            <Input
              name="tags"
              placeholder="api, critical, production"
              value={values.tags}
              onChange={(e) => setFieldValue("tags", e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label>Enabled Service</Label>
            <Switch
              checked={values.is_enabled}
              onCheckedChange={(checked) =>
                setFieldValue("is_enabled", checked)
              }
            />
          </div>
          {/*  HTTP/HTTPS */}
          {values.protocol === "http" && (
            <HTTPForm values={values} setFieldValue={setFieldValue} />
          )}
          {/*  TCP */}
          {values.protocol === "tcp" && (
            <TCPForm values={values} setFieldValue={setFieldValue} />
          )}
          {/*  GRPC */}
          {values.protocol === "grpc" && (
            <GRPCForm values={values} setFieldValue={setFieldValue} />
          )}
          <hr />
          <div className="flex justify-end">
            <Button type="submit">Create</Button>
          </div>
        </Form>
      )}
    </Formik>
  );
};
