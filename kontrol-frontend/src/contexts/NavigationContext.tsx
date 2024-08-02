import { createContext, useContext, useState, ReactNode } from "react";

// Define the shape of your context
interface NavigationContextType {
  // Add your context properties here
  isSidebarCollapsed: boolean;
  setIsSidebarCollapsed: (c: boolean) => void;
}

// Create the context with a default value
const NavigationContext = createContext<NavigationContextType | undefined>(
  undefined,
);

// Create a provider component
interface NavigationContextProviderProps {
  children: ReactNode;
}

export const NavigationContextProvider = ({
  children,
}: NavigationContextProviderProps) => {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  return (
    <NavigationContext.Provider
      value={{
        isSidebarCollapsed,
        setIsSidebarCollapsed,
      }}
    >
      {children}
    </NavigationContext.Provider>
  );
};

// Create a custom hook to use the context
export const useNavigationContext = (): NavigationContextType => {
  const context = useContext(NavigationContext);
  if (context === undefined) {
    throw new Error(
      "useNavigationContext must be used within a NavigationContextProvider",
    );
  }
  return context;
};
