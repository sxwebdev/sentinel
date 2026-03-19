import * as React from "react";
import { Highlight, themes, Prism, type Language } from "prism-react-renderer";
import { cn } from "@/shared/lib/utils";
import { useUserPreferenceStore } from "@/app/stores/userPreferience";
import { useShallow } from "zustand/react/shallow";

type Props = {
  code: string;
  language?: Language;
  maxHeight?: string;
  className?: string;
};

export function CodeBlock({
  code,
  language = "tsx",
  maxHeight,
  className = "",
}: Props) {
  const [copied, setCopied] = React.useState(false);
  const { currentTheme } = useUserPreferenceStore(
    useShallow((s) => ({ currentTheme: s.theme })),
  );

  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // ignore
    }
  };

  const theme = currentTheme === "dark" ? themes.oneDark : themes.github;

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-lg border border-black/10 dark:border-white/10",
        className,
      )}
    >
      <button
        onClick={onCopy}
        aria-label="Copy code"
        className="absolute top-2 right-2 z-10 cursor-pointer rounded-md bg-[#ebeced] px-2 py-1 text-xs font-medium transition hover:bg-[#dedfe1] dark:bg-[#3d4148] dark:hover:bg-[#54565d]"
        type="button"
      >
        {copied ? "Copied" : "Copy"}
      </button>

      <Highlight
        code={code.trimEnd()}
        language={language}
        theme={theme}
        prism={Prism}
      >
        {({ className, style, tokens, getLineProps, getTokenProps }) => (
          <pre
            className={`${className} m-0 overflow-auto p-4`}
            style={{ ...style, maxHeight }}
          >
            {tokens.map((line, i) => (
              <div key={i} {...getLineProps({ line })}>
                {line.map((token, key) => (
                  <span key={key} {...getTokenProps({ token })} />
                ))}
              </div>
            ))}
          </pre>
        )}
      </Highlight>
    </div>
  );
}
