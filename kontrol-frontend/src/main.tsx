import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import theme from "./theme";
import { ChakraProvider } from "@chakra-ui/react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Fallback from "./components/Fallback";
import Dashboard from "@/pages/Dashboard";
import DataConfiguration from "@/pages/DataConfiguration";
import Flows from "@/pages/Flows";
import MaturityGates from "@/pages/MaturityGates";
import TrafficConfiguration from "@/pages/TrafficConfiguration";
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
        element: <TrafficConfiguration />,
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

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ChakraProvider theme={theme} resetCSS>
      <RouterProvider router={router} />
    </ChakraProvider>
  </React.StrictMode>,
);
