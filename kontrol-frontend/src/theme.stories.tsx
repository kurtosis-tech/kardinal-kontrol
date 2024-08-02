import type { Meta, StoryObj } from "@storybook/react";

import theme, { colorOverrides } from "./theme";
import { VStack, HStack, Text, Flex } from "@chakra-ui/react";

const meta: Meta<typeof theme> = {
  component: () => {
    return (
      <div>
        <VStack>
          {Object.entries(theme.colors).map(
            ([key, value]: [string, unknown]) => (
              <HStack>
                <Text width={"100px"}>{key}</Text>
                {Object.entries(value as Record<string, string>).map(
                  ([shade, color]: [string, string]) => (
                    <Flex
                      border={
                        colorOverrides[key]?.[shade] != null
                          ? "5px dashed pink"
                          : "none"
                      }
                      alignItems="center"
                      justifyContent="center"
                      flexDir="column"
                      h="80px"
                      w="80px"
                      bg={color as string}
                      color={parseInt(shade) > 300 ? "white" : "black"}
                    >
                      <span>{shade}:</span>
                      <br />
                      <span>{color as string}</span>
                    </Flex>
                  ),
                )}
              </HStack>
            ),
          )}
        </VStack>
      </div>
    );
  },
};

export default meta;
type Story = StoryObj<typeof theme>;

export const Colors: Story = {
  args: {},
};
