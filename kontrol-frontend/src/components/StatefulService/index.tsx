import { Text, Stack, Flex } from "@chakra-ui/react";
import { FiInfo } from "react-icons/fi";
import Chip from "@/components/Chip";
import { ElementType } from "react";
import { FaStripeS, FaAmazon } from "react-icons/fa";
import Input, { SelectInputProps, TextInputProps } from "@/components/Input";

export type StatefulServiceType = "stripe" | "rds";
type InputProps =
  | (TextInputProps & { type: "text" })
  | (SelectInputProps & { type: "select" });

export interface ServiceConfig {
  icon: ElementType;
  name: string;
  inputs: InputProps[];
}

const serviceMeta: Record<StatefulServiceType, ServiceConfig> = {
  stripe: {
    icon: FaStripeS,
    name: "Stripe",
    inputs: [
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
    ],
  },
  rds: {
    icon: FaAmazon,
    name: "Amazon RDS",
    inputs: [
      {
        type: "select",
        id: "access-mode",
        label: "Access mode",
        value: "empty-with-seed",
        options: [
          {
            label: "Empty with seed script",
            value: "empty-with-seed",
          },
          { label: "Snapshot", value: "snapshot" },
        ],
        onChange: () => {},
      },
      {
        type: "text",
        id: "seed-script",
        label: "Seed script file",
        value:
          "https://github.com/kardinal/seed-scripts/blob/tree/master/seed.sql",
        onChange: () => {},
      },
    ],
  },
};

export interface Props {
  type: StatefulServiceType;
}

const StatefulService = ({ type }: Props) => {
  const inputs = serviceMeta[type].inputs;
  return (
    <Stack gap={4} width={"100%"}>
      <Flex alignItems={"center"} gap={2}>
        <Text mb={2} m={0} fontWeight={400}>
          Stateful service
        </Text>
        <FiInfo />
      </Flex>
      <Flex>
        <Chip icon={serviceMeta[type].icon} colorScheme="blue">
          {serviceMeta[type].name}
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
