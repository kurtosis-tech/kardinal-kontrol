import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import theme from "./theme";
import { ChakraProvider } from "@chakra-ui/react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Fallback from "./components/Fallback";
import Dashboard from "@/pages/Dashboard";
import DataConfiguration from "@/pages/DataConfiguration";
import FlowsCreate from "@/pages/FlowsCreate";
import FlowsIndex from "@/pages/FlowsIndex";
import MaturityGates from "@/pages/MaturityGates";
import TrafficConfiguration from "@/pages/TrafficConfiguration";
import Profile from "@/pages/Profile";
import NotFound from "@/pages/NotFound";

import { ErrorBoundary } from "react-error-boundary";
import { NavigationContextProvider } from "@/contexts/NavigationContext";
import { ApiContextProvider } from "@/contexts/ApiContext";
import { FlowsContextProvider } from "@/contexts/FlowsContext";

const router = createBrowserRouter([
  {
    path: "/",
    element: <NotFound />,
  },
  {
    path: "/:uuid/",
    element: (
      <ErrorBoundary FallbackComponent={Fallback}>
        <App />
      </ErrorBoundary>
    ),
    children: [
      {
        index: true,
        element: <Dashboard />,
      },
      {
        path: "maturity-gates",
        element: <MaturityGates />,
      },
      {
        path: "flows",
        children: [
          {
            index: true,
            element: <FlowsIndex />,
          },
          {
            path: "create",
            element: <FlowsCreate />,
          },
        ],
      },
      {
        path: "traffic-configuration",
        element: <TrafficConfiguration />,
      },
      {
        path: "data-configuration",
        element: <DataConfiguration />,
      },
      {
        path: "profile",
        element: <Profile />,
      },
    ],
  },
  {
    path: "*",
    element: <NotFound />,
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ChakraProvider theme={theme} resetCSS>
      <ApiContextProvider>
        <NavigationContextProvider>
          <FlowsContextProvider>
            <RouterProvider router={router} />
          </FlowsContextProvider>
        </NavigationContextProvider>
      </ApiContextProvider>
    </ChakraProvider>
  </React.StrictMode>,
);
