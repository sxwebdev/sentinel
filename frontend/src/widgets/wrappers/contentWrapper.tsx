import {type PropsWithChildren} from "react";

const ContentWrapper = ({children}: PropsWithChildren) => {
  return (
    <div className="flex justify-center">
      <div className="flex flex-col p-4 w-full max-w-7xl mx-auto">
        {children}
      </div>
    </div>
  );
};

export default ContentWrapper;
