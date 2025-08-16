import { Button } from "@/shared/components/ui/button";
import { Link } from "@tanstack/react-router";

const NotFound = () => {
  return (
    <div className="py-20 flex w-full flex-col items-center justify-center">
      <h1 className="text-4xl font-bold">404</h1>
      <p className="mt-4 text-xl">Page not found</p>
      <Link to="/" className="mt-8">
        <Button>Return to home</Button>
      </Link>
    </div>
  );
};

export default NotFound;
