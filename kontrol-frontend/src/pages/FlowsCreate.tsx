import Layout from "@/components/Layout";
import Section from "@/components/Section";
import PageTitle from "@/components/PageTitle";
import Footer from "@/components/Footer";
import Input from "@/components/Input";
import { Stack, Flex } from "@chakra-ui/react";
import { devFlow } from "@/components/TrafficConfiguration/mocks";
import CytoscapeGraph from "@/components/TrafficConfiguration/CytoscapeGraph";
import StatefulService from "@/components/StatefulService";

const Page = () => {
  return (
    <Layout>
      <PageTitle title="Create new flow configuration">
        Update traffic control and data isolation details below
      </PageTitle>
      <Section title="Preview">
        <CytoscapeGraph elements={devFlow} />
      </Section>
      <Section title="Flow configuration">
        <Stack w={"100%"} gap={8} as={"fieldset"}>
          <Input.Text
            label="Name"
            id="name"
            value="Early development flow"
            onChange={() => {}}
          />
          <Flex gap={4}>
            <Input.Select
              label="Name"
              options={[{ label: "Option 1", value: "option1" }]}
              value={"option1"}
              id="name"
              onChange={() => {}}
            />
            <Input.Text
              label="Image"
              id="image"
              value="registry.kardinal.dev/pr-1234"
              onChange={() => {}}
            />
          </Flex>
        </Stack>
      </Section>
      <Section title="Data configuration">
        <StatefulService type="stripe" />
      </Section>
      <Section offsetTop={9}>
        <StatefulService type="rds" />
      </Section>
      <Footer />
    </Layout>
  );
};

export default Page;
