import { Spinner } from "./ui/spinner";

const PageLoader = () => {
  return (
    <div className="flex h-screen w-full items-center justify-center">
      <Spinner className="size-6" />
    </div>
  );
};

export default PageLoader;
