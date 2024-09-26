import { paths, components } from "cli-kontrol-api/api/typescript/client/types";
import { matchPath } from "react-router-dom";
import createClient from "openapi-fetch";
import type { HttpMethod } from "openapi-typescript-helpers";
import { Flow, Template } from "@/types";

const client = createClient<paths>({ baseUrl: import.meta.env.VITE_API_URL });

import {
  createContext,
  useContext,
  useState,
  useCallback,
  PropsWithChildren,
} from "react";

// infer the request body type from the OpenAPI schema
export type RequestBody<
  T extends keyof paths,
  M extends keyof paths[T] & HttpMethod,
> = paths[T][M] extends {
  requestBody: { content: { "application/json": infer R } };
}
  ? R
  : never;

export interface ApiContextType {
  deleteFlow: (flowId: string) => Promise<Flow[]>;
  deleteTemplate: (templateName: string) => Promise<void>;
  error: string | null;
  getFlows: () => Promise<Flow[]>;
  getTemplates: () => Promise<void>;
  getTopology: () => Promise<components["schemas"]["ClusterTopology"]>;
  loading: boolean;
  postFlowCreate: (
    b: RequestBody<"/tenant/{uuid}/flow/create", "post">,
  ) => Promise<Flow>;
  postTemplateCreate: (
    b: RequestBody<"/tenant/{uuid}/templates/create", "post">,
  ) => Promise<void>;
  templates: Template[];
}

const defaultContextValue: ApiContextType = {
  deleteFlow: async () => [],
  deleteTemplate: async () => { },
  error: null,
  getFlows: async () => [],
  getTemplates: async () => { },
  getTopology: async () => {
    return { nodes: [], edges: [] };
  },
  loading: false,
  postFlowCreate: async () => ({
    "flow-id": "",
    "access-entry": [],
    isBaseline: false,
  }),
  postTemplateCreate: async () => { },
  templates: [],
};

const ApiContext = createContext<ApiContextType>(defaultContextValue);

export const ApiContextProvider = ({ children }: PropsWithChildren) => {
  const match = matchPath(
    {
      path: "/:uuid/*",
    },
    location.pathname,
  );
  const uuid = match?.params.uuid;

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [templates, setTemplates] = useState<Template[]>([]);

  // boilerplate loading state, error handling for any API call
  const handleApiCall = useCallback(async function <T>(
    pendingRequest: Promise<{ data?: T }>, // Api fetch promise
  ): Promise<T> {
    setLoading(true);
    try {
      const response = await pendingRequest;
      if (response.data == null) {
        throw new Error("API response data is null");
      }
      return response.data;
    } catch (error: unknown) {
      console.error("Failed to fetch route:", error);
      setError((error as Error).message);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  // POST "/tenant/{uuid}/flow/create"
  const postFlowCreate = useCallback(
    async (body: RequestBody<"/tenant/{uuid}/flow/create", "post">) => {
      if (uuid == null) throw new Error("Invalid or missing tenant UUID");
      const flow = await handleApiCall(
        client.POST("/tenant/{uuid}/flow/create", {
          params: { path: { uuid } },
          body,
        }),
      );
      return flow;
    },
    [uuid, handleApiCall],
  );

  // GET "/tenant/{uuid}/flows"
  const getFlows = useCallback(async () => {
    if (uuid == null) throw new Error("Invalid or missing tenant UUID");
    const flows = await handleApiCall(
      client.GET("/tenant/{uuid}/flows", {
        params: { path: { uuid } },
      }),
    );
    return flows;
  }, [uuid, handleApiCall]);

  // DELETE "/tenant/{uuid}/flow/{flow-id}"
  const deleteFlow = useCallback(
    async (flowId: string) => {
      if (uuid == null) throw new Error("Invalid or missing tenant UUID");
      const flows = await handleApiCall(
        client.DELETE("/tenant/{uuid}/flow/{flow-id}", {
          params: { path: { uuid, "flow-id": flowId } },
        }),
      );
      return flows;
    },
    [uuid, handleApiCall],
  );

  // GET "/tenant/{uuid}/topology"
  const getTopology = useCallback(async () => {
    if (uuid == null) {
      throw new Error("Invalid or missing tenant UUID");
    }
    const topology = await handleApiCall(
      client.GET("/tenant/{uuid}/topology", {
        params: { path: { uuid } },
      }),
    );
    return topology;
  }, [uuid, handleApiCall]);

  // POST "/tenant/{uuid}/templates/create"
  const postTemplateCreate = useCallback(
    async (body: RequestBody<"/tenant/{uuid}/templates/create", "post">) => {
      if (uuid == null) throw new Error("Invalid or missing tenant UUID");
      const template = await handleApiCall(
        client.POST("/tenant/{uuid}/templates/create", {
          params: { path: { uuid } },
          body,
        }),
      );
      setTemplates((state) => [...state, template]);
    },
    [uuid, handleApiCall],
  );

  // GET "/tenant/{uuid}/templates"
  const getTemplates = useCallback(async () => {
    if (uuid == null) throw new Error("Invalid or missing tenant UUID");
    const templates = await handleApiCall(
      client.GET("/tenant/{uuid}/templates", {
        params: { path: { uuid } },
      }),
    );
    setTemplates(templates);
  }, [uuid, handleApiCall]);

  // DELETE "/tenant/{uuid}/templates/{template-name}"
  const deleteTemplate = useCallback(
    async (templateName: string) => {
      if (uuid == null) throw new Error("Invalid or missing tenant UUID");
      const templates = await handleApiCall(
        client.DELETE("/tenant/{uuid}/templates/{template-name}", {
          params: { path: { uuid, "template-name": templateName } },
        }),
      );
      setTemplates(templates);
    },
    [uuid, handleApiCall],
  );

  return (
    <ApiContext.Provider
      value={{
        deleteFlow,
        deleteTemplate,
        error,
        getFlows,
        getTemplates,
        getTopology,
        loading,
        postFlowCreate,
        postTemplateCreate,
        templates,
      }}
    >
      {children}
    </ApiContext.Provider>
  );
};

// eslint-disable-next-line react-refresh/only-export-components
export const useApi = (): ApiContextType => {
  const context = useContext(ApiContext);
  if (!context) {
    throw new Error("useApi must be used within an ApiContextProvider");
  }
  return context;
};
