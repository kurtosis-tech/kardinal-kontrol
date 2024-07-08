import { useCallback, useEffect, useState } from "react";
import { Flex } from "@chakra-ui/react";
import CytoscapeComponent from "react-cytoscapejs";
import cytoscape from "cytoscape";
import dagre from "cytoscape-dagre";
import stylesheet from "./stylesheet";
import { useParams } from "react-router-dom";
import { paths } from "cli-kontrol-api/api/typescript/client/types";
import createClient from "openapi-fetch";

cytoscape.use(dagre);

const layout = {
  name: "dagre",
  nodeSep: 50,
  nodeDimensionsIncludeLabels: true,
  rankDir: "LR",
  align: "UL",
};

const pullingIntervalSeconds = 1;

const client = createClient<paths>({ baseUrl: import.meta.env.VITE_API_URL });

const TrafficConfiguration = () => {
  const [elems, setElems] = useState<cytoscape.ElementDefinition[]>([]);
  const { uuid } = useParams<{ uuid: string }>();
  if (!uuid) {
    console.error("UUID is undefined");
    // TODO: Handle error better
    return null;
  }

  useEffect(() => {
    const fetchElems = async () => {
      try {
        const response = await client.GET("/tenant/{uuid}/topology", {
          params: { path: { uuid } },
        });
        setElems(
          CytoscapeComponent.normalizeElements({
            nodes: response.data!.nodes.map((node) => ({
              data: node,
            })),
            edges: response.data!.edges.map((edge) => ({
              data: edge,
            })),
          }),
        );
      } catch (error) {
        console.error("Failed to fetch elements:", error);
      }
    };

    // Continuously fetch elements
    const intervalId = setInterval(fetchElems, pullingIntervalSeconds * 1000);
    fetchElems();
    return () => clearInterval(intervalId);
  }, [uuid]);

  const handleCy = useCallback((cy: cytoscape.Core) => {
    const edges = cy.edges();
    const allNodeIds = cy.nodes().map((n) => n.id());
    const animateEdge = (edge: cytoscape.EdgeSingular) => {
      const trafficNodeId = `traffic:${edge.source().id()}:${edge.target().id()}`;
      // we shouldnt need to do this but destroying nodes on unmount is not
      // working as expected, so this can break on hot reloads
      if (allNodeIds.includes(trafficNodeId)) {
        return;
      }

      const trafficNode = cy.add({
        group: "nodes",
        data: {
          id: trafficNodeId,
        },
        position: { ...edge.source().position() },
      });
      trafficNode.ungrabify(); // prevent mouse from moving this node

      const animationLoop = () => {
        trafficNode.position({ ...edge.source().position() });
        // generate a random number between 2000 and 2500
        const duration = Math.floor(Math.random() * 500) + 2000;

        // @ts-expect-error poor typing upstream
        const animation = trafficNode.animation({
          position: edge.target().position(),
          style: {},
          easing: "ease-in-out-cubic",
          duration,
        });

        animation.play();
        setTimeout(animationLoop, duration);
      };
      animationLoop();
    };

    cy.layout(layout).run();
    edges.forEach(animateEdge);
  }, []);
  return (
    <Flex
      alignItems={"center"}
      justifyContent={"center"}
      height={"100%"}
      flexDir={"column"}
    >
      <CytoscapeComponent
        elements={elems}
        style={{ width: "100%", height: "100%" }}
        layout={layout}
        // @ts-expect-error cytoscape
        stylesheet={stylesheet}
        cy={handleCy}
      />
    </Flex>
  );
};

export default TrafficConfiguration;