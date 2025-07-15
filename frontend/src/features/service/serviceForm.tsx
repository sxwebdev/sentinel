import React, {useEffect, useCallback} from "react";
import {Form, Formik, FastField, Field, useFormikContext} from "formik";
import type {FieldProps} from "formik";
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
import type {HTTPEndpoint, ServiceForm as ServiceFormType} from "./types/type";

interface ServiceFormProps {
  initialValues: ServiceFormType;
  // resetForm можно использовать для сброса формы после успешного создания
  onSubmit: (values: ServiceFormType) => Promise<void> | void;
}

const GRPCForm = React.memo(
  ({
    setFieldValue,
  }: {
    setFieldValue: (field: string, value: unknown) => void;
  }) => {
    return (
      <Card>
        <CardHeader>
          <CardTitle>gRPC Configuration</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label required>Endpoint</Label>
            <FastField name="config.grpc.endpoint">
              {({field}: FieldProps) => (
                <Input {...field} placeholder="localhost:50051" />
              )}
            </FastField>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="flex flex-col gap-2">
              <Label required>Check Type</Label>
              <Field name="config.grpc.check_type">
                {({field}: FieldProps) => (
                  <Select
                    value={field.value}
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
                )}
              </Field>
            </div>
            <div className="flex flex-col gap-2">
              <Label>Service Name</Label>
              <FastField name="config.grpc.service_name">
                {({field}: FieldProps) => (
                  <Input {...field} placeholder="Service Name" />
                )}
              </FastField>
            </div>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="flex flex-col gap-2">
              <Label>Use TLS</Label>
              <Field name="config.grpc.tls">
                {({field}: FieldProps) => (
                  <Switch
                    checked={field.value}
                    onCheckedChange={(checked) =>
                      setFieldValue("config.grpc.tls", checked)
                    }
                  />
                )}
              </Field>
            </div>
            <div className="flex flex-col gap-2">
              <Label>Insecure TLS</Label>
              <Field name="config.grpc.insecure_tls">
                {({field}: FieldProps) => (
                  <Switch
                    checked={field.value}
                    onCheckedChange={(checked) =>
                      setFieldValue("config.grpc.insecure_tls", checked)
                    }
                  />
                )}
              </Field>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }
);

const TCPForm = React.memo(() => {
  return (
    <Card>
      <CardHeader>
        <CardTitle>TCP Configuration</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          <Label required>Endpoint</Label>
          <FastField name="config.tcp.endpoint">
            {({field}: FieldProps) => (
              <Input {...field} placeholder="localhost:8080" />
            )}
          </FastField>
        </div>
        <div className="flex flex-col gap-2">
          <Label>Send Data</Label>
          <FastField name="config.tcp.send_data">
            {({field}: FieldProps) => (
              <Textarea {...field} placeholder="Send Data" />
            )}
          </FastField>
        </div>
        <div className="flex flex-col gap-2">
          <Label>Expected Response</Label>
          <FastField name="config.tcp.expect_data">
            {({field}: FieldProps) => (
              <Input {...field} placeholder="Expected Response" />
            )}
          </FastField>
        </div>
      </CardContent>
    </Card>
  );
});

const HTTPForm = React.memo(
  ({
    values,
    setFieldValue,
  }: {
    values: ServiceFormType;
    setFieldValue: (field: string, value: unknown) => void;
  }) => {
    // Мемоизированные обработчики
    const handleEndpointChange = useCallback(
      (index: number, field: string, value: any) => {
        const endpoints = [...(values.config?.http?.endpoints || [])];
        endpoints[index] = {...endpoints[index], [field]: value};
        setFieldValue("config.http.endpoints", endpoints);
      },
      [setFieldValue, values.config?.http?.endpoints]
    );

    const handleRemoveEndpoint = useCallback(
      (index: number) => {
        setFieldValue(
          "config.http.endpoints",
          (values.config?.http?.endpoints || []).filter((_, i) => i !== index)
        );
      },
      [setFieldValue, values.config?.http?.endpoints]
    );

    const handleAddEndpoint = useCallback(() => {
      setFieldValue("config.http.endpoints", [
        ...(values.config?.http?.endpoints || []),
        {
          name: "",
          url: "",
          method: "GET",
          expected_status: 200,
          json_path: "",
          headers: "",
          username: "",
          password: "",
          body: "",
        },
      ]);
    }, [setFieldValue, values.config?.http?.endpoints]);

    return (
      <Card>
        <CardHeader>
          <CardTitle>HTTP Configuration</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label>JavaScript Condition</Label>
            <FastField name="config.http.condition">
              {({field}: FieldProps) => (
                <Textarea
                  {...field}
                  placeholder="// Example: return Math.abs(results.main.value - results.backup.value) > 5;"
                />
              )}
            </FastField>
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
          {(values.config?.http?.endpoints || []).map(
            (endpoint: HTTPEndpoint, index: number) => (
              <Card key={index}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle>Endpoint {index + 1}</CardTitle>
                    <Button
                      type="button"
                      variant="destructive"
                      size="sm"
                      onClick={() => handleRemoveEndpoint(index)}
                    >
                      <TrashIcon /> Remove
                    </Button>
                  </div>
                </CardHeader>
                <CardContent className="flex flex-col gap-4">
                  <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <div className="flex flex-col gap-2">
                      <Label required>Name</Label>
                      <FastField name={`config.http.endpoints.${index}.name`}>
                        {({field}: FieldProps) => (
                          <Input {...field} placeholder="Name" />
                        )}
                      </FastField>
                    </div>
                    <div className="flex flex-col gap-2">
                      <Label required>URL</Label>
                      <FastField name={`config.http.endpoints.${index}.url`}>
                        {({field}: FieldProps) => (
                          <Input {...field} placeholder="URL" />
                        )}
                      </FastField>
                    </div>
                    <div className="flex flex-col gap-2">
                      <Label>Method</Label>
                      <Field name={`config.http.endpoints.${index}.method`}>
                        {({field}: FieldProps) => (
                          <Select
                            value={field.value}
                            onValueChange={(value) =>
                              handleEndpointChange(index, "method", value)
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
                        )}
                      </Field>
                    </div>
                    <div className="flex flex-col gap-2">
                      <Label>Expected Status</Label>
                      <FastField
                        name={`config.http.endpoints.${index}.expected_status`}
                      >
                        {({field}: FieldProps) => (
                          <Input {...field} placeholder="Expected Status" />
                        )}
                      </FastField>
                    </div>
                    <div className="flex flex-col gap-2">
                      <Label>JSON Path</Label>
                      <FastField
                        name={`config.http.endpoints.${index}.json_path`}
                      >
                        {({field}: FieldProps) => (
                          <Input
                            {...field}
                            placeholder="results.block_number"
                          />
                        )}
                      </FastField>
                      <small className="text-muted-foreground text-xs">
                        Path to extract value from JSON response (e.g.,
                        "result.block_number")
                      </small>
                    </div>
                  </div>
                  <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <div className="flex flex-col gap-2">
                      <Label>Username</Label>
                      <FastField
                        name={`config.http.endpoints.${index}.username`}
                      >
                        {({field}: FieldProps) => (
                          <Input {...field} placeholder="Username" />
                        )}
                      </FastField>
                    </div>
                    <div className="flex flex-col gap-2">
                      <Label>Password</Label>
                      <FastField
                        name={`config.http.endpoints.${index}.password`}
                      >
                        {({field}: FieldProps) => (
                          <Input {...field} placeholder="Password" />
                        )}
                      </FastField>
                    </div>
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label>Headers</Label>
                    <FastField name={`config.http.endpoints.${index}.headers`}>
                      {({field}: FieldProps) => (
                        <Textarea
                          {...field}
                          placeholder="Content-Type: application/json"
                        />
                      )}
                    </FastField>
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label>Body</Label>
                    <FastField name={`config.http.endpoints.${index}.body`}>
                      {({field}: FieldProps) => (
                        <Textarea {...field} placeholder="Body" />
                      )}
                    </FastField>
                  </div>
                </CardContent>
              </Card>
            )
          )}
          <Button
            type="button"
            className="w-fit"
            variant="outline"
            onClick={handleAddEndpoint}
          >
            <PlusIcon /> Add Endpoint
          </Button>
        </CardContent>
      </Card>
    );
  }
);

const ServiceFormInner = () => {
  const {values, setFieldValue} = useFormikContext<ServiceFormType>();
  useEffect(() => {
    if (values.protocol === "http") {
      if (
        !values.config?.http?.endpoints ||
        values.config.http.endpoints.length === 0
      ) {
        setFieldValue("config.http.timeout", 10000);
        setFieldValue("config.http.condition", "");
        setFieldValue("config.http.endpoints", [
          {
            name: "",
            url: "",
            method: "GET",
            expected_status: 200,
            json_path: "",
            headers: "",
            username: "",
            password: "",
            body: "",
          },
        ]);
      }
      if (values.config?.tcp) setFieldValue("config.tcp", null);
      if (values.config?.grpc) setFieldValue("config.grpc", null);
    }
    if (values.protocol === "tcp") {
      if (!values.config?.tcp) {
        setFieldValue("config.tcp", {
          endpoint: "",
          send_data: "",
          expect_data: "",
        });
      }
      if (values.config?.http) setFieldValue("config.http", null);
      if (values.config?.grpc) setFieldValue("config.grpc", null);
    }
    if (values.protocol === "grpc") {
      if (!values.config?.grpc) {
        setFieldValue("config.grpc", {
          check_type: "health",
          endpoint: "",
          tls: true,
          service_name: "",
          insecure_tls: false,
        });
      } else {
        if (values.config.grpc.tls === undefined) {
          setFieldValue("config.grpc.tls", true);
        }
        if (!values.config.grpc.check_type) {
          setFieldValue("config.grpc.check_type", "health");
        }
      }
      if (values.config?.http) setFieldValue("config.http", null);
      if (values.config?.tcp) setFieldValue("config.tcp", null);
    }
  }, [
    values.protocol,
    values.config?.http?.endpoints,
    values.config?.grpc,
    values.config?.tcp,
    setFieldValue,
  ]);
  return (
    <Form className="flex flex-col gap-4">
      <div className="flex flex-col gap-2">
        <Label required>Service Name</Label>
        <FastField name="name">
          {({field}: FieldProps) => <Input {...field} placeholder="Name" />}
        </FastField>
      </div>
      <div className="flex flex-col gap-2">
        <Label required>Protocol</Label>
        <Field name="protocol">
          {({field}: FieldProps) => (
            <Select
              value={field.value}
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
          )}
        </Field>
      </div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div className="flex flex-col gap-2">
          <Label>Interval(milliseconds)</Label>
          <FastField name="interval">
            {({field}: FieldProps) => (
              <Input
                {...field}
                placeholder="Interval"
                onChange={(e) => {
                  if (!isNaN(Number(e.target.value))) {
                    setFieldValue("interval", Number(e.target.value));
                  }
                }}
              />
            )}
          </FastField>
        </div>
        <div className="flex flex-col gap-2">
          <Label>Timeout(milliseconds)</Label>
          <FastField name="timeout">
            {({field}: FieldProps) => (
              <Input
                {...field}
                placeholder="Timeout"
                onChange={(e) => {
                  if (!isNaN(Number(e.target.value))) {
                    setFieldValue("timeout", Number(e.target.value));
                  }
                }}
              />
            )}
          </FastField>
        </div>
        <div className="flex flex-col gap-2">
          <Label>Retries</Label>
          <FastField name="retries">
            {({field}: FieldProps) => (
              <Input
                {...field}
                placeholder="Retries"
                onChange={(e) => {
                  if (!isNaN(Number(e.target.value))) {
                    setFieldValue("retries", Number(e.target.value));
                  }
                }}
              />
            )}
          </FastField>
        </div>
      </div>
      <div className="flex flex-col gap-2">
        <Label>Tags (comma-separated)</Label>
        <FastField name="tags">
          {({field}: FieldProps) => (
            <Input
              {...field}
              placeholder="api, critical, production"
              value={field.value}
              onChange={(e) => setFieldValue("tags", e.target.value)}
            />
          )}
        </FastField>
      </div>
      <div className="flex flex-col gap-2">
        <Label>Enabled Service</Label>
        <Field name="is_enabled">
          {({field}: FieldProps) => (
            <Switch
              checked={field.value}
              onCheckedChange={(checked) =>
                setFieldValue("is_enabled", checked)
              }
            />
          )}
        </Field>
      </div>
      {/*  HTTP/HTTPS */}
      {values.protocol === "http" && (
        <HTTPForm values={values} setFieldValue={setFieldValue} />
      )}
      {/*  TCP */}
      {values.protocol === "tcp" && <TCPForm />}
      {/*  GRPC */}
      {values.protocol === "grpc" && <GRPCForm setFieldValue={setFieldValue} />}
      <hr />
      <div className="flex justify-end">
        <Button type="submit">Create</Button>
      </div>
    </Form>
  );
};

export const ServiceForm = ({initialValues, onSubmit}: ServiceFormProps) => {
  return (
    <Formik
      initialValues={initialValues}
      enableReinitialize
      onSubmit={(values) => {
        if (values.tags) {
          values.tags = values.tags.split(",").map((tag: string) => tag.trim());
        }
        if (values.config?.http?.endpoints) {
          values.config.http.endpoints.forEach((endpoint) => {
            if (endpoint.headers) {
              endpoint.headers = JSON.parse(endpoint.headers);
            }
          });
        }
        onSubmit(values);
      }}
    >
      <ServiceFormInner />
    </Formik>
  );
};
