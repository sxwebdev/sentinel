import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/components/ui";
import { useIsMobile } from "@/shared/hooks/useIsMobile";
import { cn } from "@/shared/lib/utils";

interface InfoCardStatsProps {
  title: string;
  value: string;
}

export const InfoCardStats = ({ title, value }: InfoCardStatsProps) => {
  const isMobile = useIsMobile();
  return (
    <Card
      className={cn(
        "p-6 gap-2 flex flex-col justify-center",
        isMobile && "p-4",
      )}
    >
      <CardHeader>
        <CardTitle
          className={cn(
            "text-2xl font-bold text-center",
            isMobile && "text-lg",
          )}
        >
          {value}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-center text-muted-foreground">{title}</p>
      </CardContent>
    </Card>
  );
};
