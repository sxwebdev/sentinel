import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/shared/components/ui";

interface InfoCardStatsProps {
  title: string;
  value: string;
}

export const InfoCardStats = ({ title, value }: InfoCardStatsProps) => {
  return (
    <Card className="flex flex-col justify-center gap-0 py-4 md:gap-2 md:py-6">
      <CardHeader className="px-4 md:px-6">
        <CardTitle className="text-center text-xl font-bold md:text-2xl">
          {value}
        </CardTitle>
      </CardHeader>
      <CardContent className="px-4 md:px-6">
        <p className="text-muted-foreground text-center text-sm md:text-base">
          {title}
        </p>
      </CardContent>
    </Card>
  );
};
