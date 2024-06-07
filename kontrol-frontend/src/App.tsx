import {
  Box,
  Flex,
  Stack,
  Button,
  useSteps,
  Stepper,
  Step,
  StepIndicator,
  StepStatus,
  StepIcon,
  StepNumber,
  StepTitle,
  StepDescription,
  StepSeparator,
} from "@chakra-ui/react";
import { useEffect } from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";

const steps = [
  { title: "Create", description: "New deployment", path: "/" },
  { title: "Review", description: "Created deployment", path: "/review" },
  { title: "Manage", description: "Deployment", path: "/manage" },
];

function App() {
  const location = useLocation();
  const pathname = location.pathname;
  const navigate = useNavigate();

  const { activeStep, setActiveStep } = useSteps({
    index: 1,
    count: steps.length,
  });

  useEffect(() => {
    steps.forEach((step, index) => {
      if (step.path === pathname) {
        setActiveStep(index);
      }
    });
  }, [pathname, setActiveStep]);

  return (
    <Flex height={"100%"} alignItems={"center"} justifyContent={"center"} p={8}>
      <Stack
        direction={"column"}
        height={"100vh"}
        maxWidth={"1280px"}
        maxHeight={"920px"}
        width={"100%"}
      >
        <Stepper
          index={activeStep}
          p={8}
          borderRadius={12}
          borderWidth={1}
          borderColor={"gray.100"}
        >
          {steps.map((step, index) => (
            <Step key={index}>
              <StepIndicator>
                <StepStatus
                  complete={<StepIcon />}
                  incomplete={<StepNumber />}
                  active={<StepNumber />}
                />
              </StepIndicator>

              <Box flexShrink="0">
                <StepTitle>{step.title}</StepTitle>
                <StepDescription>{step.description}</StepDescription>
              </Box>

              <StepSeparator />
            </Step>
          ))}
        </Stepper>
        <Box
          borderRadius={12}
          borderWidth={1}
          borderColor={"gray.100"}
          p={8}
          flex={1}
        >
          <Outlet />
        </Box>

        <Stack
          p={8}
          spacing={4}
          direction="row"
          align="center"
          borderRadius={12}
          borderWidth={1}
          borderColor={"gray.100"}
          justifyContent={"space-between"}
        >
          <Button
            colorScheme="blue"
            variant="outline"
            size="lg"
            onClick={() => {
              const prevPath = steps[activeStep - 1]?.path;
              if (prevPath) {
                navigate(prevPath);
              }
            }}
          >
            Back
          </Button>
          <Button
            colorScheme="blue"
            size="lg"
            onClick={() => {
              const nextPath = steps[activeStep + 1]?.path;
              if (nextPath) {
                navigate(nextPath);
              }
            }}
          >
            Next
          </Button>
        </Stack>
      </Stack>
    </Flex>
  );
}

export default App;
