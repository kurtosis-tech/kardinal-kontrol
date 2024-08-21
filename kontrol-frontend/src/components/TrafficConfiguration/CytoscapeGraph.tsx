import CytoscapeComponent from "react-cytoscapejs";
import { useCallback, useEffect, useRef, useState } from "react";
import cytoscape from "cytoscape";
import stylesheet, { trafficNodeSelector } from "./stylesheet";
import { useInterval } from "@react-hooks-library/core";
import dagrePlugin, { dagreLayout } from "./plugins/dagre";
import tippyPlugin, { createTooltip, TooltipInstance } from "./plugins/tippy";

// register plugins with cytoscape
cytoscape.use(dagrePlugin);
cytoscape.use(tippyPlugin);

// base animation timing, which we add a random amount of 0-500ms to
const BASE_ANIMATION_DURATION = 2000;
// enable animations. false is for debug or when using taxi edges (animations
// only really make sense with straight edges)
const INIT_ANIMATIONS_ENABLED = false;

interface Props {
  elements: cytoscape.ElementDefinition[];
  layout?: cytoscape.LayoutOptions;
}

const CytoscapeGraph = ({ elements, layout = dagreLayout }: Props) => {
  // keep a ref to the cy instance. using state will cause infinite re-renders
  const cy = useRef<cytoscape.Core>();
  const tooltip = useRef<TooltipInstance | null>(null);
  const [animationsEnabled, setAnimationsEnabled] = useState(
    INIT_ANIMATIONS_ENABLED,
  );

  // remove all animated traffic nodes
  const removeAnimatedTrafficNodes = useCallback(() => {
    if (!cy.current) return;

    const animatedNodes = cy.current.nodes(trafficNodeSelector);
    animatedNodes.remove();
  }, [cy]);

  // handle cytoscape instance callback when cy is initialized
  const handleCy = useCallback(
    (cyInstance: cytoscape.Core) => {
      // set mutable cy instance
      cy.current = cyInstance;
      // add event listeners to create tooltips on hover
      cy.current.on("mouseover", function (ele: cytoscape.EventObject) {
        if (tooltip.current != null) {
          tooltip.current.destroy();
        }
        tooltip.current = createTooltip(ele.target);
      });
      cy.current.on("mouseout", function () {
        if (tooltip.current != null) {
          tooltip.current.destroy();
        }
      });

      // stop animations when the user is dragging nodes around
      cy.current.on("tapstart", function () {
        setAnimationsEnabled(false);
      });
      // remove tooltip when a node is moved
      cy.current.on("drag pan zoom", function () {
        cy.current?.nodes().removeClass("selected");
        if (tooltip.current == null) return;
        tooltip.current?.destroy();
        tooltip.current = null;
      });
      // re-start animations when the user is done with interactions
      cy.current.on("tapend", function () {
        setTimeout(() => setAnimationsEnabled(INIT_ANIMATIONS_ENABLED), 0);
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
  }, [elements, removeAnimatedTrafficNodes, cy, layout]);

  const animateEdges = useCallback(() => {
    if (!cy.current) {
      throw new Error("animateEdges: cytoscape instance is not initialized");
    }

    const edges = cy.current.edges();

    // create a animation for each edge
    const animationPromises = edges.map(async (edge) => {
      if (!cy.current) {
        throw new Error(
          "animationPromises: cytoscape instance is not initialized",
        );
      }

      // construct a unique id for the traffic node that will be created and animated
      const trafficNodeId = `traffic:${edge.source().id()}:${edge.target().id()}`;
      // if the traffic node already exists, skip it. this should only happen
      // in dev live reload since the component may not fully re-mount
      if (cy.current.getElementById(trafficNodeId).length > 0) {
        return;
      }
      // create the animated traffic node element
      const trafficNode = cy.current.add({
        group: "nodes",
        data: {
          id: trafficNodeId,
        },
        position: { ...edge.source().position() },
      });
      // prevent mouse from moving the animated traffic node
      trafficNode.ungrabify();
      // set the initial position of the traffic node to the edge source node
      trafficNode.position({ ...edge.source().position() });
      // generate a random number between 2000 and 2400
      const duration =
        Math.floor(Math.random() * 400) + BASE_ANIMATION_DURATION;

      // animate the traffic node to the edge target node
      // @ts-expect-error poor typing upstream
      const animation = trafficNode.animation({
        position: { ...edge.target().position() },
        style: {},
        easing: "ease-in-out-cubic",
        duration,
      });

      // play the animation
      animation
        .play()
        .promise()
        .then(() => {
          // remove each traffic node once its animation has completed
          trafficNode.remove();
        });
    });
    return Promise.all(animationPromises);
  }, []);

  // trigger the animation every 2.5 seconds
  useInterval(
    () => {
      if (!cy) return;
      if (!animationsEnabled) return;
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
      id="cytoscape-graph"
      elements={elements}
      style={{
        width: "100%",
        height: "100%",
        minHeight: "267px",
        display: "flex",
      }}
      layout={layout}
      // @ts-expect-error cytoscape types are not great
      stylesheet={stylesheet}
      cy={handleCy}
    />
  );
};
export default CytoscapeGraph;
