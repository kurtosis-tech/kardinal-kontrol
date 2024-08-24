import { createContext, useContext, useState, ReactNode } from "react";

interface NavigationContextType {
  isSidebarCollapsed: boolean;
  setIsSidebarCollapsed: (c: boolean) => void;
  isBannerVisible: boolean;
  setIsBannerVisible: (c: boolean) => void;
}

// Create the context with a default value
const NavigationContext = createContext<NavigationContextType>({
  isSidebarCollapsed: false,
  setIsSidebarCollapsed: () => {},
  isBannerVisible: true,
  setIsBannerVisible: () => {},
});

// Create a provider component
interface NavigationContextProviderProps {
  children: ReactNode;
}

export const NavigationContextProvider = ({
  children,
}: NavigationContextProviderProps) => {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [isBannerVisible, setIsBannerVisible] = useState(true);

  return (
    <NavigationContext.Provider
      value={{
        isSidebarCollapsed,
        setIsSidebarCollapsed,
        isBannerVisible,
        setIsBannerVisible,
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
