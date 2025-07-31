import {Badge} from "@/shared/components/ui/badge";

import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/shared/components/ui";
import type {WebServerInfoResponse} from "@/shared/types/model";
import Markdown from "react-markdown";
import {useState} from "react";

export const UpdateVersion = ({apiInfo}: {apiInfo: WebServerInfoResponse}) => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Badge className="bg-orange-500 cursor-pointer">Available update</Badge>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{apiInfo?.available_update?.tag_name}</DialogTitle>
          <DialogDescription>
            <Markdown>{apiInfo?.available_update?.description}</Markdown>
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            onClick={() => {
              setIsOpen(false);
            }}
          >
            Cancel
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              window.open(apiInfo?.available_update?.url, "_blank");
            }}
          >
            Open release
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
