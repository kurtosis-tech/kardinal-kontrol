import CytoscapeComponent from "react-cytoscapejs";
import { useCallback, useEffect, useRef, useState } from "react";
import cytoscape from "cytoscape";
import dagre from "cytoscape-dagre";
import stylesheet, { trafficNodeSelector } from "./stylesheet";
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

const BASE_ANIMATION_DURATION = 2000;

const CytoscapeGraph = ({ elements }: Props) => {
  // keep a ref to the cy instance
  const cy = useRef<cytoscape.Core>();
  const [animationsEnabled, setAnimationsEnabled] = useState(true);
  // remove all animated traffic nodes
  const removeAnimatedTrafficNodes = useCallback(() => {
    if (!cy.current) return;

    const animatedNodes = cy.current.nodes(trafficNodeSelector);
    animatedNodes.remove();
  }, [cy]);

  const handleCy = useCallback(
    (cyInstance: cytoscape.Core) => {
      cy.current = cyInstance;
      // stop animations when the user is dragging nodes around
      cy.current.on("tapstart", function () {
        setAnimationsEnabled(false);
      });
      // re-start animations when the user is done with interactions
      cy.current.on("tapend", function () {
        setTimeout(() => setAnimationsEnabled(true), 0);
      });
    },
    [setAnimationsEnabled],
  );

  // when animation is disabled, remove all animated traffic nodes
  useEffect(() => {
    if (animationsEnabled) return;
    removeAnimatedTrafficNodes();
  }, [removeAnimatedTrafficNodes, animationsEnabled, cy]);

  // when graph changes, remove all animated traffic nodes
  useEffect(() => {
    removeAnimatedTrafficNodes();
    cy?.current?.layout(layout).run();
  }, [elements, removeAnimatedTrafficNodes, cy]);

  const animateEdges = useCallback(() => {
    if (!cy.current) {
      throw new Error("animateEdges: cytoscape instance is not initialized");
    }

    const edges = cy.current.edges();

    const animationPromises = edges.map(async (edge) => {
      if (!cy.current) {
        throw new Error(
          "animationPromises: cytoscape instance is not initialized",
        );
      }

      const trafficNodeId = `traffic:${edge.source().id()}:${edge.target().id()}`;
      // if the traffic node already exists, skip it. this should only happen in dev hot reload
      if (cy.current.getElementById(trafficNodeId).length > 0) {
        return;
      }
      const trafficNode = cy.current.add({
        group: "nodes",
        data: {
          id: trafficNodeId,
        },
        position: { ...edge.source().position() },
      });
      trafficNode.ungrabify(); // prevent mouse from moving this node
      trafficNode.position({ ...edge.source().position() });
      // generate a random number between 2000 and 2400
      const duration =
        Math.floor(Math.random() * 400) + BASE_ANIMATION_DURATION;

      // @ts-expect-error poor typing upstream
      const animation = trafficNode.animation({
        position: { ...edge.target().position() },
        style: {},
        easing: "ease-in-out-cubic",
        duration,
      });

      animation
        .play()
        .promise()
        .then(() => {
          // each traffic node is removed after the animation is done
          trafficNode.remove();
        });
    });
    return Promise.all(animationPromises);
  }, []);

  // trigger the animation every 2.5 seconds
  useInterval(
    () => {
      if (!cy) return;
      // this is technically a promise / async but we dont really need to await
      // this right now, but maybe later we might want to do something after
      // each cycle
      animateEdges();
    },
    BASE_ANIMATION_DURATION + 500,
    { immediate: true, paused: !animationsEnabled },
  );

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
