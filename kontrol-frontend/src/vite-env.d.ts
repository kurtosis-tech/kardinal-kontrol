/// <reference types="vite/client" />
/// <reference types="vite-plugin-virtual-plain-text/virtual-assets" />
import type { ComputePositionConfig } from "@floating-ui/dom";

declare module "cytoscape-popper" {
  interface PopperOptions extends ComputePositionConfig {}
  interface PopperInstance {
    update(): void;
  }
}
