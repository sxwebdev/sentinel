import { useState, useRef, useLayoutEffect } from "react";
import { cn } from "@/shared/lib/utils";
import { ChevronDown, ChevronUp } from "lucide-react";

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
        <button
          onClick={toggleExpansion}
          className="inline-flex items-center gap-1 mt-2 px-2 py-1 text-xs font-medium text-blue-600 bg-blue-50 hover:bg-blue-100 hover:text-blue-700 rounded-md transition-colors duration-200"
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
        </button>
      )}
    </div>
  );
};
