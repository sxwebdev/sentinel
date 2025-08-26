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
    <Card className="flex flex-col justify-center gap-2">
      <CardHeader>
        <CardTitle className="text-center text-xl font-bold md:text-2xl">
          {value}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-muted-foreground text-center">{title}</p>
      </CardContent>
    </Card>
  );
};
