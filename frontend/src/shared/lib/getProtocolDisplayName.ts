export const getProtocolDisplayName = (protocol: string) => {
  switch (protocol) {
    case "http":
      return "HTTP/HTTPS";
    case "tcp":
      return "TCP";
    case "grpc":
      return "gRPC";
  }
};
