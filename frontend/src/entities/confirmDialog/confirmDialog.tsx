import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogAction,
  AlertDialogCancel,
} from "@/shared/components/ui/alert-dialog";

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  type: "delete" | "default";
  description?: string;
  content?: React.ReactNode;
  setOpen: (open: boolean | null) => void;
  onSubmit: () => void;
}

export const ConfirmDialog = ({
  open,
  setOpen,
  onSubmit,
  title,
  type = "default",
  description,
  content,
}: ConfirmDialogProps) => {
  return (
    <AlertDialog open={open} onOpenChange={() => setOpen(null)}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        {content && content}
        <AlertDialogFooter>
          <AlertDialogCancel onClick={() => setOpen(null)}>
            Cancel
          </AlertDialogCancel>
          {type === "delete" && (
            <AlertDialogAction
              className="bg-destructive hover:bg-destructive/90 dark:text-white"
              onClick={onSubmit}
            >
              Delete
            </AlertDialogAction>
          )}
          {type === "default" && (
            <AlertDialogAction onClick={onSubmit}></AlertDialogAction>
          )}
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};
