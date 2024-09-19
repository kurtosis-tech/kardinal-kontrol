import Layout from "@/components/Layout";
import Button from "@/components/Button";
import Section from "@/components/Section";
import PageTitle from "@/components/PageTitle";
import Input, { Option } from "@/components/Input";
import { Stack, Flex, Grid } from "@chakra-ui/react";
import StatefulService from "@/components/StatefulService";
import CytoscapeGraph, { utils } from "@/components/CytoscapeGraph";
import { ChangeEvent, useEffect, useState } from "react";
import { ClusterTopology, Node, NodeVersion } from "@/types";
import { useApi } from "@/contexts/ApiContext";
import { useNavigate } from "react-router-dom";

// TODO: This should be imported from the OpenAPI schema once the types there
// are generated correctly
interface TemplateConfig {
  service: Option[];
  /** @description The name to give the template */
  name: string;
  /** @description The description of the template */
  description?: string;
}

// fake preview of a flow built on the base topology
const PREVIEW_DEV_NODE_VERSION: NodeVersion = {
  flowId: "new-dev-flow",
  imageTag: "TBD",
  isBaseline: false,
};
const PREVIEW_BASELINE_NODE_VERSION: NodeVersion = {
  flowId: "baseline",
  imageTag: "TBD",
  isBaseline: true,
};

const Page = () => {
  const navigate = useNavigate();
  const { getTopology, postTemplateCreate } = useApi();

  const [loading, setLoading] = useState<boolean>(true);

  // All service names, derived from the cluster topology
  const [services, setServices] = useState<string[]>([]);

  // Existing cluster topology to build preview on
  const [topology, setTopology] = useState<ClusterTopology>({
    nodes: [],
    edges: [],
  });

  // All input values for the form
  const [formState, setFormState] = useState<Required<TemplateConfig>>({
    name: "",
    service: [],
    description: "",
  });

  // fake preview of a flow built on the base topology
  const previewTopology = utils.normalizeData({
    ...topology,
    nodes: topology.nodes.map((node) => {
      return {
        ...node,
        versions:
          formState.service.find((o) => o.value === node.id) != null
            ? [PREVIEW_BASELINE_NODE_VERSION, PREVIEW_DEV_NODE_VERSION]
            : [PREVIEW_BASELINE_NODE_VERSION],
      };
    }),
  });

  // TODO: Not hardcoding this
  const templateContainsExternalApi =
    formState.service.find((o) => o.value === "jsdelivr-api") != null;
  // TODO: Not hardcoding this
  const templateContainsPostgres =
    formState.service.find((o) => o.value === "postgres") != null;

  useEffect(() => {
    async function fetchData() {
      const topology = await getTopology();
      setTopology(topology);
      setLoading(false);
    }
    fetchData();
  }, [getTopology]);

  // Extract service names for inputs when the topology changes
  useEffect(() => {
    if (loading) return;
    const services = topology.nodes.map((node: Node) => {
      if (node.label == null) {
        throw new Error("Node label is missing in the cluster topology");
      }
      return node.label;
    });
    if (services == null || services.length === 0) {
      throw new Error("No services found in the cluster topology");
    }
    setServices(services);
  }, [topology, loading]);

  const handleInputChange =
    (field: keyof TemplateConfig) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
      setFormState((prevState) => ({
        ...prevState,
        [field]: event.target.value,
      }));
    };

  const handleMultiSelectChange =
    (field: keyof TemplateConfig) => (option: Option | Option[]) => {
      console.log("option", option);
      setFormState((prevState) => ({
        ...prevState,
        [field]: option,
      }));
    };

  const handleCreateFlowTemplate = async () => {
    await postTemplateCreate(formState);
    navigate("../");
  };

  const handleNodeClick = (node: cytoscape.NodeSingular) => {
    const nodeId = node.data("id");
    handleMultiSelectChange("service")([
      ...formState.service,
      {
        label: nodeId,
        value: nodeId,
      },
    ]);
  };

  const formIsValid = formState.name.length > 0 && formState.service.length > 0;

  return (
    <Layout showBanner>
      <PageTitle title="Create new flow configuration">
        Update traffic control and data isolation details below
      </PageTitle>
      <Section title="Preview">
        {!loading && (
          <CytoscapeGraph
            elements={previewTopology}
            onNodeClick={handleNodeClick}
          />
        )}
      </Section>
      <Section title="Flow configuration">
        <Stack w={"100%"} gap={8} as={"fieldset"}>
          <Flex gap={4}>
            <Input.Text
              label="Name"
              id="name"
              value={formState.name}
              placeholder="Enter a name for this flow"
              onChange={handleInputChange("name")}
            />
            <Input.Text
              label="Description"
              id="description"
              value={formState.description}
              placeholder="Enter the description for this flow"
              onChange={handleInputChange("description")}
            />
          </Flex>
          <Grid templateColumns={"1fr"}>
            <Input.MultiSelect
              label="Services"
              options={services.map(
                (service) => ({ label: service, value: service }) as Option,
              )}
              value={formState.service}
              id="service"
              onChange={handleMultiSelectChange("service")}
              width={"100%"}
            />
          </Grid>
        </Stack>
      </Section>
      {templateContainsExternalApi && (
        <Section title="Data configuration">
          <StatefulService type="external_api" />
        </Section>
      )}
      {templateContainsPostgres && (
        <Section offsetTop={9}>
          <StatefulService type="postgres" />
        </Section>
      )}
      <Flex justifyContent={"flex-end"}>
        <Button isDisabled={!formIsValid} onClick={handleCreateFlowTemplate}>
          Create flow template
        </Button>
      </Flex>
    </Layout>
  );
};

export default Page;
