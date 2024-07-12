import CytoscapeComponent from "react-cytoscapejs";
import { useCallback, useEffect, useState } from "react";
import cytoscape from "cytoscape";
import dagre from "cytoscape-dagre";
import stylesheet from "./stylesheet";
import { useInterval } from "@react-hooks-library/core";

const layout = {
  name: "dagre",
  nodeSep: 50,
  nodeDimensionsIncludeLabels: true,
  rankDir: "LR",
  align: "UL",
};

interface Props {
  elements: cytoscape.ElementDefinition[];
}

cytoscape.use(dagre);

const CytoscapeGraph = ({ elements }: Props) => {
  // keep a ref to the cy instance
  const [cy, setCy] = useState<cytoscape.Core>();

  const handleCy = useCallback((cy: cytoscape.Core) => {
    setCy(cy);
    cy.layout(layout).run();
  }, []);

  const animateEdges = useCallback(
    async (edges: cytoscape.EdgeCollection) => {
      if (!cy) return;
      const animationPromises = edges.map(async (edge) => {
        const trafficNodeId = `traffic:${edge.source().id()}:${edge.target().id()}`;
        // if the traffic node already exists, skip it. this should only happen in dev hot reload
        if (cy.getElementById(trafficNodeId).length > 0) {
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
        trafficNode.position({ ...edge.source().position() });
        // generate a random number between 2000 and 2400
        const duration = Math.floor(Math.random() * 400) + 2000;

        // @ts-expect-error poor typing upstream
        const animation = trafficNode.animation({
          position: { ...edge.target().position() },
          style: {},
          easing: "ease-in-out-cubic",
          duration,
        });

        await animation.play().promise();

        // each traffic node is removed after the animation is done
        trafficNode.remove();
      });
      return Promise.all(animationPromises);
    },
    [cy],
  );

  useInterval(async () => {
    if (!cy) return;
    // dont really need to await this right now, but maybe later we might want to do something after each cycle
    await animateEdges(cy.edges());
  }, 2500);

  useEffect(() => {
    if (!cy) return;
    animateEdges(cy.edges()); // initial run, the rest is handled by the interval
    console.log("Animating edges because of new elements.");
  }, [cy, elements, animateEdges]);

  return (
    <CytoscapeComponent
      elements={elements}
      style={{ width: "100%", height: "100%", minHeight: "400px" }}
      layout={layout}
      // @ts-expect-error cytoscape
      stylesheet={stylesheet}
      cy={handleCy}
    />
  );
};
export default CytoscapeGraph;
