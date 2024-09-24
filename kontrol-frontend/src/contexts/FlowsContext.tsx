import {
  createContext,
  useContext,
  useState,
  ReactNode,
  useEffect,
} from "react";
import { useApi } from "@/contexts/ApiContext";
import { Flow } from "@/types";

interface FlowsContextType {
  flows: Flow[];
  refetchFlows: () => Promise<Flow[]>;
  flowVisibility: Record<string, boolean>;
  setFlowVisibility: (flowId: string, visible: boolean) => void;
}

// Create the context with a default value
const FlowsContext = createContext<FlowsContextType>({
  flows: [],
  refetchFlows: async () => [],
  flowVisibility: {},
  setFlowVisibility: () => {},
});

// Create a provider component
interface FlowsContextProviderProps {
  children: ReactNode;
}

export const FlowsContextProvider = ({
  children,
}: FlowsContextProviderProps) => {
  const { getFlows } = useApi();
  const [flows, setFlows] = useState<Flow[]>([]);

  const [flowVisibility, setFlowVisibility] = useState<Record<string, boolean>>(
    {},
  );

  useEffect(() => {
    setFlowVisibility((prevState) =>
      flows.reduce(
        (acc, flow) => ({
          ...acc,
          [flow["flow-id"]]: prevState[flow["flow-id"]] ?? true,
        }),
        {},
      ),
    );
  }, [flows]);

  useEffect(() => {
    getFlows().then(setFlows);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    console.log("flowVisibility", flowVisibility);
  }, [flowVisibility]);

  const refetchFlows = async () => {
    const newFlows = await getFlows();
    setFlows(newFlows);
    return newFlows;
  };

  return (
    <FlowsContext.Provider
      value={{
        flows,
        refetchFlows,
        flowVisibility,
        setFlowVisibility: (flowId: string, visible: boolean) => {
          setFlowVisibility({ ...flowVisibility, [flowId]: visible });
        },
      }}
    >
      {children}
    </FlowsContext.Provider>
  );
};

// Create a custom hook to use the context
// eslint-disable-next-line react-refresh/only-export-components
export const useFlowsContext = (): FlowsContextType => {
  const context = useContext(FlowsContext);
  if (context === undefined) {
    throw new Error(
      "useFlowsContext must be used within a FlowsContextProvider",
    );
  }
  return context;
};
