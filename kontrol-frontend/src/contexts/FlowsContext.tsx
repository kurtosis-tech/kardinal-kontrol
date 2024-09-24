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
  activeFlowId: string | null;
  setActiveFlowId: (flowId: string | null) => void;
}

// Create the context with a default value
const FlowsContext = createContext<FlowsContextType>({
  flows: [],
  refetchFlows: async () => [],
  activeFlowId: null,
  setActiveFlowId: () => {},
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

  const [activeFlowId, setActiveFlowId] = useState<string | null>(null);

  useEffect(() => {
    getFlows().then(setFlows);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
        activeFlowId,
        setActiveFlowId,
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
