import Layout from "@/components/Layout";
import Section from "@/components/Section";
import PageTitle from "@/components/PageTitle";
import Footer from "@/components/Footer";
import Input from "@/components/Input";
import { Stack, Flex } from "@chakra-ui/react";
import StatefulService from "@/components/StatefulService";
import TrafficConfiguration from "@/components/TrafficConfiguration";

const Page = () => {
  return (
    <Layout>
      <PageTitle title="Create new flow configuration">
        Update traffic control and data isolation details below
      </PageTitle>
      <Section title="Preview">
        <TrafficConfiguration />
      </Section>
      <Section title="Flow configuration">
        <Stack w={"100%"} gap={8} as={"fieldset"}>
          <Input.Text
            label="Name"
            id="name"
            value="Online boutique demo flow"
            onChange={() => {}}
          />
          <Flex gap={4}>
            <Input.Select
              label="Service"
              options={[
                { label: "cartservice", value: "cartservice" },
                { label: "frontend", value: "frontend" },
              ]}
              value={"frontend"}
              id="name"
              onChange={() => {}}
            />
            <Input.Text
              label="Image"
              id="image"
              value="leoporoli/newobd-frontend:0.0.6"
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
