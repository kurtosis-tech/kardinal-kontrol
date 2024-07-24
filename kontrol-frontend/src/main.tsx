import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import theme from "./theme";
import { ChakraProvider } from "@chakra-ui/react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Fallback from "./components/Fallback";
import Dashboard from "@/pages/Dashboard";
import DataConfiguration from "@/pages/DataConfiguration";
import Flows from "@/pages/Flows";
import MaturityGates from "@/pages/MaturityGates";
import TrafficConfiguration from "@/pages/TrafficConfiguration";
import MockTrafficConfiguration from "@/pages/MockTrafficConfiguration";
import NotFound from "@/pages/NotFound";

import { ErrorBoundary } from "react-error-boundary";

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
        element: <Flows />,
      },
      {
        path: "traffic-configuration",
        children: [
          {
            index: true,
            element: <TrafficConfiguration />,
          },
          {
            path: "dev",
            element: <MockTrafficConfiguration variant="dev" />,
          },
          {
            path: "dev2",
            element: <MockTrafficConfiguration variant="dev2" />,
          },
          {
            path: "main",
            element: <MockTrafficConfiguration variant="main" />,
          },
          {
            path: "all",
            element: <MockTrafficConfiguration variant="all" />,
          },
        ],
      },
      {
        path: "data-configuration",
        element: <DataConfiguration />,
      },
    ],
  },
  {
    path: "*",
    element: <NotFound />,
  },
]);

console.info("Using base URL:", import.meta.env.VITE_API_URL);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ChakraProvider theme={theme} resetCSS>
      <RouterProvider router={router} />
    </ChakraProvider>
  </React.StrictMode>,
);
