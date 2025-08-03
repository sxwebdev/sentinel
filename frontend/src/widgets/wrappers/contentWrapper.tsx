import { ServerInfo } from "@/features/apiInfo/server-info";
import { UpdateBanner } from "@/features/apiInfo/update-banner";
import { type PropsWithChildren } from "react";

const ContentWrapper = ({ children }: PropsWithChildren) => {
  return (
    <div className="flex justify-center">
      <div className="flex flex-col p-3 md:p-5 xl:px-0 w-full max-w-6xl mx-auto">
        <UpdateBanner />
        {children}
        <ServerInfo />
      </div>
    </div>
  );
};

export default ContentWrapper;
