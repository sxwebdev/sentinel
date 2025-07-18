import {
  Dialog,
  DialogTitle,
  DialogHeader,
  DialogContent,
  DialogDescription,
  DialogFooter,
  Button,
} from "@/shared/components/ui";

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
    <Dialog open={open} onOpenChange={() => setOpen(null)}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        {content && content}
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(null)}>
            Cancel
          </Button>
          {type === "delete" && (
            <Button variant="destructive" onClick={onSubmit}>
              Delete
            </Button>
          )}
          {type === "default" && <Button onClick={onSubmit}>Submit</Button>}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
