import {Button} from "@/shared/components/ui/button";
import {useNavigate} from "react-router";
import {ROUTES} from "@/app/routes/constants";

const NotFound = () => {
  const navigate = useNavigate();
  return (
    <div className="flex h-screen w-full flex-col items-center justify-center">
      <h1 className="text-4xl font-bold">404</h1>
      <p className="mt-4 text-xl">Page not found</p>
      <Button className="mt-8" onClick={() => navigate(ROUTES.DASHBOARD)}>
        Return to home
      </Button>
    </div>
  );
};

export default NotFound;
