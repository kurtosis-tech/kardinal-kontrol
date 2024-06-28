import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import theme from "./theme";
import { ChakraProvider } from "@chakra-ui/react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Create from "./components/Create";
import Review from "./components/Review/index.tsx";
import Manage from "./components/Manage";
import Fallback from "./components/Fallback";

import { ErrorBoundary } from "react-error-boundary";

const router = createBrowserRouter([
  {
    path: "/",
    element: (
      <ErrorBoundary FallbackComponent={Fallback}>
        <App />
      </ErrorBoundary>
    ),
    children: [
      {
        index: true,
        element: <Create />,
      },
      {
        path: "review",
        element: <Review />,
      },
      {
        path: "manage",
        element: <Manage />,
      },
    ],
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ChakraProvider theme={theme} resetCSS>
      <RouterProvider router={router} />
    </ChakraProvider>
  </React.StrictMode>,
);
