/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */


export interface paths {
  "/flow/create": {
    post: {
      /** @description Create a dev flow */
      requestBody: {
        content: {
          "application/json": components["schemas"]["DevFlowSpec"];
        };
      };
      responses: {
        /** @description Dev flow creation status */
        200: {
          content: {
            "application/json": string;
          };
        };
      };
    };
  };
  "/flow/delete": {
    post: {
      /** @description Delete dev flow (revert back to prod only) */
      requestBody: {
        content: {
          "application/json": components["schemas"]["ProdFlowSpec"];
        };
      };
      responses: {
        /** @description Dev flow creation status */
        200: {
          content: {
            "application/json": string;
          };
        };
      };
    };
  };
  "/deploy": {
    post: {
      /** @description Deploy a prod only cluster */
      requestBody: {
        content: {
          "application/json": components["schemas"]["ProdFlowSpec"];
        };
      };
      responses: {
        /** @description Dev flow creation status */
        200: {
          content: {
            "application/json": string;
          };
        };
      };
    };
  };
  "/topology": {
    get: {
      parameters: {
        query?: {
          /** @description The namespace for which to retrieve the topology */
          namespace?: string;
        };
      };
      responses: {
        /** @description Topology information */
        200: {
          content: {
            "application/json": components["schemas"]["Topology"];
          };
        };
      };
    };
  };
}

export type webhooks = Record<string, never>;

export interface components {
  schemas: {
    ProdFlowSpec: {
      "docker-compose"?: unknown[];
    };
    DevFlowSpec: {
      /** @example backend-a:latest */
      "image-locator"?: string;
      /** @example backend-service-a */
      "service-name"?: string;
      "docker-compose"?: unknown[];
    };
    Topology: {
      graph?: components["schemas"]["Graph"];
    };
    Graph: {
      nodes?: components["schemas"]["Node"][];
    };
    Node: {
      /** @example backend-service-a */
      serviceName?: string;
      /** @example 1.0.0 */
      serviceVersion?: string;
      /** @example node-1 */
      id?: string;
      /**
       * @example [
       *   "node-2",
       *   "node-3"
       * ]
       */
      talks_to?: string[];
    };
  };
  responses: never;
  parameters: never;
  requestBodies: never;
  headers: never;
  pathItems: never;
}

export type $defs = Record<string, never>;

export type external = Record<string, never>;

export type operations = Record<string, never>;
