import { useCallback } from "react";
import { Flex } from "@chakra-ui/react";
import CytoscapeComponent from "react-cytoscapejs";
import cytoscape from "cytoscape";
// import sbgnStylesheet from "cytoscape-sbgn-stylesheet";
import dagre from "cytoscape-dagre";
import stylesheet from "./stylesheet";
import data from "./data";
import { paths } from "cli-kontrol-api/api/typescript/client/types"


cytoscape.use(dagre);

const layout = {
  name: "dagre",
  nodeSep: 50,
  nodeDimensionsIncludeLabels: true,
  rankDir: "LR",
  align: "UL",
};

const elements = [...data.nodes, ...data.edges].map((element) => ({
  data: element,
}));

const TrafficConfiguration = () => {
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
        elements={elements}
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
