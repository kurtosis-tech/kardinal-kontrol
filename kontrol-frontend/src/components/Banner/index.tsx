import { useNavigationContext } from "@/contexts/NavigationContext";
import { Box, Flex, Text, CloseButton } from "@chakra-ui/react";

const Banner = () => {
  const { isBannerVisible, setIsBannerVisible } = useNavigationContext();

  if (!isBannerVisible) {
    return null;
  }

  return (
    <Box bg="orange.100" borderRadius={12} p={4} mb={"32px"}>
      <Flex justifyContent="space-between">
        <Flex gap={"10px"}>
          <Flex
            alignItems="center"
            justifyContent="center"
            borderRadius={8}
            w={"40px"}
            h={"40px"}
            background={"white"}
            flexShrink={0}
          >
            <Text style={{ fontSize: 25 }}>🚧</Text>
          </Flex>
          <Box>
            <Text fontWeight="bold" color="orange.500">
              WELCOME! KARDINAL IS STILL IN BUILD MODE
            </Text>
            <Text fontSize="sm" color={"gray.700"}>
              The features you see in the UI might differ from the CLI, we're
              currently working towards feature parity. <br /> If you have
              comments or suggestions, please{" "}
              <Text
                as="a"
                textDecoration="underline"
                cursor="pointer"
                href="https://github.com/kurtosis-tech/kardinal/issues"
                target="_blank"
              >
                open a github issue
              </Text>{" "}
              or{" "}
              <Text
                as="a"
                textDecoration="underline"
                cursor="pointer"
                href="mailto:hello@kardinal.dev"
              >
                reach out via email
              </Text>{" "}
              and help us shape the product!
            </Text>
          </Box>
        </Flex>
        <CloseButton
          size="sm"
          onClick={() => {
            setIsBannerVisible(false);
          }}
          color="orange.500"
        />
      </Flex>
    </Box>
  );
};

export default Banner;
