import { ServerInfo } from "@/features/apiInfo/server-info";
import { UpdateBanner } from "@/features/apiInfo/update-banner";
import { useServerStore } from "@/pages/dashboard/store/useServerStore";
import { type PropsWithChildren } from "react";

const ContentWrapper = ({ children }: PropsWithChildren) => {
  const serverInfo = useServerStore((s) => s.serverInfo);

  return (
    <div className="flex justify-center">
      <div className="flex flex-col p-3 md:p-5 xl:px-0 w-full max-w-6xl mx-auto">
        <UpdateBanner serverInfo={serverInfo} />
        {children}
        <ServerInfo serverInfo={serverInfo} />
      </div>
    </div>
  );
};

export default ContentWrapper;
