import CytoscapeGraph from "@/components/TrafficConfiguration/CytoscapeGraph";
import * as mockData from "@/components/TrafficConfiguration/mocks";

const Page = ({ variant }: { variant: "dev" | "dev2" | "main" | "all" }) => {
  return (
    <CytoscapeGraph
      elements={
        variant === "dev"
          ? mockData.devFlow
          : variant === "dev2"
            ? mockData.devFlow2
            : mockData.noDevFlows
      }
    />
  );
};

export default Page;
