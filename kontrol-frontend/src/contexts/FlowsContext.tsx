import { createContext, useContext, useState, ReactNode } from "react";

type AccessMode = "empty-with-seed" | "empty" | "snapshot" | "proxy";

interface FlowConfiguration {
  service: string;
  image: string;
  accessMode: AccessMode;
  additionalFields?: Record<string, string>;
}

interface Flow {
  id: string;
  name: string;
  configurations: FlowConfiguration[];
}

// Define the shape of your context
interface FlowsContextType {
  createExampleFlow: () => void;
  deleteExampleFlow: () => void;
  flows: Flow[];
}

// Create the context with a default value
const FlowsContext = createContext<FlowsContextType | undefined>(undefined);

// Create a provider component
interface FlowsContextProviderProps {
  children: ReactNode;
}

const exampleFlow: Flow = {
  id: crypto.randomUUID(),
  name: "Early development flow",
  configurations: [
    {
      service: "awesome-kardinal-example",
      image: "kurtosistech/awesome-kardinal-example",
      accessMode: "empty-with-seed",
      additionalFields: {
        seedScriptUrl:
          "https://github.com/kardinal/awesome-kardinal-example/tree/main/seed.sql",
      },
    },
    {
      service: "stripe",
      image: "kurtosistech/stripe-proxy",
      accessMode: "proxy",
    },
    {
      service: "postgres",
      image: "kurtosistech/postgres-proxy",
      accessMode: "proxy",
    },
  ],
};

export const FlowsContextProvider = ({
  children,
}: FlowsContextProviderProps) => {
  const [flows, setFlows] = useState<Flow[]>([]);

  const createExampleFlow = () => {
    setFlows((prevFlows) => [...prevFlows, exampleFlow]);
  };

  const deleteExampleFlow = () => {
    setFlows([]);
  };

  return (
    <FlowsContext.Provider
      value={{
        flows,
        createExampleFlow,
        deleteExampleFlow,
      }}
    >
      {children}
    </FlowsContext.Provider>
  );
};

// Create a custom hook to use the context
export const useFlowsContext = (): FlowsContextType => {
  const context = useContext(FlowsContext);
  if (context === undefined) {
    throw new Error(
      "useFlowsContext must be used within a FlowsContextProvider",
    );
  }
  return context;
};
