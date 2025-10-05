import { cn } from "@/shared/lib/utils";

export function H1({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <h1
      className={cn(
        "scroll-m-20 text-4xl font-extrabold tracking-tight text-balance",
        className,
      )}
    >
      {children}
    </h1>
  );
}

export function H2({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <h2
      className={cn(
        "scroll-m-20 border-b pb-2 text-3xl font-semibold tracking-tight first:mt-0",
        className,
      )}
    >
      {children}
    </h2>
  );
}

export function H3({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <h3
      className={cn(
        "scroll-m-20 text-2xl font-semibold tracking-tight",
        className,
      )}
    >
      {children}
    </h3>
  );
}

export function H4({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <h4
      className={cn(
        "scroll-m-20 text-xl font-medium tracking-tight",
        className,
      )}
    >
      {children}
    </h4>
  );
}

export function H5({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <h5
      className={cn(
        "scroll-m-20 text-lg font-medium tracking-tight",
        className,
      )}
    >
      {children}
    </h5>
  );
}

export function H6({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <h6
      className={cn(
        "scroll-m-20 text-base font-medium tracking-tight",
        className,
      )}
    >
      {children}
    </h6>
  );
}

export function P({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <p className={cn("leading-7 [&:not(:first-child)]:mt-3", className)}>
      {children}
    </p>
  );
}

export function Blockquote({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <blockquote className={cn("mt-6 border-l-2 pl-6 italic", className)}>
      {children}
    </blockquote>
  );
}

export function InlineCode({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <code
      className={cn(
        "bg-muted relative rounded px-[0.3rem] py-[0.2rem] font-mono text-sm",
        className,
      )}
    >
      {children}
    </code>
  );
}

export function BlockCode({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <pre
      className={cn(
        "bg-muted relative rounded px-[0.3rem] py-[0.2rem] font-mono text-sm",
        className,
      )}
    >
      {children}
    </pre>
  );
}
