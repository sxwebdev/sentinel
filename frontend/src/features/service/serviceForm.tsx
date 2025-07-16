import React, {useCallback} from "react";
import {Form, Formik, FastField, Field} from "formik";
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
import type {ServiceForm as ServiceFormType} from "./types/type";
import {toast} from "sonner";
import * as Yup from "yup";
import InputTag from "@/shared/components/ui/inputTag";
interface ServiceFormProps {
  initialValues: ServiceFormType;
  onSubmit: (values: ServiceFormType) => Promise<void>;
  type: "create" | "update";
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
                <Input
                  {...field}
                  value={field.value ?? ""}
                  placeholder="localhost:50051"
                />
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
                  <Input
                    {...field}
                    value={field.value ?? ""}
                    placeholder="Service Name"
                  />
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
              <Input
                {...field}
                value={field.value ?? ""}
                placeholder="localhost:8080"
              />
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
              <Input
                {...field}
                value={field.value ?? ""}
                placeholder="Expected Response"
              />
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
      (index: number, field: string, value: unknown) => {
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
                  value={field.value ?? ""}
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
          <div className="flex flex-col gap-2">
            <Label>Timeout(milliseconds)</Label>
            <FastField name="config.http.timeout">
              {({field}: FieldProps) => (
                <Input
                  {...field}
                  value={field.value ?? ""}
                  placeholder="Timeout"
                  onChange={(e) => {
                    if (!isNaN(Number(e.target.value))) {
                      setFieldValue(
                        "config.http.timeout",
                        Number(e.target.value)
                      );
                    }
                  }}
                />
              )}
            </FastField>
          </div>
          {(values.config?.http?.endpoints || []).map((_, index: number) => (
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
                        <Input
                          {...field}
                          value={field.value ?? ""}
                          placeholder="Name"
                        />
                      )}
                    </FastField>
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label required>URL</Label>
                    <FastField name={`config.http.endpoints.${index}.url`}>
                      {({field}: FieldProps) => (
                        <Input
                          {...field}
                          value={field.value ?? ""}
                          placeholder="URL"
                        />
                      )}
                    </FastField>
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label>Method</Label>
                    <Field name={`config.http.endpoints.${index}.method`}>
                      {({field}: FieldProps) => (
                        <Select
                          value={field.value ?? ""}
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
                        <Input
                          {...field}
                          value={field.value ?? ""}
                          placeholder="Expected Status"
                          onChange={(e) => {
                            if (!isNaN(Number(e.target.value))) {
                              setFieldValue(
                                `config.http.endpoints.${index}.expected_status`,
                                Number(e.target.value)
                              );
                            }
                          }}
                        />
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
                          value={field.value ?? ""}
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
                    <FastField name={`config.http.endpoints.${index}.username`}>
                      {({field}: FieldProps) => (
                        <Input
                          {...field}
                          value={field.value ?? ""}
                          placeholder="Username"
                        />
                      )}
                    </FastField>
                  </div>
                  <div className="flex flex-col gap-2">
                    <Label>Password</Label>
                    <FastField name={`config.http.endpoints.${index}.password`}>
                      {({field}: FieldProps) => (
                        <Input
                          {...field}
                          value={field.value ?? ""}
                          placeholder="Password"
                        />
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
                        value={
                          typeof field.value === "string"
                            ? field.value
                            : field.value
                              ? JSON.stringify(field.value, null, 2)
                              : ""
                        }
                        onChange={(
                          e: React.ChangeEvent<HTMLTextAreaElement>
                        ) => {
                          setFieldValue(
                            `config.http.endpoints.${index}.headers`,
                            e.target.value
                          );
                        }}
                        placeholder={'{"Content-Type": "application/json"}'}
                      />
                    )}
                  </FastField>
                </div>
                <div className="flex flex-col gap-2">
                  <Label>Body</Label>
                  <FastField name={`config.http.endpoints.${index}.body`}>
                    {({field}: FieldProps) => (
                      <Textarea
                        {...field}
                        value={field.value ?? ""}
                        placeholder="Body"
                      />
                    )}
                  </FastField>
                </div>
              </CardContent>
            </Card>
          ))}
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

export const ServiceForm = ({
  initialValues,
  onSubmit,
  type,
}: ServiceFormProps) => {
  const grpcSchema = Yup.object({
    endpoint: Yup.string().required("GRPC endpoint is required"),
  });

  const httpSchema = Yup.object({
    endpoints: Yup.array().of(
      Yup.object({
        name: Yup.string().required("Endpoint name is required"),
        url: Yup.string().required("URL is required"),
      })
    ),
  });

  const tcpSchema = Yup.object({
    endpoint: Yup.string().required("TCP endpoint is required"),
  });

  const validateSchema = Yup.object().shape({
    name: Yup.string().required("Name is required"),
    protocol: Yup.string()
      .oneOf(["grpc", "http", "tcp"])
      .required("Protocol is required"),
  });

  const headersModificate = (values: ServiceFormType) => {
    if (values.config?.http?.endpoints) {
      values.config.http.endpoints.forEach((endpoint) => {
        if (endpoint.headers) {
          try {
            endpoint.headers = JSON.parse(endpoint.headers);
          } catch {
            toast.error("Invalid headers format");
            delete endpoint.headers;
          }
        } else {
          delete endpoint.headers;
        }
      });
    }
  };

  const configModificate = (values: ServiceFormType) => {
    switch (values.protocol) {
      case "http":
        values.config.grpc = null;
        values.config.tcp = null;
        headersModificate(values);
        break;
      case "tcp":
        values.config.grpc = null;
        values.config.http = null;
        break;
      case "grpc":
        values.config.http = null;
        values.config.tcp = null;
        break;
    }
    return values;
  };

  return (
    <Formik
      initialValues={initialValues}
      validationSchema={validateSchema}
      enableReinitialize
      onSubmit={(values, {setSubmitting}) => {
        configModificate(values);
        onSubmit(values).finally(() => {
          setSubmitting(false);
        });
      }}
      validate={async (values) => {
        try {
          if (values.protocol) {
            switch (values.protocol) {
              case "http":
                await httpSchema.validate(values.config.http, {
                  abortEarly: false,
                });
                break;
              case "tcp":
                await tcpSchema.validate(values.config.tcp, {
                  abortEarly: false,
                });
                break;
              case "grpc":
                await grpcSchema.validate(values.config.grpc, {
                  abortEarly: false,
                });
                break;
            }
          }
          return {};
        } catch (err) {
          if (err instanceof Yup.ValidationError) {
            const errors: Record<string, string> = {};
            err.inner.forEach((e) => {
              if (e.path) {
                errors[e.path] = e.message;
              }
            });
            return errors;
          }
          throw err;
        }
      }}
    >
      {({isSubmitting, isValid, dirty, values, setFieldValue}) => {
        return (
          <Form className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <Label required>Service Name</Label>
              <FastField name="name">
                {({field}: FieldProps) => (
                  <Input
                    {...field}
                    value={field.value ?? ""}
                    placeholder="Name"
                  />
                )}
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
                      <SelectValue
                        className="w-full"
                        placeholder="Select Protocol"
                      />
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
                      value={field.value ?? ""}
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
                      value={field.value ?? ""}
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
                      value={field.value ?? ""}
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
                  <InputTag
                    tags={field.value.map((tag: string, index: number) => ({
                      id: index.toString(),
                      text: tag,
                    }))}
                    setTags={(tags) => {
                      setFieldValue(
                        "tags",
                        typeof tags === "object"
                          ? tags.map((tag) => tag.text)
                          : []
                      );
                    }}
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
            {values.protocol === "grpc" && (
              <GRPCForm setFieldValue={setFieldValue} />
            )}
            <hr />
            <div className="flex justify-end">
              <Button
                type="submit"
                disabled={isSubmitting || !isValid || !dirty}
              >
                {type === "create" ? "Create" : "Update"}
              </Button>
            </div>
          </Form>
        );
      }}
    </Formik>
  );
};
