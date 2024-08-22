import EmptyState from "@/components/EmptyState";
import PageTitle from "@/components/PageTitle";
import { Spinner, Stack } from "@chakra-ui/react";
import { FiShield } from "react-icons/fi";
import TemplatesTable from "@/components/TemplatesTable";
import { useApi } from "@/contexts/ApiContext";
import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

const Page = () => {
  const [loading, setLoading] = useState<boolean>(true);
  const { templates, getTemplates } = useApi();
  const navigate = useNavigate();
  useEffect(() => {
    const fetchTemplates = async () => {
      await getTemplates();
    };
    fetchTemplates();
    setLoading(false);
  }, [getTemplates]);

  if (loading) {
    return <Spinner size="xl" />;
  }

  return (
    <Stack h="100%">
      <PageTitle
        title="Org-level Flow Configurations"
        buttonText="New Flow Configuration"
        onClick={() => {
          navigate("create");
        }}
      >
        See a list of all available flow configurations for your organization
      </PageTitle>
      {templates.length > 0 ? (
        <TemplatesTable templates={templates} />
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
