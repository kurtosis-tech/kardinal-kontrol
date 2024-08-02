import { useRef, useEffect, useState } from "react";
import { Flex } from "@chakra-ui/react";
import { useParams } from "react-router-dom";
import { paths } from "cli-kontrol-api/api/typescript/client/types";
import createClient from "openapi-fetch";
import CytoscapeGraph from "./CytoscapeGraph";
import { normalizeData } from "./utils";
import { ElementDefinition } from "cytoscape";

const pollingIntervalSeconds = 1;

const client = createClient<paths>({ baseUrl: import.meta.env.VITE_API_URL });

const TrafficConfiguration = () => {
  const [elems, setElems] = useState<ElementDefinition[]>([]);
  const prevResponse = useRef<string>();

  const { uuid } = useParams<{ uuid: string }>();

  useEffect(() => {
    if (!uuid) {
      throw new Error("UUID is undefined. Make sure to include it in the URL!");
    }

    const fetchElems = async () => {
      try {
        const response = await client.GET("/tenant/{uuid}/topology", {
          params: { path: { uuid } },
        });
        if (response.data == null) {
          throw new Error(
            "Topology response data is null, check network connection or API status",
          );
        }
        const newElems = normalizeData(response.data);

        // dont update react state if the API response is identical to the previous one
        // This avoids unnecessary re-renders
        if (JSON.stringify(newElems) === prevResponse.current) {
          return;
        }
        prevResponse.current = JSON.stringify(newElems);
        setElems(newElems);
      } catch (error) {
        console.error("Failed to fetch elements:", error);
      }
    };

    // Continuously fetch elements
    const intervalId = setInterval(fetchElems, pollingIntervalSeconds * 1000);
    fetchElems();
    return () => clearInterval(intervalId);
  }, [uuid]);

  return (
    <Flex
      alignItems={"center"}
      justifyContent={"center"}
      height={"100%"}
      flexDir={"column"}
    >
      <CytoscapeGraph elements={elems} />
    </Flex>
  );
};

export default TrafficConfiguration;
