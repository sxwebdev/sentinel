import { useState, useRef, useLayoutEffect } from "react";
import { cn } from "@/shared/lib/utils";
import { ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "./ui";

interface ExpandableTextProps {
  content: string;
  className?: string;
}

export const ExpandableText = ({ content, className }: ExpandableTextProps) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isTruncated, setIsTruncated] = useState(false);
  const textRef = useRef<HTMLDivElement>(null);

  // Check if text is truncated
  useLayoutEffect(() => {
    const checkTruncation = () => {
      if (textRef.current && !isExpanded) {
        const element = textRef.current;
        const isOverflowing =
          element.scrollWidth > element.clientWidth ||
          element.scrollHeight > element.clientHeight;

        setIsTruncated(isOverflowing);
      }
    };

    checkTruncation();

    // Add resize observer for responsive behavior
    const resizeObserver = new ResizeObserver(checkTruncation);

    if (textRef.current) {
      resizeObserver.observe(textRef.current);
    }

    return () => {
      resizeObserver.disconnect();
    };
  }, [content, isExpanded]);

  const toggleExpansion = () => {
    setIsExpanded((prev) => !prev);
  };

  return (
    <div>
      <div
        ref={textRef}
        className={cn(className, !isExpanded && "line-clamp-1")}
        dangerouslySetInnerHTML={{ __html: content }}
      />

      {/* Show toggle button if content is truncated OR currently expanded */}
      {(isTruncated || isExpanded) && (
        <Button
          onClick={toggleExpansion}
          variant="secondary"
          className="mt-2 h-8 gap-1 px-1.5 py-1 text-xs"
        >
          {isExpanded ? (
            <>
              <ChevronUp size={18} />
              Show less
            </>
          ) : (
            <>
              <ChevronDown size={18} />
              Show more
            </>
          )}
        </Button>
      )}
    </div>
  );
};
