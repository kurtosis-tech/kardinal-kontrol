import { useRef, useEffect, useState } from "react";
import { Grid } from "@chakra-ui/react";
import CytoscapeGraph from "./CytoscapeGraph";
import { normalizeData } from "./utils";
import { ElementDefinition } from "cytoscape";
import { useApi } from "@/contexts/ApiContext";

const pollingIntervalSeconds = 1;

const TrafficConfiguration = () => {
  const [elems, setElems] = useState<ElementDefinition[]>([]);
  const prevResponse = useRef<string>();
  const { getTopology } = useApi();

  useEffect(() => {
    const fetchElems = async () => {
      const response = await getTopology();
      const newElems = normalizeData(response);

      // dont update react state if the API response is identical to the previous one
      // This avoids unnecessary re-renders
      if (JSON.stringify(newElems) === prevResponse.current) {
        return;
      }
      prevResponse.current = JSON.stringify(newElems);
      setElems(newElems);
    };

    // Continuously fetch elements
    const intervalId = setInterval(fetchElems, pollingIntervalSeconds * 1000);
    fetchElems();
    return () => clearInterval(intervalId);
  }, [getTopology]);

  return (
    <Grid
      height={"100%"}
      width={"100%"}
      maxH={"520px"}
      templateColumns={"1fr"}
      templateRows={"1fr"}
      id="traffic-configuration"
    >
      <CytoscapeGraph elements={elems} />
    </Grid>
  );
};

export default TrafficConfiguration;
