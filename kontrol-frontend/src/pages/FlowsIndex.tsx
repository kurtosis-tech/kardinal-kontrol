import EmptyState from "@/components/EmptyState";
import PageTitle from "@/components/PageTitle";
import { useFlowsContext } from "@/contexts/FlowsContext";
import { Stack } from "@chakra-ui/react";
import { FiShield } from "react-icons/fi";
import Table from "@/components/Table";

const Page = () => {
  const { flows } = useFlowsContext();
  return (
    <Stack h="100%">
      <PageTitle title="Org-level Flow Configurations">
        See a list of all available flow configurations for your organization
      </PageTitle>
      {flows.length > 0 ? (
        <Table />
      ) : (
        <EmptyState
          icon={FiShield}
          buttonText="Create new flow configuration"
          buttonTo="create"
        >
          No active flow configurations created yet
        </EmptyState>
      )}
    </Stack>
  );
};

export default Page;
