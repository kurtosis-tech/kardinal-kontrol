import { useRef, useEffect, useState, useCallback } from "react";
import { Grid } from "@chakra-ui/react";
import CytoscapeGraph, { utils } from "@/components/CytoscapeGraph";
import { ElementDefinition } from "cytoscape";
import { useApi } from "@/contexts/ApiContext";
import { useFlowsContext } from "@/contexts/FlowsContext";

const pollingIntervalSeconds = 1;

const Page = () => {
  const [elems, setElems] = useState<ElementDefinition[]>([]);
  const prevResponse = useRef<string>();
  const { getTopology } = useApi();
  const { refetchFlows, flowVisibility } = useFlowsContext();
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  const fetchElems = useCallback(async () => {
    const response = await getTopology();
    const filtered = {
      ...response,
      nodes: response.nodes.map((node) => {
        return {
          ...node,
          versions: node.versions?.filter((version) => {
            const currentVisibility = flowVisibility[version.flowId];
            // un-set visibility is considered visible, only false is hidden
            return currentVisibility == null || currentVisibility === true;
          }),
        };
      }),
    };
    const newElems = utils.normalizeData(filtered);

    // dont update react state if the API response is identical to the previous one
    // This avoids unnecessary re-renders
    if (JSON.stringify(newElems) === prevResponse.current) {
      return;
    }
    prevResponse.current = JSON.stringify(newElems);
    setElems(newElems);
    // re-fetch flows if topology changes
    refetchFlows();
  }, [getTopology, flowVisibility, refetchFlows]);

  const startPolling = useCallback(() => {
    timerRef.current = setInterval(fetchElems, pollingIntervalSeconds * 1000);
  }, [fetchElems]);

  const stopPolling = useCallback(() => {
    clearInterval(timerRef.current!);
    timerRef.current = null;
  }, []);

  useEffect(() => {
    fetchElems();
    startPolling();
    return stopPolling;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [startPolling, stopPolling, fetchElems]);

  return (
    <Grid
      height={"100%"}
      width={"100%"}
      templateColumns={"1fr"}
      templateRows={"1fr"}
      id="traffic-configuration"
    >
      <CytoscapeGraph elements={elems} />
    </Grid>
  );
};

export default Page;
