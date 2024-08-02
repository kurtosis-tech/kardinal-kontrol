import CytoscapeGraph from "@/components/TrafficConfiguration/CytoscapeGraph";
import { mockResponse } from "@/components/TrafficConfiguration/mocks";

const Page = () => {
  return <CytoscapeGraph elements={mockResponse} />;
};

export default Page;
