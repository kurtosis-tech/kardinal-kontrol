import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import theme from "./theme";
import { ChakraProvider } from "@chakra-ui/react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Create from "./components/Create";
import Review from "./components/Review";
import Manage from "./components/Manage";

import { ErrorBoundary } from "react-error-boundary";

function Fallback({ error, resetErrorBoundary }) {
  // Call resetErrorBoundary() to reset the error boundary and retry the render.
  console.debug({ error });
  return (
    <div role="alert" style={{ padding: 24 }}>
      <p>Something went wrong:</p>
      <pre style={{ color: "red" }}>{error.message}</pre>
      <pre>{error.stack}</pre>
    </div>
  );
}

const router = createBrowserRouter([
  {
    path: "/",
    element: (
      <ErrorBoundary
        FallbackComponent={Fallback}
        onReset={(details) => {
          // Reset the state of your app so the error doesn't happen again
        }}
      >
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
