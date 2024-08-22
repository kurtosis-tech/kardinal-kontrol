import Layout from "@/components/Layout";
import Section from "@/components/Section";
import PageTitle from "@/components/PageTitle";
import Footer from "@/components/Footer";
import Input from "@/components/Input";
import { Stack, Flex } from "@chakra-ui/react";
import StatefulService from "@/components/StatefulService";
import CytoscapeGraph, { utils } from "@/components/CytoscapeGraph";
import { ChangeEvent, useEffect, useState } from "react";
import { ClusterTopology, Node } from "@/types";
import { useApi } from "@/contexts/ApiContext";

// TODO: This should be imported from the OpenAPI schema once the types there
// are generated correctly
interface TemplateConfig {
  service: string[];
  /** @description The name to give the template */
  name: string;
  /** @description The description of the template */
  description?: string;
}

const Page = () => {
  const { getTopology } = useApi();

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
    service: ["frontend"],
    description: "",
  });

  // fake preview of a flow built on the base topology
  const previewTopology = utils.normalizeData({
    ...topology,
    nodes: topology.nodes.map((node) => {
      console.log(services, "includes", node.id, services.includes(node.id));
      return {
        ...node,
        versions: formState.service.includes(node.id)
          ? [...(node.versions || []), "preview"]
          : node.versions,
      };
    }),
  });

  // TODO: Not hardcoding this
  const templateContainsExternalApi =
    formState.service.includes("jsdelivr-api");
  const templateContainsPostgres = formState.service.includes("postgres");

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

  return (
    <Layout>
      <PageTitle title="Create new flow configuration">
        Update traffic control and data isolation details below
      </PageTitle>
      <Section title="Preview">
        {!loading && <CytoscapeGraph elements={previewTopology} />}
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
          <Flex gap={4}>
            <Input.Select
              label="Services"
              options={services.map((s) => ({ label: s, value: s })) || []}
              value={formState.service[0] ?? ""}
              id="service"
              onChange={handleInputChange("service")}
            />
          </Flex>
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
      <Footer />
    </Layout>
  );
};

export default Page;
