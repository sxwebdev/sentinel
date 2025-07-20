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
    <Card className="gap-2 flex flex-col justify-center">
      <CardHeader>
        <CardTitle className="text-xl md:text-2xl font-bold text-center">
          {value}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-center text-muted-foreground">{title}</p>
      </CardContent>
    </Card>
  );
};
