import { type PropsWithChildren } from "react";

const ContentWrapper = ({ children }: PropsWithChildren) => {
  return (
    <div className="flex justify-center">
      <div className="flex flex-col p-3 md:p-5 xl:px-0 w-full max-w-6xl mx-auto">
        {children}
      </div>
    </div>
  );
};

export default ContentWrapper;
