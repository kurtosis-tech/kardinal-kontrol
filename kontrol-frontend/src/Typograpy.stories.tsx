import { Stack, Text } from "@chakra-ui/react";

import type { Meta, StoryObj } from "@storybook/react";

const meta: Meta = {
  render() {
    return (
      <Stack spacing={4} color={"black"}>
        <Text fontSize="6xl">
          6xl/60px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="5xl">
          5xl/48px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="4xl">
          4xl/36px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="3xl">
          3xl/30px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="2xl">
          2xl/24px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="xl">
          xl/20px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="lg">
          lg/18px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="md">
          md/16px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="sm">
          sm/14px: The quick brown fox jumps over the lazy dog
        </Text>
        <Text fontSize="xs">
          xs/12px: The quick brown fox jumps over the lazy dog
        </Text>
      </Stack>
    );
  },
};

export default meta;
type Story = StoryObj<typeof Text>;

export const FontSizes: Story = {};
