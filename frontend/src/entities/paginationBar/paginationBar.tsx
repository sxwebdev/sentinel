import React from "react";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationPrevious,
  PaginationNext,
  PaginationEllipsis,
} from "@/shared/components/ui";

interface PaginationBarProps {
  page: number;
  setPage: (page: number) => void;
  total: number;
  pageSize?: number;
  siblingCount?: number; // сколько кнопок вокруг текущей
}

export const PaginationBar: React.FC<PaginationBarProps> = ({
  page,
  setPage,
  total,
  pageSize = 10,
  siblingCount = 1,
}) => {
  const totalPages = Math.ceil(total / pageSize);

  if (totalPages <= 1) return null;

  const getPageNumbers = () => {
    const pages: (number | string)[] = [];
    const leftSibling = Math.max(page - siblingCount, 1);
    const rightSibling = Math.min(page + siblingCount, totalPages);

    if (leftSibling > 2) {
      pages.push(1, "...");
    } else {
      for (let i = 1; i < leftSibling; i++) pages.push(i);
    }

    for (let i = leftSibling; i <= rightSibling; i++) pages.push(i);

    if (rightSibling < totalPages - 1) {
      pages.push("...", totalPages);
    } else {
      for (let i = rightSibling + 1; i <= totalPages; i++) pages.push(i);
    }

    return pages;
  };

  const handlePageChange = (p: number) => {
    if (p < 1 || p > totalPages || p === page) return;
    setPage(p);
  };

  return (
    <div className="flex justify-end px-6">
      <div>
        <Pagination>
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious
                size="sm"
                onClick={() => handlePageChange(page - 1)}
                aria-disabled={page === 1}
                tabIndex={page === 1 ? -1 : 0}
              />
            </PaginationItem>
            {getPageNumbers().map((p, idx) =>
              typeof p === "number" ? (
                <PaginationItem key={p}>
                  <PaginationLink
                    size="sm"
                    isActive={p === page}
                    onClick={() => handlePageChange(p)}
                  >
                    {p}
                  </PaginationLink>
                </PaginationItem>
              ) : (
                <PaginationItem key={`ellipsis-${idx}`}>
                  <PaginationEllipsis />
                </PaginationItem>
              )
            )}
            <PaginationItem>
              <PaginationNext
                size="sm"
                onClick={() => handlePageChange(page + 1)}
                aria-disabled={page === totalPages}
                tabIndex={page === totalPages ? -1 : 0}
              />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </div>
    </div>
  );
};
