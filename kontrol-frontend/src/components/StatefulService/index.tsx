import { Text, Stack, Flex } from "@chakra-ui/react";
import { FiInfo } from "react-icons/fi";
import Chip from "@/components/Chip";
import { ElementType } from "react";
import { FaStripeS, FaAmazon, FaDatabase } from "react-icons/fa";
import Input, { SelectInputProps, TextInputProps } from "@/components/Input";
import { BiCodeCurly } from "react-icons/bi";

export type StatefulServiceType =
  | "stripe"
  | "rds"
  | "postgres"
  | "external_api";

type InputProps =
  | (TextInputProps & { type: "text" })
  | (SelectInputProps & { type: "select" });

export interface ServiceConfig {
  icon: ElementType;
  name: string;
  inputs: InputProps[];
  statefulLabel: string;
}

const postgresInputs: InputProps[] = [
  {
    type: "select",
    id: "access-mode",
    label: "Access mode",
    value: "empty-with-seed",
    options: [
      {
        label: "Shared",
        value: "shared",
      },
      {
        label: "Empty with seed script -- coming soon!",
        value: "empty-with-seed",
        disabled: true,
      },
      { label: "Snapshot -- coming soon!", value: "snapshot", disabled: true },
    ],
    onChange: () => {},
  },
  // {
  //   type: "text",
  //   id: "seed-script",
  //   label: "Seed script file",
  //   value: "https://github.com/kardinal/seed-scripts/blob/tree/master/seed.sql",
  //   onChange: () => {},
  // },
];

const apiKeyInputs: InputProps[] = [
  {
    type: "select",
    id: "access-mode",
    label: "Access mode",
    value: "development",
    options: [
      {
        label: "Development API keys",
        value: "development",
      },
      { label: "Production API keys", value: "production" },
    ],
    onChange: () => {},
  },
];

const serviceMeta: Record<StatefulServiceType, ServiceConfig> = {
  stripe: {
    icon: FaStripeS,
    name: "Stripe",
    statefulLabel: "Stateful service",
    inputs: [...apiKeyInputs],
  },
  external_api: {
    icon: BiCodeCurly,
    name: "External API",
    statefulLabel: "External service",
    inputs: [...apiKeyInputs],
  },
  rds: {
    icon: FaAmazon,
    name: "Amazon RDS",
    statefulLabel: "Stateful service",
    inputs: [...postgresInputs],
  },
  postgres: {
    icon: FaDatabase,
    name: "Postgres",
    statefulLabel: "Stateful service",
    inputs: [...postgresInputs],
  },
};

export interface Props {
  type: StatefulServiceType;
}

const StatefulService = ({ type }: Props) => {
  const inputs = serviceMeta[type].inputs;
  const meta = serviceMeta[type];
  return (
    <Stack gap={4} width={"100%"}>
      <Flex alignItems={"center"} gap={2}>
        <Text mb={2} m={0} fontWeight={400}>
          {meta.statefulLabel}
        </Text>
        <FiInfo />
      </Flex>
      <Flex>
        <Chip icon={serviceMeta[type].icon} colorScheme="blue">
          {meta.name}
        </Chip>
      </Flex>
      {inputs.map((input) =>
        input.type === "select" ? (
          <Input.Select {...input} />
        ) : (
          <Input.Text {...input} />
        ),
      )}
    </Stack>
  );
};

export default StatefulService;
